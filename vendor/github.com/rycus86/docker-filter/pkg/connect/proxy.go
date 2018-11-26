package connect

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"
)

var (
	proxyIndex = 0
	connIndex  = 0
)

func NewProxy(remote func() (net.Conn, error), nonManagedResponses ...string) *Proxy {
	proxyIndex += 1

	return &Proxy{
		listeners: []*localListener{},
		dialer:    remote,
		handlers:  []*handler{},

		idx: proxyIndex,

		nonManagedResponses: nonManagedResponses,
	}
}

func (p *Proxy) AddListener(prefix string, listener net.Listener) {
	if prefix == "" {
		prefix = listener.Addr().Network()
	}

	p.listeners = append(p.listeners, &localListener{
		Listener:  listener,
		logPrefix: prefix,
	})
}

func (p *Proxy) Handle(urlPattern string, filterFunc FilterFunc) {
	p.FilterRequests(urlPattern, RequestFilterFunc(filterFunc))
}

func (p *Proxy) FilterRequests(urlPattern string, filterFunc RequestFilterFunc) {
	p.handlers = append(p.handlers, &handler{
		pattern:       regexp.MustCompile(urlPattern),
		requestFilter: filterFunc,
	})
}

func (p *Proxy) FilterResponses(urlPattern string, filterFunc ResponseFilterFunc) {
	p.handlers = append(p.handlers, &handler{
		pattern:        regexp.MustCompile(urlPattern),
		responseFilter: filterFunc,
	})
}

func (p *Proxy) Process() error {
	if len(p.listeners) == 0 {
		return errors.New("no local listeners are registered")
	}

	acceptChan := p.startPolling()

	for {
		polled := <-acceptChan

		if polled.err != nil {
			p.closeAll()
			return polled.err
		}

		if pair, err := polled.conn.connectToRemote(p); err != nil {
			polled.conn.writeFailedResponse("failed to connect to the remote", err)
			continue

		} else {
			go pair.handleRequests()
			go pair.handleResponses()

		}
	}
}

func (p *Proxy) startPolling() chan *pollResult {
	acceptChan := make(chan *pollResult, 1)
	listenLoop := func(listener *localListener, num int) {
		for {
			conn, err := listener.Accept()
			if err != nil {
				break
			}

			acceptChan <- &pollResult{
				conn: &localConnection{
					Conn:      conn,
					proxy:     p,
					idx:       num,
					logPrefix: listener.logPrefix,
				},
				err: nil,
			}
		}
	}

	for idx, local := range p.listeners {
		go listenLoop(local, idx+1)
	}

	return acceptChan
}

func (p *Proxy) closeAll() {
	for _, listener := range p.listeners {
		listener.Close()
	}
}

func (lc *localConnection) nextLogPrefix() string {
	connIndex += 1

	return fmt.Sprintf("(%02d|%s|%02d|%04d)",
		lc.proxy.idx, lc.logPrefix, lc.idx, connIndex)
}

func (lc *localConnection) connectToRemote(p *Proxy) (*connectionPair, error) {
	remoteConn, err := p.dialer()
	if err != nil {
		lc.writeFailedResponse("Failed to connect the proxy to the remote", err)
		lc.Close()
		return nil, err
	}

	return &connectionPair{
		localConn:  lc,
		remoteConn: remoteConn,
		proxy:      p,

		logPrefix: lc.nextLogPrefix(),
	}, nil
}

func (lc *localConnection) writeFailedResponse(reason string, err error) error {
	message := fmt.Sprintf("{\"message\":\"%s: %s\"}", reason, err.Error())
	response := &http.Response{
		StatusCode:    503,
		ProtoMajor:    1,
		ProtoMinor:    1,
		Close:         true,
		ContentLength: int64(len(message)),
		Header:        http.Header{"Content-Type": {"application/json"}},
		Body:          ioutil.NopCloser(strings.NewReader(message)),
	}

	return response.Write(lc)
}

func (cp *connectionPair) handleRequests() {
	buffer := make([]byte, 16000)

	cp.upgraded = false

	for {
		n, err := cp.localConn.Read(buffer)
		if err != nil {
			if cp.upgraded && err == io.EOF {
				cp.closeReading()
				return
			}

			cp.close("request", err)
			return
		}

		if n > 0 {
			data := buffer[0:n]
			request, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(data)))

			if err != nil {
				cp.remoteConn.Write(data)
				cp.debug("Sent request data:", n, "bytes (failed to parse request)")
			} else if body, err := ioutil.ReadAll(request.Body); err != nil {
				cp.remoteConn.Write(data)
				cp.debug("Sent request data:", n, "bytes (failed to read request body)")
			} else {
				if request.Header.Get("Upgrade") == "tcp" {
					cp.upgraded = true
				}

				for _, handler := range cp.proxy.handlers {
					if handler.requestFilter == nil {
						continue
					}

					if !handler.pattern.MatchString(request.URL.Path) {
						continue
					}

					if changedRequest, err := runRequestHandler(handler, request, body); err != nil {
						if _, ok := err.(CriticalFailure); ok {
							cp.error("Critical:", "Failed to execute request filter on", request.URL, ":", err)

							cp.localConn.writeFailedResponse("Failed to apply filter", err)
							cp.close("request", err)
							return

						} else {
							cp.warn("Request filter warning on", request.URL, ":", err)
						}

					} else if changedRequest != nil {
						request = changedRequest
						body, _ = ioutil.ReadAll(changedRequest.Body)
						changedRequest.Body.Close()

					}
				}

				if len(body) > 0 {
					request.Body = ioutil.NopCloser(bytes.NewReader(body))
				}

				cp.latestRequest = request

				request.Write(cp.remoteConn)
				cp.info("Sent HTTP request to", request.URL, ":", len(body), "bytes")
			}
		}
	}
}

func runRequestHandler(handler *handler, request *http.Request, body []byte) (changed *http.Request, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case SoftFailure, CriticalFailure:
				err = r.(error)
			default:
				err = NewCriticalFailure(r, "RequestFilter")
			}
		}
	}()

	return handler.requestFilter(request, body)
}

func (cp *connectionPair) handleResponses() {
	buffer := make([]byte, 16000)
	reader := &responseReader{r: cp.remoteConn}

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			cp.close("response", err)
			return
		}

		if n > 0 {
			data := buffer[0:n]

			if response, err := reader.toResponse(cp.latestRequest); response != nil && err == nil {
				requestUrl := "/<unknown>"
				if response.Request != nil {
					requestUrl = response.Request.URL.Path
				}

				var body []byte

				if cp.allowReadingResponseBody(response) {
					body, err = ioutil.ReadAll(response.Body)
					if err != nil {
						cp.close("response", err)
						return
					}
					response.Body.Close()

					response.Request = cp.latestRequest
					response.Body = ioutil.NopCloser(bytes.NewReader(body))
				}

				for _, handler := range cp.proxy.handlers {
					if handler.responseFilter == nil {
						continue
					}

					if !handler.pattern.MatchString(requestUrl) {
						continue
					}

					if changedResponse, err := runResponseHandler(handler, response, body); err != nil {
						if _, ok := err.(CriticalFailure); ok {
							cp.error("Critical:", "Failed to execute response filter on", requestUrl, ":", err)

							cp.localConn.writeFailedResponse("Failed to apply filter", err)
							cp.close("response", err)
							return

						} else {
							cp.warn("Response filter warning on", requestUrl, ":", err)
						}

					} else if changedResponse != nil {
						response = changedResponse
						body, _ = ioutil.ReadAll(changedResponse.Body)
						changedResponse.Body.Close()

					}
				}

				if len(body) > 0 {
					response.Body = ioutil.NopCloser(bytes.NewReader(body))
				}

				response.Write(cp.localConn)

				cp.info("Response: HTTP", response.StatusCode)
				cp.debug("Sent response data:", len(body), "bytes")

			} else {
				cp.localConn.Write(data)
				cp.debug("Sent response data:", n, "bytes")

			}
		}
	}
}

func (cp *connectionPair) allowReadingResponseBody(response *http.Response) bool {
	if cp.upgraded {
		return false // the connection is upgraded (to raw stream)
	}

	if response == nil {
		return false // we couldn't parse the response
	}

	if response.Request == nil {
		return false // we don't have the original request
	}

	url := response.Request.URL.Path
	for _, path := range cp.proxy.nonManagedResponses {
		if strings.Contains(url, path) {
			return false
		}
	}

	return true
}

func runResponseHandler(handler *handler, response *http.Response, body []byte) (changed *http.Response, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(string(debug.Stack()))

			switch r.(type) {
			case SoftFailure, CriticalFailure:
				err = r.(error)
			default:
				err = NewCriticalFailure(r, "ResponseFilter")
			}
		}
	}()

	return handler.responseFilter(response, body)
}

func (cp *connectionPair) close(from string, err error) {
	cp.debug("Closing the connections:", err, "(from "+from+")")
	cp.localConn.Close()
	cp.remoteConn.Close()
}

type responseReader struct {
	r io.Reader

	cached []byte

	startOver bool
}

func (cr *responseReader) StartOver() {
	cr.startOver = true
}

func (cr *responseReader) Read(b []byte) (int, error) {
	if cr.startOver {
		ll := len(b)
		if len(cr.cached) <= ll {
			ll = len(cr.cached)
			cr.startOver = false
		}

		copy(b[0:ll], cr.cached[0:ll])

		cr.cached = cr.cached[ll:]

		return ll, nil
	}

	n, err := cr.r.Read(b)
	if err == nil {
		cr.cached = b[0:n]
	}

	return n, err
}

func (cr *responseReader) toResponse(req *http.Request) (*http.Response, error) {
	if len(cr.cached) > 0 {
		_, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(cr.cached)), req)
		if err != nil {
			return nil, err
		}

		cr.StartOver()
		return http.ReadResponse(bufio.NewReaderSize(cr, len(cr.cached)), req)
	} else {
		return nil, fmt.Errorf("no request available")
	}
}

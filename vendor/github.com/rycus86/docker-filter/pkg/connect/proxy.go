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
	"strings"
	"time"
)

var (
	proxyIndex = 0
	connIndex  = 0
)

func NewProxy(remote func() (net.Conn, error)) *Proxy {
	proxyIndex += 1

	return &Proxy{
		listeners: []*localListener{},
		dialer:    remote,
		handlers:  []*handler{},

		idx: proxyIndex,
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
	p.handlers = append(p.handlers, &handler{
		pattern: regexp.MustCompile(urlPattern),
		filter:  filterFunc,
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

		closeAfterResponse: false,
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

	upgraded := false

	for {
		n, err := cp.localConn.Read(buffer)
		if err != nil {
			if upgraded && err == io.EOF {
				cp.closeAfterResponse = true
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
					upgraded = true
				}

				for _, handler := range cp.proxy.handlers {
					if handler.pattern.MatchString(request.URL.Path) {
						if changedRequest, err := runHandler(handler, request, body); err != nil {
							if _, ok := err.(CriticalFailure); ok {
								cp.error("Critical:", "Failed to execute filter on", request.URL, ":", err)

								cp.localConn.writeFailedResponse("Failed to apply filter", err)
								cp.close("request", err)
								return

							} else {
								cp.warn("Filter warning on", request.URL, ":", err)
							}

						} else if changedRequest != nil {
							request = changedRequest
							body, _ = ioutil.ReadAll(changedRequest.Body)

						}
					}
				}

				if len(body) > 0 {
					request.Body = ioutil.NopCloser(bytes.NewReader(body))
				}

				request.Write(cp.remoteConn)
				cp.info("Sent HTTP request to", request.URL, ":", len(body), "bytes")
			}
		}
	}
}

func runHandler(handler *handler, request *http.Request, body []byte) (changed *http.Request, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case SoftFailure, CriticalFailure:
				err = r.(error)
			default:
				err = NewCriticalFailure(r, "Filter")
			}
		}
	}()

	return handler.filter(request, body)
}

func (cp *connectionPair) handleResponses() {
	buffer := make([]byte, 8000)

	for {
		n, err := cp.remoteConn.Read(buffer)
		if err != nil {
			cp.close("response", err)
			return
		}

		if n > 0 {
			data := buffer[0:n]
			cp.localConn.Write(data)
			cp.debug("Sent response data:", n, "bytes")

			if cp.closeAfterResponse {
				// TODO this feels really hacky
				cp.remoteConn.SetReadDeadline(time.Now().Add(1 * time.Millisecond))
			}
		}
	}
}

func (cp *connectionPair) close(from string, err error) {
	cp.debug("Closing the connections:", err, "(from "+from+")")
	cp.localConn.Close()
	cp.remoteConn.Close()
}

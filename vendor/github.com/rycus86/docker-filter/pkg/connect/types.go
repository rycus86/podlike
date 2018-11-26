package connect

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
)

type Proxy struct {
	listeners []*localListener
	dialer    func() (net.Conn, error)
	handlers  []*handler

	idx int

	nonManagedResponses []string
}

type RequestFilterFunc func(req *http.Request, body []byte) (*http.Request, error)
type ResponseFilterFunc func(resp *http.Response, body []byte) (*http.Response, error)

type FilterFunc RequestFilterFunc

type handler struct {
	pattern *regexp.Regexp

	requestFilter  RequestFilterFunc
	responseFilter ResponseFilterFunc
}

type localListener struct {
	net.Listener
	logPrefix string
}

type localConnection struct {
	net.Conn

	proxy *Proxy

	idx       int
	logPrefix string
}

type connectionPair struct {
	localConn  *localConnection
	remoteConn net.Conn
	proxy      *Proxy

	logPrefix string

	upgraded      bool
	latestRequest *http.Request
}

type pollResult struct {
	conn *localConnection
	err  error
}

type filterFailure struct {
	Cause    error
	Category string
}

type CriticalFailure filterFailure

func (cf CriticalFailure) Error() string {
	if cf.Category != "" {
		return fmt.Sprintf("%s: %s", cf.Category, cf.Cause.Error())
	} else {
		return cf.Cause.Error()
	}
}

func NewCriticalFailure(cause interface{}, category string) CriticalFailure {
	if cause == nil {
		panic("no failure cause given")
	}

	if asError, ok := cause.(error); ok {
		return CriticalFailure{asError, category}
	} else {
		return CriticalFailure{
			Cause:    errors.New(fmt.Sprintf("%s", cause)),
			Category: category,
		}
	}
}

type SoftFailure filterFailure

func (sf SoftFailure) Error() string {
	if sf.Category != "" {
		return fmt.Sprintf("%s: %s", sf.Category, sf.Cause.Error())
	} else {
		return sf.Cause.Error()
	}
}

func NewSoftFailure(cause interface{}, category string) SoftFailure {
	if cause == nil {
		panic("no failure cause given")
	}

	if asError, ok := cause.(error); ok {
		return SoftFailure{asError, category}
	} else {
		return SoftFailure{
			Cause:    errors.New(fmt.Sprintf("%s", cause)),
			Category: category,
		}
	}
}

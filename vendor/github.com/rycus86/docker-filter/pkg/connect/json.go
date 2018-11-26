package connect

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type T interface{}

func FilterRequestAsJson(provideValue func() T, change func(T) T) RequestFilterFunc {
	return RequestFilterFunc(FilterAsJson(provideValue, change))
}

func FilterAsJson(provideValue func() T, change func(T) T) FilterFunc {
	return func(req *http.Request, body []byte) (*http.Request, error) {
		v := provideValue()

		if err := json.NewDecoder(bytes.NewReader(body)).Decode(v); err == nil {
			if changed := change(v); changed != nil {
				v = changed
			} else {
				return nil, nil
			}

			if body, err = json.Marshal(v); err != nil {
				return nil, NewCriticalFailure(err, "JSON")
			}

			res, err := http.NewRequest(req.Method, req.URL.String(), bytes.NewReader(body))
			if err != nil {
				return nil, NewCriticalFailure(err, "JSON")
			}

			for headerName, headerValues := range req.Header {
				for _, value := range headerValues {
					res.Header.Add(headerName, value)
				}
			}

			return res, nil
		} else {
			return nil, NewCriticalFailure(err, "JSON")
		}
	}
}

func FilterResponseAsJson(provideValue func() T, change func(T) T) ResponseFilterFunc {
	return func(resp *http.Response, body []byte) (*http.Response, error) {
		v := provideValue()

		if err := json.NewDecoder(bytes.NewReader(body)).Decode(v); err == nil {
			if changed := change(v); changed != nil {
				v = changed
			} else {
				return nil, nil
			}

			if body, err = json.Marshal(v); err != nil {
				return nil, NewCriticalFailure(err, "JSON")
			}

			res := new(http.Response)
			*res = *resp

			res.Body = ioutil.NopCloser(bytes.NewReader(body))
			res.ContentLength = int64(len(body))

			return res, nil
		} else {
			return nil, NewCriticalFailure(err, "JSON")
		}
	}
}

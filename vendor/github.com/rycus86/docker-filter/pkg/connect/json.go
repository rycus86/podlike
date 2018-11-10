package connect

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type T interface{}

func FilterAsJson(provideValue func() T, change func(T) T) FilterFunc {
	return func(req *http.Request, body []byte) (*http.Request, error) {
		v := provideValue()

		if err := json.NewDecoder(bytes.NewReader(body)).Decode(v); err == nil {
			if changed := change(v); changed != nil {
				v = changed
			}

			body, _ = json.Marshal(v)
			res, _ := http.NewRequest(req.Method, req.URL.String(), bytes.NewReader(body))

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

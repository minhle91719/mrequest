package mrequest

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
)

func ReadBody(res *http.Response) (data []byte, err error) {
	reader := res.Body
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(res.Body)
		if err != nil {
			return nil, err
		}
	}
	defer func() {
		res.Body.Close()
		reader.Close()
	}()

	return ioutil.ReadAll(reader)
}

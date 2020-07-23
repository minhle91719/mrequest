package mrequest

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strings"
)

type requestBuilder struct {
	_url     string
	method   string
	ct       contentType
	ctLength int
	data     io.Reader

	mapHeader map[string]string
	cookies   []*http.Cookie
}

func (r *requestBuilder) AddCookie(cookies []*http.Cookie) RequestBuilder {
	r.cookies = append(r.cookies, cookies...)
	return r
}

func (r *requestBuilder) URL(_url string) RequestBuilder {
	r._url = _url
	return r
}

func (r *requestBuilder) Method(method string) RequestBuilder {
	r.method = method
	return r
}

func (r *requestBuilder) Body(ct contentType, data io.Reader, contentLength int) RequestBuilder {
	r.ct = ct
	r.data = data
	return r
}

func (r *requestBuilder) RandomUserAgent(deviceType userAgentType) RequestBuilder {
	r.SetUserAgent(randomUA())
	return r
}

func (r *requestBuilder) SetUserAgent(ua string) RequestBuilder {
	r.mapHeader[strings.ToLower("user-agent")] = ua
	return r
}

func (r *requestBuilder) AddHeader(m map[string]string) RequestBuilder {
	for k, v := range m {
		r.mapHeader[strings.ToLower(k)] = v
	}
	return r
}

func (r *requestBuilder) Build() (request *http.Request, err error) {
	if r.method == "" {
		return nil, errors.New("method missing")
	}
	if r._url == "" {
		return nil, errors.New("method missing")
	}
	request, err = http.NewRequest(r.method, r._url, r.data)
	if err != nil {
		return request, errors.WithStack(err)
	}
	for k, v := range r.mapHeader {
		request.Header.Set(k, v)
	}
	for _, v := range r.cookies {
		request.AddCookie(v)
	}
	if r.ctLength > 0 {
		request.Header.Set("content-length", fmt.Sprintf("%d", r.ctLength))
	}
	return request, nil
}

type RequestBuilder interface {
	URL(_url string) RequestBuilder
	Method(method string) RequestBuilder
	Body(ct contentType, data io.Reader, contentLength int) RequestBuilder

	RandomUserAgent(deviceType userAgentType) RequestBuilder
	SetUserAgent(ua string) RequestBuilder
	AddHeader(map[string]string) RequestBuilder
	AddCookie(cookies []*http.Cookie) RequestBuilder
	Build() (r *http.Request, err error)
}

func NewRequestBuilder() RequestBuilder {
	return &requestBuilder{
		_url:      "",
		method:    "",
		ct:        "",
		data:      nil,
		mapHeader: map[string]string{},
	}
}

package mrequest

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	DefaultMaxRequestPerSecond = 1
)

type RQ struct {
	_host      string
	_mapCookie map[string]http.Cookie
	_ref       string
	ua         string
	client     *http.Client

	requestAccess chan int

	ctx    context.Context
	cancel context.CancelFunc
}

type contentType string

const (
	WWWType    contentType = "application/x-www-form-urlencoded"
	JSONType   contentType = "application/json"
	TextJSType contentType = "text/javascript"
)

func NewRequest(host string, client *http.Client, rps int) IRequest {
	if rps == 0 {
		rps = DefaultMaxRequestPerSecond
	}
	if client == nil {
		client = http.DefaultClient
		client.Timeout = 5 * time.Second
	}
	rq := &RQ{
		_host:         host,
		_mapCookie:    make(map[string]http.Cookie),
		_ref:          host,
		ua:            randomUA(),
		requestAccess: make(chan int, rps),
		client:        client,
	}
	rq.ctx, rq.cancel = context.WithCancel(context.Background())
	for i:= 0 ; i < rps;i++ {
		rq.requestAccess <- 1
	}
	go rq.grantRequest(rps)
	return rq
}

type IRequest interface {
	Request(f func() (*http.Request, error)) ([]byte, error)
	Close()
}

func (r *RQ) grantRequest(maxRequestPerSecond int) {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		select {
		case <-r.ctx.Done():
			return
		default:
			for len(r.requestAccess) < maxRequestPerSecond {
				r.requestAccess <- 1
			}
		}
	}
}
func (r *RQ) Close() {
	r.cancel()
}
func (r *RQ) Request(f func() (*http.Request, error)) ([]byte, error) {
	<-r.requestAccess
	var (
		req *http.Request
		res *http.Response

		reader io.ReadCloser

		err error
	)
	req, err = f()
	if err != nil {
		return nil, err
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Add("User-Agent", r.ua)
	}
	switch req.Method {
	case http.MethodPost:
		contentIO, err := req.GetBody()
		// req.GetBody return copy of body, dont use ioutil.NopCloser
		if err != nil {
			return nil, err
		}
		content, err := ioutil.ReadAll(contentIO)
		req.Header.Add("Content-Length", fmt.Sprintf("%d", len(string(content))))
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Referer", r._ref)
	for _, v := range r._mapCookie {
		ck := http.Cookie{}
		ck = v
		req.AddCookie(&ck)
	}
	if res, err = r.client.Do(req); err != nil {
		return nil, err
	}
	listCookie := res.Cookies()
	for _, v := range listCookie {
		r._mapCookie[v.Name] = *v
	}
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(res.Body)
	default:
		reader = res.Body
	}
	defer func() {
		res.Body.Close()
		reader.Close()
	}()
	r._ref = req.URL.String()
	return ioutil.ReadAll(reader)
}

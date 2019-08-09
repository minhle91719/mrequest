package mrequest

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	
	"golang.org/x/time/rate"
	"time"
)

var (
	defaultLimiter = rate.NewLimiter(rate.Every(5*time.Second), 1)
)
// TODO: add request file return io.ReadSeeker
type RQ struct {
	_host      string
	_mapCookie map[string]http.Cookie
	_ref       string
	ua         string
	client     *http.Client
	
	limiter *rate.Limiter
	
	ctx context.Context
}

type contentType string

const (
	WWWType    contentType = "application/x-www-form-urlencoded"
	JSONType   contentType = "application/json"
	TextJSType contentType = "text/javascript"
)

func NewRequest(ctx context.Context, host string, client *http.Client, limiter *rate.Limiter) IRequest {
	if limiter == nil {
		limiter = defaultLimiter
	}
	if client == nil {
		client = http.DefaultClient
		client.Timeout = 5 * time.Second
	}
	rq := &RQ{
		_host:      host,
		_mapCookie: make(map[string]http.Cookie),
		_ref:       host,
		ua:         randomUA(),
		limiter:    limiter,
		client:     client,
		ctx:        ctx,
	}
	return rq
}

type IRequest interface {
	Request(f func() (*http.Request, error)) ([]byte, error)
}

func (r *RQ) Request(f func() (*http.Request, error)) ([]byte, error) {
	for !r.limiter.Allow() {
	}
	if _, ok := <-r.ctx.Done(); ok {
		return nil, errors.New("context canceled")
	}
	
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
		//fmt.Println(v)
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

package mrequest

import (
	"golang.org/x/time/rate"
	"net/http"
	"time"
)

type clientLimit struct {
	cli     *http.Client
	limiter *rate.Limiter
}

func (c *clientLimit) GetClient() *http.Client {
	for !c.limiter.Allow() {
	}
	return c.cli
}

type Client interface {
	GetClient() *http.Client
}

func NewClient(cli *http.Client, limitRequestPerSec int) Client {
	return &clientLimit{
		cli:     cli,
		limiter: rate.NewLimiter(rate.Every(1*time.Second), limitRequestPerSec),
	}
}

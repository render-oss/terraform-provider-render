package provider

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)

type RateLimitHTTPClient struct {
	client      *http.Client
	rateLimiter *rate.Limiter
	sleepFunc   func(time.Duration)
}

func NewRateLimitHTTPClient(client *http.Client, rateLimiter *rate.Limiter, sleepFunc func(time.Duration)) *RateLimitHTTPClient {
	return &RateLimitHTTPClient{
		client:      client,
		rateLimiter: rateLimiter,
		sleepFunc:   sleepFunc,
	}
}

var backoff = []int{1, 5, 10, 20, 40, 60, 120}

func (c *RateLimitHTTPClient) Do(req *http.Request) (*http.Response, error) {
	err := c.rateLimiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusTooManyRequests {
		if retryAt := res.Header.Get("Retry-After"); retryAt != "" {
			seconds, err := strconv.Atoi(retryAt)
			if err == nil {
				c.sleepFunc(time.Duration(seconds) * time.Second)
				res, err = c.client.Do(req)
				if err != nil {
					return nil, err
				}
				if res.StatusCode != http.StatusTooManyRequests {
					return res, nil
				}
			}
		}

		for _, seconds := range backoff {
			c.sleepFunc(time.Duration(seconds) * time.Second)
			res, err = c.client.Do(req)
			if err != nil {
				return nil, err
			}
			if res.StatusCode != http.StatusTooManyRequests {
				return res, nil
			}
		}

		return nil, fmt.Errorf("requests consistently rate limited")
	}

	return res, nil
}

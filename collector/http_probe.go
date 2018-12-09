package collector

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// HTTPProbe is an empty struct
type HTTPProbe struct {
	urls            []*url.URL
	timeoutDuration time.Duration
	c               *http.Client
	timeoutError    error
}

// NewHTTPProbe create a new HTTPProbe.
func NewHTTPProbe(urls []*url.URL, c *http.Client) HTTPProbe {

	uris := make([]string, 0)
	for _, u := range urls {
		uris = append(uris, u.String())
	}

	return HTTPProbe{
		urls:            urls,
		c:               c,
		timeoutDuration: c.Timeout,
		timeoutError:    fmt.Errorf("all uris unreachable, %s", strings.Join(uris, ",")),
	}
}

// ProbeURL sends http head request to all urls and return the first url that responsed successfully.
// Return err when non of the url is successful.
func (p *HTTPProbe) ProbeURL() (*url.URL, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chans := make([]<-chan string, len(p.urls))

	for _, u := range p.urls {
		chans = append(chans, httpProbe(ctx, cancel, u.String()))
	}

	if p.timeoutDuration > 0 {
		timeout := make(chan string, 1)
		go func() {
			<-time.After(p.timeoutDuration)
			cancel()
			timeout <- "timedout"
		}()
		chans = append(chans, timeout)
	}

	cases := make([]reflect.SelectCase, len(chans))

	for i, ch := range chans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}

	_, value, _ := reflect.Select(cases)

	if value.String() == "timedout" {
		return nil, p.timeoutError
	}

	v, err := url.Parse(value.String())

	if err != nil {
		return nil, err
	}

	return v, nil
}

func httpProbe(ctx context.Context, cancel context.CancelFunc, uri string) <-chan string {

	c1 := make(chan string, 1)

	go func() {

		req, err := http.NewRequest("HEAD", uri, nil)

		if err != nil {
			return
		}

		req = req.WithContext(ctx)

		resp, err := http.DefaultClient.Do(req)

		if err == nil && resp.StatusCode >= 200 && resp.StatusCode <= 399 {
			cancel()
			c1 <- uri
		}
	}()
	return c1
}

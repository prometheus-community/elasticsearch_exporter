package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type Nodes struct {
	logger log.Logger
	client *http.Client
	url    url.URL
	all    bool

	foobar *prometheus.Desc
}

func NewNodes(logger log.Logger, client *http.Client, url url.URL, all bool) *Nodes {
	return &Nodes{
		logger: logger,
		client: client,
		url:    url,
		all:    all,

		foobar: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "foo", "bar"),
			"bla bla bla.",
			nil, nil,
		),
	}
}

func (c *Nodes) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.foobar
}

func (c *Nodes) Collect(ch chan<- prometheus.Metric) {
	path := "/_nodes/_local/stats"
	if c.all {
		path = "/_nodes/stats"
	}
	c.url.Path = path

	res, err := c.client.Get(c.url.String())
	if err != nil {
		level.Warn(c.logger).Log(
			"msg", "failed to get nodes",
			"url", c.url.String(),
			"err", err,
		)
		return
	}
	defer res.Body.Close()

	dec := json.NewDecoder(res.Body)

	var nodeStatsResponse NodeStatsResponse
	if err := dec.Decode(&nodeStatsResponse); err != nil {
		level.Warn(c.logger).Log(
			"msg", "failed to decode nodes",
			"err", err,
		)
		return
	}

	for _, node := range nodeStatsResponse.Nodes {
		fmt.Printf("host: %+v\n", node.Host)
	}
}

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

type MetricsCollector struct {
	clusterHealth *ClusterHealth
	nodes         *Nodes
	indices       *Indices

	clusterHealthResponse *clusterHealthResponse
	nodeStatsResponse     *nodeStatsResponse
	indexStatsResponse    *indexStatsResponse
}

func NewMetricsCollector(logger log.Logger, client *http.Client, url *url.URL, all bool) *MetricsCollector {
	return &MetricsCollector{
		clusterHealth: NewClusterHealth(logger, client, url),
		nodes:         NewNodes(logger, client, url, all),
		indices:       NewIndices(logger, client, url, all),
	}
}

func (c *MetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.clusterHealth.Describe(ch)
	c.nodes.Describe(ch)
	c.indices.Describe(ch)
}

func (c *MetricsCollector) Collect(ch chan<- prometheus.Metric) {
	// clusterHealth
	clusterHealthResponse, err := c.clusterHealth.fetchAndDecodeClusterHealth()
	if err != nil {
		c.clusterHealth.up.Set(0)
		level.Warn(c.clusterHealth.logger).Log(
			"msg", "failed to fetch and decode cluster health",
			"err", err,
		)
		return
	}
	c.clusterHealth.up.Set(1)

	// nodes
	nodeStatsResponse, err := c.nodes.fetchAndDecodeNodeStats()
	if err != nil {
		c.nodes.up.Set(0)
		level.Warn(c.nodes.logger).Log(
			"msg", "failed to fetch and decode node stats",
			"err", err,
		)
		return
	}
	c.nodes.up.Set(1)

	// indices
	indexStatsResponse, err := c.indices.fetchAndDecodeIndexStats()
	if err != nil {
		c.indices.up.Set(0)
		level.Warn(c.indices.logger).Log(
			"msg", "failed to fetch and decode index stats",
			"err", err,
		)
		return
	}
	c.indices.up.Set(1)

	c.clusterHealth.Collect(ch, clusterHealthResponse)
	c.nodes.Collect(ch, nodeStatsResponse)
	c.indices.Collect(ch, clusterHealthResponse, nodeStatsResponse, indexStatsResponse)
}

func (c *ClusterHealth) fetchAndDecodeClusterHealth() (clusterHealthResponse, error) {
	var chr clusterHealthResponse

	u := *c.url
	u.Path = "/_cluster/health"
	res, err := c.client.Get(u.String())
	if err != nil {
		return chr, fmt.Errorf("failed to get cluster health from %s://%s:%s/%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return chr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&chr); err != nil {
		c.jsonParseFailures.Inc()
		return chr, err
	}

	return chr, nil
}

func (c *Nodes) fetchAndDecodeNodeStats() (nodeStatsResponse, error) {
	var nsr nodeStatsResponse

	u := *c.url
	u.Path = "/_nodes/_local/stats"
	if c.all {
		u.Path = "/_nodes/stats"
	}

	res, err := c.client.Get(u.String())
	if err != nil {
		return nsr, fmt.Errorf("failed to get cluster health from %s://%s:%s/%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nsr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&nsr); err != nil {
		c.jsonParseFailures.Inc()
		return nsr, err
	}
	return nsr, nil
}

func (c *Indices) fetchAndDecodeIndexStats() (indexStatsResponse, error) {
	var isr indexStatsResponse

	u := *c.url
	u.Path = "/_all/_stats"

	res, err := c.client.Get(u.String())
	if err != nil {
		return isr, fmt.Errorf("failed to get index stats from %s://%s:%s/%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return isr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&isr); err != nil {
		c.jsonParseFailures.Inc()
		return isr, err
	}
	return isr, nil
}

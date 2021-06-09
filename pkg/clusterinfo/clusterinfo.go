// Copyright 2021 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clusterinfo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "elasticsearch"
	subsystem = "clusterinfo"
)

var (
	// ErrConsumerAlreadyRegistered is returned if a consumer is already registered
	ErrConsumerAlreadyRegistered = errors.New("consumer already registered")
	// ErrInitialCallTimeout is returned if the initial clusterinfo call timed out
	ErrInitialCallTimeout = errors.New("initial cluster info call timed out")
	initialTimeout        = 10 * time.Second
)

type consumer interface {
	// ClusterLabelUpdates returns a pointer to channel for cluster label updates
	ClusterLabelUpdates() *chan *Response
	// String implements the stringer interface
	String() string
}

// Retriever periodically gets the cluster info from the / endpoint end
// sends it to all registered consumer channels
type Retriever struct {
	consumerChannels      map[string]*chan *Response
	logger                log.Logger
	client                *http.Client
	url                   *url.URL
	interval              time.Duration
	sync                  chan struct{}
	versionMetric         *prometheus.GaugeVec
	up                    *prometheus.GaugeVec
	lastUpstreamSuccessTs *prometheus.GaugeVec
	lastUpstreamErrorTs   *prometheus.GaugeVec
}

// New creates a new Retriever
func New(logger log.Logger, client *http.Client, u *url.URL, interval time.Duration) *Retriever {
	return &Retriever{
		consumerChannels: make(map[string]*chan *Response),
		logger:           logger,
		client:           client,
		url:              u,
		interval:         interval,
		sync:             make(chan struct{}, 1),
		versionMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(namespace, subsystem, "version_info"),
				Help: "Constant metric with ES version information as labels",
			},
			[]string{
				"cluster",
				"cluster_uuid",
				"build_date",
				"build_hash",
				"version",
				"lucene_version",
			},
		),
		up: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(namespace, subsystem, "up"),
				Help: "Up metric for the cluster info collector",
			},
			[]string{"url"},
		),
		lastUpstreamSuccessTs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(namespace, subsystem, "last_retrieval_success_ts"),
				Help: "Timestamp of the last successful cluster info retrieval",
			},
			[]string{"url"},
		),
		lastUpstreamErrorTs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(namespace, subsystem, "last_retrieval_failure_ts"),
				Help: "Timestamp of the last failed cluster info retrieval",
			},
			[]string{"url"},
		),
	}
}

// Describe implements the prometheus.Collector interface
func (r *Retriever) Describe(ch chan<- *prometheus.Desc) {
	r.versionMetric.Describe(ch)
	r.up.Describe(ch)
	r.lastUpstreamSuccessTs.Describe(ch)
	r.lastUpstreamErrorTs.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (r *Retriever) Collect(ch chan<- prometheus.Metric) {
	r.versionMetric.Collect(ch)
	r.up.Collect(ch)
	r.lastUpstreamSuccessTs.Collect(ch)
	r.lastUpstreamErrorTs.Collect(ch)
}

func (r *Retriever) updateMetrics(res *Response) {
	u := *r.url
	u.User = nil
	url := u.String()
	_ = level.Debug(r.logger).Log("msg", "updating cluster info metrics")
	// scrape failed, response is nil
	if res == nil {
		r.up.WithLabelValues(url).Set(0.0)
		r.lastUpstreamErrorTs.WithLabelValues(url).Set(float64(time.Now().Unix()))
		return
	}
	r.up.WithLabelValues(url).Set(1.0)
	r.versionMetric.WithLabelValues(
		res.ClusterName,
		res.ClusterUUID,
		res.Version.BuildDate,
		res.Version.BuildHash,
		res.Version.Number.String(),
		res.Version.LuceneVersion.String(),
	)
	r.lastUpstreamSuccessTs.WithLabelValues(url).Set(float64(time.Now().Unix()))
}

// Update triggers an external cluster info label update
func (r *Retriever) Update() {
	r.sync <- struct{}{}
}

// RegisterConsumer registers a consumer for cluster info updates
func (r *Retriever) RegisterConsumer(c consumer) error {
	if _, registered := r.consumerChannels[c.String()]; registered {
		return ErrConsumerAlreadyRegistered
	}
	r.consumerChannels[c.String()] = c.ClusterLabelUpdates()
	return nil
}

// Run starts the update loop and periodically queries the / endpoint
// The update loop is terminated upon ctx cancellation. The call blocks until the first
// call to the cluster info endpoint was successful
func (r *Retriever) Run(ctx context.Context) error {
	startupComplete := make(chan struct{})
	// start update routine
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				_ = level.Info(r.logger).Log(
					"msg", "context cancelled, exiting cluster info update loop",
					"err", ctx.Err(),
				)
				return
			case <-r.sync:
				_ = level.Info(r.logger).Log(
					"msg", "providing consumers with updated cluster info label",
				)
				res, err := r.fetchAndDecodeClusterInfo()
				if err != nil {
					_ = level.Error(r.logger).Log(
						"msg", "failed to retrieve cluster info from ES",
						"err", err,
					)
					r.updateMetrics(nil)
					continue
				}
				r.updateMetrics(res)
				for name, consumerCh := range r.consumerChannels {
					_ = level.Debug(r.logger).Log(
						"msg", "sending update",
						"consumer", name,
						"res", fmt.Sprintf("%+v", res),
					)
					*consumerCh <- res
				}
				// close startupComplete if not already closed
				select {
				case <-startupComplete:
				default:
					close(startupComplete)
				}
			}
		}
	}(ctx)
	// trigger initial cluster info call
	_ = level.Info(r.logger).Log(
		"msg", "triggering initial cluster info call",
	)
	r.sync <- struct{}{}

	// start a ticker routine
	go func(ctx context.Context) {
		if r.interval <= 0 {
			_ = level.Info(r.logger).Log(
				"msg", "no periodic cluster info label update requested",
			)
			return
		}
		ticker := time.NewTicker(r.interval)
		for {
			select {
			case <-ctx.Done():
				_ = level.Info(r.logger).Log(
					"msg", "context cancelled, exiting cluster info trigger loop",
					"err", ctx.Err(),
				)
				return
			case <-ticker.C:
				_ = level.Debug(r.logger).Log(
					"msg", "triggering periodic update",
				)
				r.sync <- struct{}{}
			}
		}
	}(ctx)

	// block until the first retrieval was successful
	select {
	case <-startupComplete:
		// first sync has been successful
		_ = level.Debug(r.logger).Log("msg", "initial clusterinfo sync succeeded")
		return nil
	case <-time.After(initialTimeout):
		// initial call timed out
		return ErrInitialCallTimeout
	case <-ctx.Done():
		// context cancelled
		return nil
	}
}

func (r *Retriever) fetchAndDecodeClusterInfo() (*Response, error) {
	var response *Response
	u := *r.url
	u.Path = path.Join(r.url.Path, "/")

	res, err := r.client.Get(u.String())
	if err != nil {
		_ = level.Error(r.logger).Log(
			"msg", "failed to get cluster info",
			"err", err,
		)
		return nil, err
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			_ = level.Warn(r.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bts, &response); err != nil {
		return nil, err
	}

	return response, nil
}

package clusterinfo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"net/http"
	"net/url"
	"path"
	"time"
)

var (
	// ErrConsumerAlreadyRegistered is returned if a consumer is already registered
	ErrConsumerAlreadyRegistered = errors.New("consumer already registered")
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
	consumerChannels map[string]*chan *Response
	logger           log.Logger
	client           *http.Client
	url              *url.URL
	interval         time.Duration
	sync             chan struct{}
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
	}
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
func (r *Retriever) Run(ctx context.Context) {
	// start update routine
	go func(ctx context.Context) {
		for range r.sync {
			_ = level.Info(r.logger).Log(
				"msg", "providing consumers with updated cluster info label",
			)
			res, err := r.fetchAndDecodeClusterInfo()
			if err != nil {
				_ = level.Error(r.logger).Log(
					"msg", "failed to retrieve cluster info from ES",
					"err", err,
				)
				continue
			}
			for name, consumerCh := range r.consumerChannels {
				_ = level.Debug(r.logger).Log(
					"msg", "sending update",
					"consumer", name,
					"res", fmt.Sprintf("%+v", res),
				)
				*consumerCh <- res
			}
		}
	}(ctx)
	// trigger initial cluster info call
	r.sync <- struct{}{}

	if r.interval <= 0 {
		_ = level.Info(r.logger).Log(
			"msg", "no periodic cluster info label update requested",
		)
		return
	}
	// start a ticker routine
	go func(ctx context.Context) {
		ticker := time.NewTicker(r.interval)
		for {
			select {
			case <-ctx.Done():
				_ = level.Info(r.logger).Log(
					"msg", "context cancelled, exiting cluster info update loop",
					"err", ctx.Err(),
				)
				return
			case <-ticker.C:
				r.sync <- struct{}{}
			}
		}
	}(ctx)
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

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}

// Copyright The Prometheus Authors
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

package collector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus-community/elasticsearch_exporter/cluster"
	"github.com/prometheus-community/elasticsearch_exporter/config"
	"github.com/prometheus-community/elasticsearch_exporter/internal/esclient"
	"github.com/prometheus-community/elasticsearch_exporter/pkg/clusterinfo"
)

type Runtime struct {
	cfg                  config.Config
	logger               *slog.Logger
	client               *esclient.Client
	clusterInfoProvider  *cluster.InfoProvider
	clusterInfoRetriever *clusterinfo.Retriever
	collectors           []prometheus.Collector
	cancel               context.CancelFunc
}

func NewRuntime(ctx context.Context, logger *slog.Logger, cfg config.Config) (*Runtime, error) {
	if !cfg.Validated() {
		return nil, fmt.Errorf("config has not been validated; call cfg.Validate before NewRuntime")
	}
	if logger == nil {
		logger = slog.Default()
	}
	client, err := esclient.New(cfg, logger)
	if err != nil {
		return nil, err
	}
	runtime := &Runtime{
		cfg:    cfg,
		logger: logger,
		client: client,
		clusterInfoProvider: cluster.NewInfoProvider(
			logger,
			client.HTTP,
			client.URL,
			cfg.ClusterInfoInterval,
		),
		clusterInfoRetriever: clusterinfo.New(
			logger,
			client.HTTP,
			client.URL,
			cfg.ClusterInfoInterval,
		),
	}
	if err := runtime.buildCollectors(ctx); err != nil {
		runtime.Close()
		return nil, err
	}
	return runtime, nil
}

func (r *Runtime) buildCollectors(_ context.Context) error {
	exporter, err := NewElasticsearchCollector(
		r.logger,
		nil,
		WithElasticsearchURL(r.client.URL),
		WithHTTPClient(r.client.HTTP),
		WithClusterInfoProvider(r.clusterInfoProvider),
		WithCollectorStates(r.cfg.Collectors),
		WithCollectorOptions(CollectorOptions{TasksActions: r.cfg.TasksActions}),
	)
	if err != nil {
		return fmt.Errorf("create elasticsearch collector: %w", err)
	}

	collectors := []prometheus.Collector{
		exporter,
		NewClusterHealth(r.logger, r.client.HTTP, r.client.URL),
		NewNodes(r.logger, r.client.HTTP, r.client.URL, r.cfg.AllNodes, r.cfg.Node),
	}

	if r.cfg.ExportIndices || r.cfg.ExportShards {
		shardsCollector := NewShards(r.logger, r.client.HTTP, r.client.URL)
		indicesCollector := NewIndices(r.logger, r.client.HTTP, r.client.URL, r.cfg.ExportShards, r.cfg.ExportIndexAliases)
		collectors = append(collectors, shardsCollector, indicesCollector)
		if err := r.clusterInfoRetriever.RegisterConsumer(indicesCollector); err != nil {
			return fmt.Errorf("register indices collector in cluster info: %w", err)
		}
		if err := r.clusterInfoRetriever.RegisterConsumer(shardsCollector); err != nil {
			return fmt.Errorf("register shards collector in cluster info: %w", err)
		}
	}
	if r.cfg.ExportIndicesMappings {
		collectors = append(collectors, NewIndicesMappings(r.logger, r.client.HTTP, r.client.URL))
	}
	if r.cfg.Collectors[config.CollectorIndicesSettings] {
		collectors = append(collectors, NewIndicesSettings(r.logger, r.client.HTTP, r.client.URL))
	}
	if r.clusterInfoRetriever != nil {
		collectors = append(collectors, r.clusterInfoRetriever)
	}
	r.collectors = collectors
	return nil
}

func (r *Runtime) Start(ctx context.Context) error {
	if r.clusterInfoRetriever == nil {
		return nil
	}
	runCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	switch err := r.clusterInfoRetriever.Run(runCtx); {
	case err == nil:
		r.logger.Info("started cluster info retriever", "interval", r.cfg.ClusterInfoInterval.String())
		return nil
	case errors.Is(err, clusterinfo.ErrInitialCallTimeout):
		r.logger.Info("initial cluster info call timed out")
		return nil
	default:
		cancel()
		r.cancel = nil
		return err
	}
}

func (r *Runtime) Collectors() ([]prometheus.Collector, error) {
	return r.collectors, nil
}

func (r *Runtime) Close() error {
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
	if r.client != nil && r.client.Transport != nil {
		r.client.Transport.CloseIdleConnections()
	}
	return nil
}

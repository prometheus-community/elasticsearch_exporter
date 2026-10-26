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
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// ServerlessIndices exports index storage metrics without using the unsupported
// index stats API.
type ServerlessIndices struct {
	logger *slog.Logger
	client *http.Client
	url    *url.URL
}

type catIndexResponse struct {
	Index        string `json:"index"`
	PrimaryStore string `json:"pri.store.size"`
	TotalStore   string `json:"store.size"`
}

func NewServerlessIndices(logger *slog.Logger, client *http.Client, url *url.URL) *ServerlessIndices {
	return &ServerlessIndices{logger: logger, client: client, url: url}
}

func (i *ServerlessIndices) Describe(ch chan<- *prometheus.Desc) {
	ch <- indicesStoreSizeBytesPrimary
	ch <- indicesStoreSizeBytesTotal
}

func (i *ServerlessIndices) Collect(ch chan<- prometheus.Metric) {
	ctx := context.TODO()
	u := i.url.ResolveReference(&url.URL{Path: "/_cat/indices"})
	query := u.Query()
	query.Set("format", "json")
	query.Set("bytes", "b")
	query.Set("h", "index,pri.store.size,store.size")
	u.RawQuery = query.Encode()

	var indices []catIndexResponse
	if err := getAndDecodeURL(ctx, i.client, i.logger, u.String(), &indices); err != nil {
		i.logger.Warn("failed to fetch Serverless index metrics", "err", err)
		return
	}

	for _, index := range indices {
		primary, err := strconv.ParseFloat(index.PrimaryStore, 64)
		if err != nil {
			i.logger.Warn("invalid primary store size", "index", index.Index, "value", index.PrimaryStore, "err", err)
			continue
		}
		total, err := strconv.ParseFloat(index.TotalStore, 64)
		if err != nil {
			i.logger.Warn("invalid total store size", "index", index.Index, "value", index.TotalStore, "err", err)
			continue
		}
		ch <- prometheus.MustNewConstMetric(indicesStoreSizeBytesPrimary, prometheus.GaugeValue, primary, index.Index, "unknown_cluster")
		ch <- prometheus.MustNewConstMetric(indicesStoreSizeBytesTotal, prometheus.GaugeValue, total, index.Index, "unknown_cluster")
	}
}

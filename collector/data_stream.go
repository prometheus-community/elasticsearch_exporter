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
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	dataStreamBackingIndicesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "data_stream", "backing_indices_total"),
		"Number of backing indices",
		[]string{"data_stream"},
		nil,
	)
	dataStreamStoreSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "data_stream", "store_size_bytes"),
		"Store size of data stream",
		[]string{"data_stream"},
		nil,
	)
)

func init() {
	registerCollector("data-stream", defaultDisabled, NewDataStream)
}

// DataStream Information Struct
type DataStream struct {
	logger *slog.Logger
	hc     *http.Client
	u      *url.URL
}

// NewDataStream defines DataStream Prometheus metrics
func NewDataStream(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &DataStream{
		logger: logger,
		hc:     hc,
		u:      u,
	}, nil
}

// DataStreamStatsResponse is a representation of the Data Stream stats
type DataStreamStatsResponse struct {
	Shards              DataStreamStatsShards       `json:"_shards"`
	DataStreamCount     int64                       `json:"data_stream_count"`
	BackingIndices      int64                       `json:"backing_indices"`
	TotalStoreSizeBytes int64                       `json:"total_store_size_bytes"`
	DataStreamStats     []DataStreamStatsDataStream `json:"data_streams"`
}

// DataStreamStatsShards defines data stream stats shards information structure
type DataStreamStatsShards struct {
	Total      int64 `json:"total"`
	Successful int64 `json:"successful"`
	Failed     int64 `json:"failed"`
}

// DataStreamStatsDataStream defines the structure of per data stream stats
type DataStreamStatsDataStream struct {
	DataStream       string `json:"data_stream"`
	BackingIndices   int64  `json:"backing_indices"`
	StoreSizeBytes   int64  `json:"store_size_bytes"`
	MaximumTimestamp int64  `json:"maximum_timestamp"`
}

func (ds *DataStream) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
	var dsr DataStreamStatsResponse

	u := ds.u.ResolveReference(&url.URL{Path: "/_data_stream/*/_stats"})

	if err := getAndDecodeURL(ctx, ds.hc, ds.logger, u.String(), &dsr); err != nil {
		return err
	}

	for _, dataStream := range dsr.DataStreamStats {
		fmt.Printf("Metric: %+v", dataStream)

		ch <- prometheus.MustNewConstMetric(
			dataStreamBackingIndicesTotal,
			prometheus.CounterValue,
			float64(dataStream.BackingIndices),
			dataStream.DataStream,
		)

		ch <- prometheus.MustNewConstMetric(
			dataStreamStoreSizeBytes,
			prometheus.CounterValue,
			float64(dataStream.StoreSizeBytes),
			dataStream.DataStream,
		)
	}

	return nil
}

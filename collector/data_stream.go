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
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

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

	dataStreamStatsIndexingIndexTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "data_stream_stats", "indexing_index_total"),
		"Total number of documents indexed to the data stream",
		[]string{"data_stream"},
		nil,
	)
	dataStreamStatsIndexingIndexTimeSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "data_stream_stats", "indexing_index_time_seconds_total"),
		"Total time in seconds spent indexing documents to the data stream",
		[]string{"data_stream"},
		nil,
	)
	dataStreamStatsIndexingIndexCurrent = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "data_stream_stats", "indexing_index_current"),
		"Number of documents currently being indexed to the data stream",
		[]string{"data_stream"},
		nil,
	)
	dataStreamStatsIndexingDeleteTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "data_stream_stats", "indexing_delete_total"),
		"Total number of documents deleted from the data stream",
		[]string{"data_stream"},
		nil,
	)
	dataStreamStatsIndexingDeleteTimeSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "data_stream_stats", "indexing_delete_time_seconds_total"),
		"Total time in seconds spent deleting documents from the data stream",
		[]string{"data_stream"},
		nil,
	)
	dataStreamStatsSearchQueryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "data_stream_stats", "search_query_total"),
		"Total number of search queries executed on the data stream",
		[]string{"data_stream"},
		nil,
	)
	dataStreamStatsSearchQueryTimeSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "data_stream_stats", "search_query_time_seconds_total"),
		"Total time in seconds spent on search queries on the data stream",
		[]string{"data_stream"},
		nil,
	)
	dataStreamStatsSearchFetchTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "data_stream_stats", "search_fetch_total"),
		"Total number of search fetch operations on the data stream",
		[]string{"data_stream"},
		nil,
	)
	dataStreamStatsSearchFetchTimeSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "data_stream_stats", "search_fetch_time_seconds_total"),
		"Total time in seconds spent on search fetch operations on the data stream",
		[]string{"data_stream"},
		nil,
	)
	dataStreamStatsDocsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "data_stream_stats", "docs_total"),
		"Total number of documents in the data stream",
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

// dataStreamAggregatedStats holds aggregated index stats for a single data stream
type dataStreamAggregatedStats struct {
	indexingIndexTotal       int64
	indexingIndexTimeMillis  int64
	indexingIndexCurrent     int64
	indexingDeleteTotal      int64
	indexingDeleteTimeMillis int64
	searchQueryTotal         int64
	searchQueryTimeMillis    int64
	searchFetchTotal         int64
	searchFetchTimeMillis    int64
	docsCount                int64
}

func (ds *DataStream) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
	var dsr DataStreamStatsResponse

	u := ds.u.ResolveReference(&url.URL{Path: "/_data_stream/*/_stats"})

	resp, err := getURL(ctx, ds.hc, ds.logger, u.String())
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp, &dsr); err != nil {
		return err
	}

	for _, dataStream := range dsr.DataStreamStats {
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

	if len(dsr.DataStreamStats) == 0 {
		return nil
	}

	// Build comma-separated list of data stream names for index stats query
	dsNames := make([]string, 0, len(dsr.DataStreamStats))
	for _, d := range dsr.DataStreamStats {
		dsNames = append(dsNames, d.DataStream)
	}

	indexStatsURL := ds.u.ResolveReference(&url.URL{Path: "/_all/_stats/indexing,search,docs"})
	q := indexStatsURL.Query()
	q.Set("ignore_unavailable", "true")
	q.Set("filter_path", "indices.*.primaries.indexing,indices.*.primaries.search,indices.*.primaries.docs")
	indexStatsURL.RawQuery = q.Encode()

	indexResp, err := getURL(ctx, ds.hc, ds.logger, indexStatsURL.String())
	if err != nil {
		ds.logger.Warn("failed to fetch index stats for data streams", "err", err)
		return nil
	}

	var isr indexStatsResponse
	if err := json.Unmarshal(indexResp, &isr); err != nil {
		ds.logger.Warn("failed to unmarshal index stats for data streams", "err", err)
		return nil
	}

	// Aggregate index stats per data stream by matching backing index names
	aggregated := make(map[string]*dataStreamAggregatedStats)
	for _, dsName := range dsNames {
		aggregated[dsName] = &dataStreamAggregatedStats{}
	}

	for indexName, indexStats := range isr.Indices {
		dsName := resolveDataStreamName(indexName, dsNames)
		if dsName == "" {
			continue
		}

		agg := aggregated[dsName]
		p := indexStats.Primaries

		agg.indexingIndexTotal += p.Indexing.IndexTotal
		agg.indexingIndexTimeMillis += p.Indexing.IndexTimeInMillis
		agg.indexingIndexCurrent += p.Indexing.IndexCurrent
		agg.indexingDeleteTotal += p.Indexing.DeleteTotal
		agg.indexingDeleteTimeMillis += p.Indexing.DeleteTimeInMillis
		agg.searchQueryTotal += p.Search.QueryTotal
		agg.searchQueryTimeMillis += p.Search.QueryTimeInMillis
		agg.searchFetchTotal += p.Search.FetchTotal
		agg.searchFetchTimeMillis += p.Search.FetchTimeInMillis
		agg.docsCount += p.Docs.Count
	}

	for _, dsName := range dsNames {
		agg := aggregated[dsName]

		ch <- prometheus.MustNewConstMetric(
			dataStreamStatsIndexingIndexTotal,
			prometheus.CounterValue,
			float64(agg.indexingIndexTotal),
			dsName,
		)
		ch <- prometheus.MustNewConstMetric(
			dataStreamStatsIndexingIndexTimeSecondsTotal,
			prometheus.CounterValue,
			float64(agg.indexingIndexTimeMillis)/1000.0,
			dsName,
		)
		ch <- prometheus.MustNewConstMetric(
			dataStreamStatsIndexingIndexCurrent,
			prometheus.GaugeValue,
			float64(agg.indexingIndexCurrent),
			dsName,
		)
		ch <- prometheus.MustNewConstMetric(
			dataStreamStatsIndexingDeleteTotal,
			prometheus.CounterValue,
			float64(agg.indexingDeleteTotal),
			dsName,
		)
		ch <- prometheus.MustNewConstMetric(
			dataStreamStatsIndexingDeleteTimeSecondsTotal,
			prometheus.CounterValue,
			float64(agg.indexingDeleteTimeMillis)/1000.0,
			dsName,
		)
		ch <- prometheus.MustNewConstMetric(
			dataStreamStatsSearchQueryTotal,
			prometheus.CounterValue,
			float64(agg.searchQueryTotal),
			dsName,
		)
		ch <- prometheus.MustNewConstMetric(
			dataStreamStatsSearchQueryTimeSecondsTotal,
			prometheus.CounterValue,
			float64(agg.searchQueryTimeMillis)/1000.0,
			dsName,
		)
		ch <- prometheus.MustNewConstMetric(
			dataStreamStatsSearchFetchTotal,
			prometheus.CounterValue,
			float64(agg.searchFetchTotal),
			dsName,
		)
		ch <- prometheus.MustNewConstMetric(
			dataStreamStatsSearchFetchTimeSecondsTotal,
			prometheus.CounterValue,
			float64(agg.searchFetchTimeMillis)/1000.0,
			dsName,
		)
		ch <- prometheus.MustNewConstMetric(
			dataStreamStatsDocsTotal,
			prometheus.GaugeValue,
			float64(agg.docsCount),
			dsName,
		)
	}

	return nil
}

// resolveDataStreamName matches a backing index name to its data stream.
// Backing indices follow the pattern .ds-{data_stream_name}-{date}-{generation}.
func resolveDataStreamName(indexName string, dsNames []string) string {
	for _, dsName := range dsNames {
		if strings.HasPrefix(indexName, ".ds-"+dsName+"-") {
			return dsName
		}
	}
	return ""
}

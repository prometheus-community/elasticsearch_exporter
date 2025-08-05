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
	"path"
	"sort"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus-community/elasticsearch_exporter/pkg/clusterinfo"
)

var (
	indicesLabels = []string{"index", "cluster"}

	indicesDocsPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "docs_primary"),
		"Count of documents with only primary shards",
		indicesLabels, nil,
	)
	indicesDeletedDocsPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "deleted_docs_primary"),
		"Count of deleted documents with only primary shards",
		indicesLabels, nil,
	)
	indicesDocsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "docs_total"),
		"Total count of documents",
		indicesLabels, nil,
	)
	indicesDeletedDocsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "deleted_docs_total"),
		"Total count of deleted documents",
		indicesLabels, nil,
	)
	indicesStoreSizeBytesPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "store_size_bytes_primary"),
		"Current total size of stored index data in bytes with only primary shards on all nodes",
		indicesLabels, nil,
	)
	indicesStoreSizeBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "store_size_bytes_total"),
		"Current total size of stored index data in bytes with all shards on all nodes",
		indicesLabels, nil,
	)
	indicesSegmentCountPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_count_primary"),
		"Current number of segments with only primary shards on all nodes",
		indicesLabels, nil,
	)
	indicesSegmentCountTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_count_total"),
		"Current number of segments with all shards on all nodes",
		indicesLabels, nil,
	)
	indicesSegmentMemoryBytesPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_memory_bytes_primary"),
		"Current size of segments with only primary shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentMemoryBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_memory_bytes_total"),
		"Current size of segments with all shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentTermsMemoryPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_terms_memory_primary"),
		"Current size of terms with only primary shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentTermsMemoryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_terms_memory_total"),
		"Current number of terms with all shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentFieldsMemoryBytesPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_fields_memory_bytes_primary"),
		"Current size of fields with only primary shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentFieldsMemoryBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_fields_memory_bytes_total"),
		"Current size of fields with all shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentTermVectorsMemoryPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_term_vectors_memory_primary_bytes"),
		"Current size of term vectors with only primary shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentTermVectorsMemoryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_term_vectors_memory_total_bytes"),
		"Current size of term vectors with all shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentNormsMemoryPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_norms_memory_bytes_primary"),
		"Current size of norms with only primary shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentNormsMemoryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_norms_memory_bytes_total"),
		"Current size of norms with all shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentPointsMemoryPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_points_memory_bytes_primary"),
		"Current size of points with only primary shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentPointsMemoryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_points_memory_bytes_total"),
		"Current size of points with all shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentDocValuesMemoryPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_doc_values_memory_bytes_primary"),
		"Current size of doc values with only primary shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentDocValuesMemoryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_doc_values_memory_bytes_total"),
		"Current size of doc values with all shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentIndexWriterMemoryPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_index_writer_memory_bytes_primary"),
		"Current size of index writer with only primary shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentIndexWriterMemoryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_index_writer_memory_bytes_total"),
		"Current size of index writer with all shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentVersionMapMemoryPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_version_map_memory_bytes_primary"),
		"Current size of version map with only primary shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentVersionMapMemoryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_version_map_memory_bytes_total"),
		"Current size of version map with all shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentFBSMemoryPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_fixed_bit_set_memory_bytes_primary"),
		"Current size of fixed bit with only primary shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesSegmentFBSMemoryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "segment_fixed_bit_set_memory_bytes_total"),
		"Current size of fixed bit with all shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesCompletionPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "completion_bytes_primary"),
		"Current size of completion with only primary shards on all nodes in bytes",
		indicesLabels, nil,
	)
	indicesCompletionTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "completion_bytes_total"),
		"Current size of completion with all shards on all nodes in bytes",
		indicesLabels, nil,
	)
	// TODO(@sysadmind): The metrics below should change the subsystem to "indices"
	indicesSearchQueryTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "search_query_time_seconds_total"),
		"Total search query time in seconds",
		indicesLabels, nil,
	)
	indicesActiveQueries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "search", "active_queries"),
		"The number of currently active queries",
		indicesLabels, nil,
	)
	indicesSearchQueryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "search_query_total"),
		"Total number of queries",
		indicesLabels, nil,
	)
	indicesSearchFetchTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "search_fetch_time_seconds_total"),
		"Total search fetch time in seconds",
		indicesLabels, nil,
	)
	indicesSearchFetchTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "search_fetch_total"),
		"Total search fetch count",
		indicesLabels, nil,
	)
	indicesSearchScrollTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "search_scroll_time_seconds_total"),
		"Total search scroll time in seconds",
		indicesLabels, nil,
	)
	indicesSearchScrollCurrent = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "search_scroll_current"),
		"Current search scroll count",
		indicesLabels, nil,
	)
	indicesSearchScrollTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "search_scroll_total"),
		"Total search scroll count",
		indicesLabels, nil,
	)
	indicesSearchSuggestTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "search_suggest_time_seconds_total"),
		"Total search suggest time in seconds",
		indicesLabels, nil,
	)
	indicesSearchSuggestTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "search_suggest_total"),
		"Total search suggest count",
		indicesLabels, nil,
	)
	indicesIndexingTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "indexing_index_time_seconds_total"),
		"Total indexing index time in seconds",
		indicesLabels, nil,
	)
	indicesIndexCurrent = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "index_current"),
		"The number of documents currently being indexed to an index",
		indicesLabels, nil,
	)
	indicesIndexingIndexTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "indexing_index_total"),
		"Total indexing index count",
		indicesLabels, nil,
	)
	indicesIndexingDeleteSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "indexing_delete_time_seconds_total"),
		"Total indexing delete time in seconds",
		indicesLabels, nil,
	)
	indicesIndexingDeleteTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "indexing_delete_total"),
		"Total indexing delete count",
		indicesLabels, nil,
	)
	indicesIndexingNoopUpdateTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "indexing_noop_update_total"),
		"Total indexing no-op update count",
		indicesLabels, nil,
	)
	indicesIndexingThrottleSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "indexing_throttle_time_seconds_total"),
		"Total indexing throttle time in seconds",
		indicesLabels, nil,
	)
	indicesGetTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "get_time_seconds_total"),
		"Total get time in seconds",
		indicesLabels, nil,
	)
	indicesGetTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "get_total"),
		"Total get count",
		indicesLabels, nil,
	)
	indicesMergeTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "merge_time_seconds_total"),
		"Total merge time in seconds",
		indicesLabels, nil,
	)
	indicesMergeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "merge_total"),
		"Total merge count",
		indicesLabels, nil,
	)
	indicesMergeThrottleTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "merge_throttle_time_seconds_total"),
		"Total merge I/O throttle time in seconds",
		indicesLabels, nil,
	)
	indicesMergeStoppedTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "merge_stopped_time_seconds_total"),
		"Total large merge stopped time in seconds, allowing smaller merges to complete",
		indicesLabels, nil,
	)
	indicesMergeAutoThrottleBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "merge_auto_throttle_bytes_total"),
		"Total bytes that were auto-throttled during merging",
		indicesLabels, nil,
	)
	indicesRefreshTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "refresh_time_seconds_total"),
		"Total refresh time in seconds",
		indicesLabels, nil,
	)
	indicesRefreshExternalTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "refresh_external_time_seconds_total"),
		"Total external refresh time in seconds",
		indicesLabels, nil,
	)
	indicesRefreshExternalTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "refresh_external_total"),
		"Total external refresh count",
		indicesLabels, nil,
	)
	indicesRefreshTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "refresh_total"),
		"Total refresh count",
		indicesLabels, nil,
	)
	indicesFlushTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "flush_time_seconds_total"),
		"Total flush time in seconds",
		indicesLabels, nil,
	)
	indicesFlushTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "flush_total"),
		"Total flush count",
		indicesLabels, nil,
	)
	indicesWarmerTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "warmer_time_seconds_total"),
		"Total warmer time in seconds",
		indicesLabels, nil,
	)
	indicesWarmerTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "warmer_total"),
		"Total warmer count",
		indicesLabels, nil,
	)
	indicesQueryCacheMemoryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "query_cache_memory_bytes_total"),
		"Total query cache memory bytes",
		indicesLabels, nil,
	)
	indicesQueryCacheSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "query_cache_size"),
		"Total query cache size",
		indicesLabels, nil,
	)
	indicesQueryCacheHits = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "query_cache_hits_total"),
		"Total query cache hits count",
		indicesLabels, nil,
	)
	indicesQueryCacheMisses = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "query_cache_misses_total"),
		"Total query cache misses count",
		indicesLabels, nil,
	)
	indicesQueryCacheCaches = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "query_cache_caches_total"),
		"Total query cache caches count",
		indicesLabels, nil,
	)
	indicesQueryCacheEvictions = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "query_cache_evictions_total"),
		"Total query cache evictions count",
		indicesLabels, nil,
	)
	indicesRequestCacheMemory = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "request_cache_memory_bytes_total"),
		"Total request cache memory bytes",
		indicesLabels, nil,
	)
	indicesRequestCacheHits = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "request_cache_hits_total"),
		"Total request cache hits count",
		indicesLabels, nil,
	)
	indicesRequestCacheMisses = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "request_cache_misses_total"),
		"Total request cache misses count",
		indicesLabels, nil,
	)
	indicesRequestCacheEvictions = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "request_cache_evictions_total"),
		"Total request cache evictions count",
		indicesLabels, nil,
	)
	indicesFielddataMemory = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "fielddata_memory_bytes_total"),
		"Total fielddata memory bytes",
		indicesLabels, nil,
	)
	indicesFielddataEvictions = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "index_stats", "fielddata_evictions_total"),
		"Total fielddata evictions count",
		indicesLabels, nil,
	)

	indicesAliases = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "aliases"),
		"Record aliases associated with an index",
		[]string{"index", "alias", "cluster"},
		nil,
	)

	indicesShardsLabels = []string{"index", "shard", "node", "primary", "cluster"}

	indicesShardDocs = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "shards_docs"),
		"Count of documents on this shard",
		indicesShardsLabels,
		nil,
	)
	indicesShardDocsDeleted = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "shards_docs_deleted"),
		"Count of deleted documents on this shard",
		indicesShardsLabels,
		nil,
	)
	indicesShardStoreSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices", "shards_store_size_in_bytes"),
		"Store size of this shard",
		indicesShardsLabels,
		nil,
	)
)

type labels struct {
	keys   func(...string) []string
	values func(*clusterinfo.Response, ...string) []string
}

// Indices information struct
type Indices struct {
	logger          *slog.Logger
	client          *http.Client
	url             *url.URL
	shards          bool
	aliases         bool
	clusterInfoCh   chan *clusterinfo.Response
	lastClusterInfo *clusterinfo.Response
}

// NewIndices defines Indices Prometheus metrics
func NewIndices(logger *slog.Logger, client *http.Client, url *url.URL, shards bool, includeAliases bool) *Indices {
	indices := &Indices{
		logger:        logger,
		client:        client,
		url:           url,
		shards:        shards,
		aliases:       includeAliases,
		clusterInfoCh: make(chan *clusterinfo.Response),
		lastClusterInfo: &clusterinfo.Response{
			ClusterName: "unknown_cluster",
		},
	}

	// start go routine to fetch clusterinfo updates and save them to lastClusterinfo
	go func() {
		logger.Debug("starting cluster info receive loop")
		for ci := range indices.clusterInfoCh {
			if ci != nil {
				logger.Debug("received cluster info update", "cluster", ci.ClusterName)
				indices.lastClusterInfo = ci
			}
		}
		logger.Debug("exiting cluster info receive loop")
	}()
	return indices
}

// ClusterLabelUpdates returns a pointer to a channel to receive cluster info updates. It implements the
// (not exported) clusterinfo.consumer interface
func (i *Indices) ClusterLabelUpdates() *chan *clusterinfo.Response {
	return &i.clusterInfoCh
}

// String implements the stringer interface. It is part of the clusterinfo.consumer interface
func (i *Indices) String() string {
	return namespace + "indices"
}

// Describe add Indices metrics descriptions
func (i *Indices) Describe(ch chan<- *prometheus.Desc) {
	ch <- indicesDocsPrimary
	ch <- indicesDeletedDocsPrimary
	ch <- indicesDocsTotal
	ch <- indicesDeletedDocsTotal
	ch <- indicesStoreSizeBytesPrimary
	ch <- indicesStoreSizeBytesTotal
	ch <- indicesSegmentCountPrimary
	ch <- indicesSegmentCountTotal
	ch <- indicesSegmentMemoryBytesPrimary
	ch <- indicesSegmentMemoryBytesTotal
	ch <- indicesSegmentTermsMemoryPrimary
	ch <- indicesSegmentTermsMemoryTotal
	ch <- indicesSegmentFieldsMemoryBytesPrimary
	ch <- indicesSegmentFieldsMemoryBytesTotal
	ch <- indicesSegmentTermVectorsMemoryPrimary
	ch <- indicesSegmentTermVectorsMemoryTotal
	ch <- indicesSegmentNormsMemoryPrimary
	ch <- indicesSegmentNormsMemoryTotal
	ch <- indicesSegmentPointsMemoryPrimary
	ch <- indicesSegmentPointsMemoryTotal
	ch <- indicesSegmentDocValuesMemoryPrimary
	ch <- indicesSegmentDocValuesMemoryTotal
	ch <- indicesSegmentIndexWriterMemoryPrimary
	ch <- indicesSegmentIndexWriterMemoryTotal
	ch <- indicesSegmentVersionMapMemoryPrimary
	ch <- indicesSegmentVersionMapMemoryTotal
	ch <- indicesSegmentFBSMemoryPrimary
	ch <- indicesSegmentFBSMemoryTotal
	ch <- indicesCompletionPrimary
	ch <- indicesCompletionTotal
	ch <- indicesSearchQueryTimeTotal
	ch <- indicesActiveQueries
	ch <- indicesSearchQueryTotal
	ch <- indicesSearchFetchTimeTotal
	ch <- indicesSearchFetchTotal
	ch <- indicesSearchScrollTimeTotal
	ch <- indicesSearchScrollCurrent
	ch <- indicesSearchScrollTotal
	ch <- indicesSearchSuggestTimeTotal
	ch <- indicesSearchSuggestTotal
	ch <- indicesIndexingTimeTotal
	ch <- indicesIndexCurrent
	ch <- indicesIndexingIndexTotal
	ch <- indicesIndexingDeleteSecondsTotal
	ch <- indicesIndexingDeleteTotal
	ch <- indicesIndexingNoopUpdateTotal
	ch <- indicesIndexingThrottleSecondsTotal
	ch <- indicesGetTimeTotal
	ch <- indicesGetTotal
	ch <- indicesMergeTimeTotal
	ch <- indicesMergeTotal
	ch <- indicesMergeThrottleTimeTotal
	ch <- indicesMergeStoppedTimeTotal
	ch <- indicesMergeAutoThrottleBytesTotal
	ch <- indicesRefreshTimeTotal
	ch <- indicesRefreshExternalTimeTotal
	ch <- indicesRefreshExternalTotal
	ch <- indicesRefreshTotal
	ch <- indicesFlushTimeTotal
	ch <- indicesFlushTotal
	ch <- indicesWarmerTimeTotal
	ch <- indicesWarmerTotal
	ch <- indicesQueryCacheMemoryTotal
	ch <- indicesQueryCacheSize
	ch <- indicesQueryCacheHits
	ch <- indicesQueryCacheMisses
	ch <- indicesQueryCacheCaches
	ch <- indicesQueryCacheEvictions
	ch <- indicesRequestCacheMemory
	ch <- indicesRequestCacheHits
	ch <- indicesRequestCacheMisses
	ch <- indicesRequestCacheEvictions
	ch <- indicesFielddataMemory
	ch <- indicesFielddataEvictions

	ch <- indicesAliases

	ch <- indicesShardDocs
	ch <- indicesShardDocsDeleted
	ch <- indicesShardStoreSizeBytes
}

func (i *Indices) fetchAndDecodeIndexStats(ctx context.Context) (indexStatsResponse, error) {
	var isr indexStatsResponse

	u := i.url.ResolveReference(&url.URL{Path: "/_all/_stats"})
	q := u.Query()
	q.Set("ignore_unavailable", "true")
	if i.shards {
		q.Set("level", "shards")
	}
	u.RawQuery = q.Encode()

	resp, err := getURL(ctx, i.client, i.logger, u.String())
	if err != nil {
		return isr, err
	}

	if err := json.Unmarshal(resp, &isr); err != nil {
		return isr, err
	}

	if i.aliases {
		isr.Aliases = map[string][]string{}
		u := i.url.ResolveReference(&url.URL{Path: "_alias"})
		resp, err := getURL(ctx, i.client, i.logger, u.String())
		if err != nil {
			i.logger.Error("error getting alias information", "err", err)
			return isr, err
		}

		var asr aliasesResponse
		if err := json.Unmarshal(resp, &asr); err != nil {
			return isr, err
		}

		for indexName, aliases := range asr {
			var aliasList []string
			for aliasName := range aliases.Aliases {
				aliasList = append(aliasList, aliasName)
			}

			if len(aliasList) > 0 {
				sort.Strings(aliasList)
				isr.Aliases[indexName] = aliasList
			}
		}
	}

	return isr, nil
}

// getClusterName returns the cluster name. If no clusterinfo retriever is
// attached (e.g. /probe mode) it performs a lightweight call to the root
// endpoint once and caches the result.
func (i *Indices) getClusterName() string {
	if i.lastClusterInfo != nil && i.lastClusterInfo.ClusterName != "unknown_cluster" {
		return i.lastClusterInfo.ClusterName
	}
	u := *i.url
	u.Path = path.Join(u.Path, "/")
	resp, err := i.client.Get(u.String())
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var root struct {
				ClusterName string `json:"cluster_name"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&root); err == nil && root.ClusterName != "" {
				i.lastClusterInfo = &clusterinfo.Response{ClusterName: root.ClusterName}
				return root.ClusterName
			}
		}
	}
	return "unknown_cluster"
}

// Collect gets Indices metric values
func (i *Indices) Collect(ch chan<- prometheus.Metric) {
	// indices
	ctx := context.TODO()
	indexStatsResp, err := i.fetchAndDecodeIndexStats(ctx)
	if err != nil {
		i.logger.Warn(
			"failed to fetch and decode index stats",
			"err", err,
		)
		return
	}

	// Alias stats
	if i.aliases {
		for indexName, aliases := range indexStatsResp.Aliases {
			for _, alias := range aliases {
				ch <- prometheus.MustNewConstMetric(
					indicesAliases,
					prometheus.GaugeValue,
					1,
					indexName,
					alias,
					i.getClusterName(),
				)
			}
		}
	}

	// Index stats
	for indexName, indexStats := range indexStatsResp.Indices {
		ch <- prometheus.MustNewConstMetric(
			indicesDocsPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Docs.Count),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesDeletedDocsPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Docs.Deleted),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesDocsTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Docs.Count),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesDeletedDocsTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Docs.Deleted),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesStoreSizeBytesPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Store.SizeInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesStoreSizeBytesTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Store.SizeInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentCountPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Segments.Count),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentCountTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Segments.Count),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentMemoryBytesPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Segments.MemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentMemoryBytesTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Segments.MemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentTermsMemoryPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Segments.TermsMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentTermsMemoryTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Segments.TermsMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentFieldsMemoryBytesPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Segments.StoredFieldsMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentFieldsMemoryBytesTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Segments.StoredFieldsMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentTermVectorsMemoryPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Segments.TermVectorsMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentTermVectorsMemoryTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Segments.TermVectorsMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentNormsMemoryPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Segments.NormsMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentNormsMemoryTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Segments.NormsMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentPointsMemoryPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Segments.PointsMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentPointsMemoryTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Segments.PointsMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentDocValuesMemoryPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Segments.DocValuesMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentDocValuesMemoryTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Segments.DocValuesMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentIndexWriterMemoryPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Segments.IndexWriterMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentIndexWriterMemoryTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Segments.IndexWriterMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentVersionMapMemoryPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Segments.VersionMapMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentVersionMapMemoryTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Segments.VersionMapMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentFBSMemoryPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Segments.FixedBitSetMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSegmentFBSMemoryTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Segments.FixedBitSetMemoryInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesCompletionPrimary,
			prometheus.GaugeValue,
			float64(indexStats.Primaries.Completion.SizeInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesCompletionTotal,
			prometheus.GaugeValue,
			float64(indexStats.Total.Completion.SizeInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSearchQueryTimeTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Search.QueryTimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesActiveQueries,
			prometheus.GaugeValue,
			float64(indexStats.Total.Search.QueryCurrent),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSearchQueryTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Search.QueryTotal),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSearchFetchTimeTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Search.FetchTimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSearchFetchTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Search.FetchTotal),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSearchScrollTimeTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Search.ScrollTimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSearchScrollCurrent,
			prometheus.GaugeValue,
			float64(indexStats.Total.Search.ScrollCurrent),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSearchScrollTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Search.ScrollTotal),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSearchSuggestTimeTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Search.SuggestTimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesSearchSuggestTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Search.SuggestTotal),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesIndexingTimeTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Indexing.IndexTimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesIndexCurrent,
			prometheus.GaugeValue,
			float64(indexStats.Total.Indexing.IndexCurrent),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesIndexingIndexTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Indexing.IndexTotal),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesIndexingDeleteSecondsTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Indexing.DeleteTimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesIndexingDeleteTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Indexing.DeleteTotal),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesIndexingNoopUpdateTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Indexing.NoopUpdateTotal),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesIndexingThrottleSecondsTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Indexing.ThrottleTimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesGetTimeTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Get.TimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesGetTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Get.Total),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesMergeTimeTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Merges.TotalTimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesMergeTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Merges.Total),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesMergeThrottleTimeTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Merges.TotalThrottledTimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesMergeStoppedTimeTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Merges.TotalStoppedTimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesMergeAutoThrottleBytesTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Merges.TotalAutoThrottleInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesRefreshTimeTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Refresh.TotalTimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesRefreshExternalTimeTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Refresh.ExternalTotalTimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesRefreshExternalTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Refresh.ExternalTotal),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesRefreshTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Refresh.Total),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesFlushTimeTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Flush.TotalTimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesFlushTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Flush.Total),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesWarmerTimeTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Warmer.TotalTimeInMillis)/1000,
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesWarmerTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.Warmer.Total),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesQueryCacheMemoryTotal,
			prometheus.CounterValue,
			float64(indexStats.Total.QueryCache.MemorySizeInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesQueryCacheSize,
			prometheus.GaugeValue,
			float64(indexStats.Total.QueryCache.CacheSize),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesQueryCacheHits,
			prometheus.CounterValue,
			float64(indexStats.Total.QueryCache.HitCount),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesQueryCacheMisses,
			prometheus.CounterValue,
			float64(indexStats.Total.QueryCache.MissCount),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesQueryCacheCaches,
			prometheus.CounterValue,
			float64(indexStats.Total.QueryCache.CacheCount),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesQueryCacheEvictions,
			prometheus.CounterValue,
			float64(indexStats.Total.QueryCache.Evictions),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesRequestCacheMemory,
			prometheus.CounterValue,
			float64(indexStats.Total.RequestCache.MemorySizeInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesRequestCacheHits,
			prometheus.CounterValue,
			float64(indexStats.Total.RequestCache.HitCount),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesRequestCacheMisses,
			prometheus.CounterValue,
			float64(indexStats.Total.RequestCache.MissCount),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesRequestCacheEvictions,
			prometheus.CounterValue,
			float64(indexStats.Total.RequestCache.Evictions),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesFielddataMemory,
			prometheus.CounterValue,
			float64(indexStats.Total.Fielddata.MemorySizeInBytes),
			indexName,
			i.getClusterName(),
		)

		ch <- prometheus.MustNewConstMetric(
			indicesFielddataEvictions,
			prometheus.CounterValue,
			float64(indexStats.Total.Fielddata.Evictions),
			indexName,
			i.getClusterName(),
		)

		if i.shards {
			for shardNumber, shards := range indexStats.Shards {
				for _, shard := range shards {
					ch <- prometheus.MustNewConstMetric(
						indicesShardDocs,
						prometheus.GaugeValue,
						float64(shard.Docs.Count),
						indexName,
						shardNumber,
						shard.Routing.Node,
						strconv.FormatBool(shard.Routing.Primary),
						i.getClusterName(),
					)
					ch <- prometheus.MustNewConstMetric(
						indicesShardDocsDeleted,
						prometheus.GaugeValue,
						float64(shard.Docs.Deleted),
						indexName,
						shardNumber,
						shard.Routing.Node,
						strconv.FormatBool(shard.Routing.Primary),
						i.getClusterName(),
					)
					ch <- prometheus.MustNewConstMetric(
						indicesShardStoreSizeBytes,
						prometheus.GaugeValue,
						float64(shard.Store.SizeInBytes),
						indexName,
						shardNumber,
						shard.Routing.Node,
						strconv.FormatBool(shard.Routing.Primary),
						i.getClusterName(),
					)
				}
			}
		}
	}
}

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

// Package collector includes all individual collectors to gather and export elasticsearch metrics.
package collector

import (
	"context"

	"github.com/prometheus-community/elasticsearch_exporter/cluster"
)

type UpdateContext interface {
	// GetClusterInfo returns the current cluster info.
	GetClusterInfo(context.Context) (cluster.Info, error)
}

// DefaultUpdateContext is the default implementation of UpdateContext.
type DefaultUpdateContext struct {
	clusterInfo *cluster.InfoProvider
}

// NewDefaultUpdateContext creates a new DefaultUpdateContext.
func NewDefaultUpdateContext(clusterInfo *cluster.InfoProvider) *DefaultUpdateContext {
	return &DefaultUpdateContext{clusterInfo: clusterInfo}
}

// Retriever returns the cluster info retriever.
func (c *DefaultUpdateContext) GetClusterInfo(ctx context.Context) (cluster.Info, error) {
	info, err := c.clusterInfo.GetInfo(ctx)
	if err != nil {
		return cluster.Info{}, err
	}
	return info, nil
}

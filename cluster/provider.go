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

package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"
)

type Info struct {
	ClusterName string `json:"cluster_name"`
}

type InfoProvider struct {
	logger          *slog.Logger
	client          *http.Client
	url             *url.URL
	interval        time.Duration
	lastClusterInfo Info
	lastError       error
	cachedAt        time.Time    // Time when the last cluster info was fetched
	mu              sync.RWMutex // Protects lastClusterInfo, lastError, and cachedAt
}

// New creates a new Retriever.
func NewInfoProvider(logger *slog.Logger, client *http.Client, u *url.URL, interval time.Duration) *InfoProvider {
	return &InfoProvider{
		logger:   logger,
		client:   client,
		url:      u,
		interval: interval,
	}
}

func (i *InfoProvider) GetInfo(ctx context.Context) (Info, error) {
	i.mu.RLock()
	info := i.lastClusterInfo
	err := i.lastError
	cachedAt := i.cachedAt

	i.mu.RUnlock()

	// If the cached info is recent enough, return it.
	if !cachedAt.IsZero() && time.Since(cachedAt) < i.interval {
		if err != nil {
			return Info{}, err
		}

		if info.ClusterName != "" {
			return info, nil
		}
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	// If we reach here, we need to fetch the cluster info. The cache is either empty or stale.
	u := *i.url
	u.Path = path.Join(u.Path, "/")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Info{}, err
	}

	resp, err := i.client.Do(req)
	if err != nil {
		i.logger.Error("failed to fetch cluster info", "err", err)
		return Info{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		i.lastError = err
		return Info{}, err
	}

	var infoResponse Info
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		i.lastError = err
		return Info{}, err
	}

	if err := json.Unmarshal(body, &infoResponse); err != nil {
		i.lastError = err
		return Info{}, fmt.Errorf("failed to unmarshal cluster info: %w", err)
	}

	info = Info{ClusterName: infoResponse.ClusterName}
	i.lastClusterInfo = info
	i.lastError = nil
	i.cachedAt = time.Now()

	return info, nil
}

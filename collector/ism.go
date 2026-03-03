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
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	ismIndexStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ism_index", "status"),
		"Status of ISM policy for index (OpenSearch Index State Management)",
		[]string{"index", "policy_id", "state", "action", "step", "step_status"},
		nil,
	)

	ismIndexFailed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ism_index", "failed"),
		"Whether ISM is currently in a failed step/action for the index (OpenSearch Index State Management)",
		[]string{"index", "policy_id", "state", "action", "step", "step_status"},
		nil,
	)
)

func init() {
	registerCollector("ism", defaultEnabled, NewISM)
}

type ISM struct {
	logger *slog.Logger
	hc     *http.Client
	u      *url.URL
}

func NewISM(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &ISM{
		logger: logger,
		hc:     hc,
		u:      u,
	}, nil
}

type ismExplainIndex struct {
	Index    string `json:"index"`
	PolicyID string `json:"policy_id"`

	// Some versions omit this; treat missing as "enabled".
	Enabled *bool `json:"enabled"`

	State     json.RawMessage `json:"state"`
	Action    json.RawMessage `json:"action"`
	Step      json.RawMessage `json:"step"`
	RetryInfo json.RawMessage `json:"retry_info"`

	// Some versions also include a top-level "failed_step".
	FailedStep string `json:"failed_step"`
}

func (c *ISM) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Endpoint changed across distributions/versions:
	// - OpenSearch:  /_plugins/_ism/...
	// - ODFE:        /_opendistro/_ism/...
	//
	// Keep payload minimal: we only need current state/action/step.
	paths := []string{"/_plugins/_ism/explain/*", "/_opendistro/_ism/explain/*"}
	var resp []byte
	var err error
	for i, p := range paths {
		explainURL := c.u.ResolveReference(&url.URL{
			Path:     p,
			RawQuery: "show_policy=false&validate_action=false",
		})

		resp, err = getURLNoDataOn404(ctx, c.hc, c.logger, explainURL.String())
		if err != nil {
			if IsNoDataError(err) && i+1 < len(paths) {
				continue
			}
			return err
		}
		break
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(resp, &raw); err != nil {
		return fmt.Errorf("failed to decode ISM explain JSON body: %w", err)
	}

	for key, v := range raw {
		if key == "total_managed_indices" {
			continue
		}

		var idx ismExplainIndex
		if err := json.Unmarshal(v, &idx); err != nil {
			return fmt.Errorf("failed to decode ISM explain index body for %q: %w", key, err)
		}

		var fields map[string]json.RawMessage
		_ = json.Unmarshal(v, &fields)

		indexName := idx.Index
		if indexName == "" {
			indexName = key
		}

		policyID := idx.PolicyID
		if policyID == "" && fields != nil {
			for _, k := range []string{
				"index.plugins.index_state_management.policy_id",
				"index.opendistro.index_state_management.policy_id",
			} {
				if rm, ok := fields[k]; ok {
					_ = json.Unmarshal(rm, &policyID)
					if policyID != "" {
						break
					}
				}
			}
		}

		enabled := true
		if idx.Enabled != nil {
			enabled = *idx.Enabled
		}

		stateName := extractNameField(idx.State)
		actionName := extractNameField(idx.Action)
		stepName := extractNameField(idx.Step)
		stepStatus := extractStringField(idx.Step, "step_status")

		actionFailed := extractBoolField(idx.Action, "failed")
		retryFailed := extractBoolField(idx.RetryInfo, "failed")
		failed := actionFailed || retryFailed || stepStatus == "failed" || idx.FailedStep != ""

		ch <- prometheus.MustNewConstMetric(
			ismIndexStatus,
			prometheus.GaugeValue,
			bool2Float(enabled),
			indexName, policyID, stateName, actionName, stepName, stepStatus,
		)

		ch <- prometheus.MustNewConstMetric(
			ismIndexFailed,
			prometheus.GaugeValue,
			bool2Float(failed),
			indexName, policyID, stateName, actionName, stepName, stepStatus,
		)
	}

	return nil
}

func extractNameField(raw json.RawMessage) string {
	// Common shape: {"name":"..."}
	var obj struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil && obj.Name != "" {
		return obj.Name
	}

	// Less common shape: "name"
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}

	return ""
}

func extractStringField(raw json.RawMessage, field string) string {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return ""
	}
	rm, ok := obj[field]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(rm, &s); err != nil {
		return ""
	}
	return s
}

func extractBoolField(raw json.RawMessage, field string) bool {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return false
	}
	rm, ok := obj[field]
	if !ok {
		return false
	}
	var b bool
	if err := json.Unmarshal(rm, &b); err != nil {
		return false
	}
	return b
}

func getURLNoDataOn404(ctx context.Context, hc *http.Client, log *slog.Logger, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Warn("failed to close response body", "err", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNoData
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Request failed with code %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

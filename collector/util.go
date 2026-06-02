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
)

// fetchURL performs an HTTP GET against u and passes the response body to
// consume, handling the request setup, status check, and body close. It lets
// callers stream large responses instead of buffering them in full.
func fetchURL(ctx context.Context, hc *http.Client, log *slog.Logger, u string, consume func(io.Reader) error) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Warn("failed to close response body", "err", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP Request failed with code %d", resp.StatusCode)
	}

	return consume(resp.Body)
}

// getAndDecodeURL performs an HTTP GET against u and unmarshals the JSON
// response body into target. It consolidates the request/read/decode/body-close
// boilerplate that the collectors previously duplicated.
func getAndDecodeURL(ctx context.Context, hc *http.Client, log *slog.Logger, u string, target any) error {
	return fetchURL(ctx, hc, log, u, func(r io.Reader) error {
		b, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, target)
	})
}

// bool2Float converts a bool to a float64. True is 1, false is 0.
func bool2Float(managed bool) float64 {
	if managed {
		return 1
	}
	return 0
}

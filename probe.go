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

package main

import (
	"errors"
	"net/url"
	"strings"

	"github.com/prometheus-community/elasticsearch_exporter/config"
)

var (
	errMissingTarget     = errors.New("missing target parameter")
	errInvalidTarget     = errors.New("invalid target parameter")
	errModuleNotFound    = errors.New("auth_module not found")
	errUnsupportedModule = errors.New("unsupported auth_module type")
)

// validateProbeParams performs upfront validation of the query parameters.
// It returns the target string (as given), the resolved AuthModule (optional), or an error.
func validateProbeParams(cfg *config.Config, q url.Values) (string, *config.AuthModule, error) {
	target := q.Get("target")
	if target == "" {
		return "", nil, errMissingTarget
	}

	// If the target does not contain an URL scheme, default to http.
	// This allows users to pass "host:port" without the "http://" prefix.
	if !strings.Contains(target, "://") {
		target = "http://" + target
	}

	u, err := url.Parse(target)
	if err != nil {
		return "", nil, errInvalidTarget
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", nil, errInvalidTarget
	}

	modu := q.Get("auth_module")
	if modu == "" {
		return target, nil, nil // no auth module requested
	}
	if cfg == nil {
		return "", nil, errModuleNotFound
	}
	am, ok := cfg.AuthModules[modu]
	if !ok {
		return "", nil, errModuleNotFound
	}
	switch strings.ToLower(am.Type) {
	case "userpass":
		return target, &am, nil
	case "apikey":
		return target, &am, nil
	case "aws":
		// Accept module even if region omitted; environment resolver can provide it.
		return target, &am, nil
	case "tls":
		// TLS auth type is valid; detailed TLS validation is performed during config load.
		return target, &am, nil
	default:
		return "", nil, errUnsupportedModule
	}
}

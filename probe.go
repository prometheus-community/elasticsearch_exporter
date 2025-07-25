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
	if _, err := url.Parse(target); err != nil {
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
		if am.UserPass != nil {
			return target, &am, nil
		}
		return "", nil, errUnsupportedModule
	case "apikey":
		if am.APIKey != "" {
			return target, &am, nil
		}
		return "", nil, errUnsupportedModule
	default:
		return "", nil, errUnsupportedModule
	}
}

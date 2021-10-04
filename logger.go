// Copyright 2021 The Prometheus Authors
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
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"io"
	"strings"
)

func getLogger(loglevel, logoutput, logfmt string) log.Logger {
	var out *os.File
	switch strings.ToLower(logoutput) {
	case "stderr":
		out = os.Stderr
	case "stdout":
		out = os.Stdout
	default:
		out = os.Stdout
	}
	var logCreator func(io.Writer) log.Logger
	switch strings.ToLower(logfmt) {
	case "json":
		logCreator = log.NewJSONLogger
	case "logfmt":
		logCreator = log.NewLogfmtLogger
	default:
		logCreator = log.NewLogfmtLogger
	}

	// create a logger
	logger := logCreator(log.NewSyncWriter(out))

	// set loglevel
	var loglevelFilterOpt level.Option
	switch strings.ToLower(loglevel) {
	case "debug":
		loglevelFilterOpt = level.AllowDebug()
	case "info":
		loglevelFilterOpt = level.AllowInfo()
	case "warn":
		loglevelFilterOpt = level.AllowWarn()
	case "error":
		loglevelFilterOpt = level.AllowError()
	default:
		loglevelFilterOpt = level.AllowInfo()
	}
	logger = level.NewFilter(logger, loglevelFilterOpt)
	logger = log.With(logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)
	return logger
}

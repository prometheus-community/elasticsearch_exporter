package main

import (
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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

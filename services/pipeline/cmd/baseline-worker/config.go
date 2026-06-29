package main

import (
	"fmt"
	"time"

	"github.com/JDucr17/streamline/services/pipeline/internal/config"
)

const (
	envKafkaBrokers              = "KAFKA_BROKERS"
	envKafkaBaselineSignalsTopic = "KAFKA_BASELINE_SIGNALS_TOPIC"
	envBaselineWindowDays        = "BASELINE_WINDOW_DAYS"
	envActivityWindowSeconds     = "ACTIVITY_WINDOW_SECONDS"
	envWindowCadenceSeconds      = "WINDOW_CADENCE_SECONDS"
	envMinWindowEvents           = "MIN_WINDOW_EVENTS"
	envMinBaselineWindows        = "MIN_BASELINE_WINDOWS"
	envMinBaselineVisitors       = "MIN_BASELINE_VISITORS"
)

const startupTimeout = 10 * time.Second

const (
	defaultBaselineWindowDays = 30

	// Floor on the fit window, below this a source may have too little recent history
	// for a stable baseline
	minBaselineWindowDays = 7

	// The activity window is the grouping unit (how events form a window),
	// distinct from BASELINE_WINDOW_DAYS, the history horizon (which events are read).
	defaultActivityWindowSeconds = 300
	defaultWindowCadenceSeconds  = 60
	defaultMinWindowEvents       = 3
	defaultMinBaselineWindows    = 3000
	defaultMinBaselineVisitors   = 100
)

type appConfig struct {
	databaseURL          string
	kafkaBrokers         []string
	baselineSignalsTopic string
	baselineWindowDays   int

	activityWindowSeconds int
	windowCadenceSeconds  int
	minWindowEvents       int
	minBaselineWindows    int
	minBaselineVisitors   int
}

func loadConfig() (appConfig, error) {
	var opt config.OptionalVars
	windowDays := opt.IntAtLeast(envBaselineWindowDays, defaultBaselineWindowDays, minBaselineWindowDays)
	activityWindow := opt.IntAtLeast(envActivityWindowSeconds, defaultActivityWindowSeconds, 1)
	cadence := opt.IntAtLeast(envWindowCadenceSeconds, defaultWindowCadenceSeconds, 1)
	minWindowEvents := opt.IntAtLeast(envMinWindowEvents, defaultMinWindowEvents, 1)
	minWindows := opt.IntAtLeast(envMinBaselineWindows, defaultMinBaselineWindows, 1)
	minVisitors := opt.IntAtLeast(envMinBaselineVisitors, defaultMinBaselineVisitors, 1)
	if err := opt.Err(); err != nil {
		return appConfig{}, err
	}

	// Cadence must not exceed the window it samples
	if cadence > activityWindow {
		return appConfig{}, fmt.Errorf("%s (%d) must be <= %s (%d)",
			envWindowCadenceSeconds, cadence, envActivityWindowSeconds, activityWindow)
	}

	var req config.RequiredVars
	cfg := appConfig{
		databaseURL:           req.Get("DATABASE_URL"),
		kafkaBrokers:          req.CSV(envKafkaBrokers),
		baselineSignalsTopic:  req.Get(envKafkaBaselineSignalsTopic),
		baselineWindowDays:    windowDays,
		activityWindowSeconds: activityWindow,
		windowCadenceSeconds:  cadence,
		minWindowEvents:       minWindowEvents,
		minBaselineWindows:    minWindows,
		minBaselineVisitors:   minVisitors,
	}
	if err := req.Err(); err != nil {
		return appConfig{}, err
	}

	return cfg, nil
}
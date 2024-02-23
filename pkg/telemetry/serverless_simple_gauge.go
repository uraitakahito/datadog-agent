// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build serverless
// +build serverless

package telemetry

import (
	"github.com/DataDog/datadog-agent/comp/core/telemetry"
)

// SimpleGauge tracks how many times something is happening.
type SimpleGauge interface {
	telemetry.SimpleGauge
}

// NewSimpleGauge creates a new SimpleGauge with default options.
func NewSimpleGauge(subsystem, name, help string) SimpleGauge {
	return NewSimpleGaugeWithOpts(subsystem, name, help, telemetry.DefaultOptions)
}

// NewSimpleGaugeWithOpts creates a new SimpleGauge.
func NewSimpleGaugeWithOpts(subsystem, name, help string, opts telemetry.Options) SimpleGauge {
	return telemetry.GetCompatComponent().NewSimpleGaugeWithOpts(subsystem, name, help, telemetry.Options(opts))
}

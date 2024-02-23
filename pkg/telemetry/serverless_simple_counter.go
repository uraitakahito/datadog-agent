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

// SimpleCounter tracks how many times something is happening.
type SimpleCounter interface {
	telemetry.SimpleCounter
}

// NewSimpleCounter creates a new SimpleCounter with default options.
func NewSimpleCounter(subsystem, name, help string) SimpleCounter {
	return NewSimpleCounterWithOpts(subsystem, name, help, DefaultOptions)
}

// NewSimpleCounterWithOpts creates a new SimpleCounter.
func NewSimpleCounterWithOpts(subsystem, name, help string, opts Options) SimpleCounter {
	return telemetry.GetCompatComponent().NewSimpleCounterWithOpts(subsystem, name, help, telemetry.Options(opts))
}

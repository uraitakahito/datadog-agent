// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build !serverless
// +build !serverless

package telemetry

import (
	"github.com/DataDog/datadog-agent/comp/core/telemetry"
)

// Histogram tracks the value of one health metric of the Agent.
type Histogram interface {
	telemetry.Histogram
}

// NewHistogram creates a Histogram with default options for telemetry purpose.
// Current implementation used: Prometheus Histogram
func NewHistogram(subsystem, name string, tags []string, help string, buckets []float64) Histogram {
	return NewHistogramWithOpts(subsystem, name, tags, help, buckets, DefaultOptions)
}

// NewHistogramWithOpts creates a Histogram with the given options for telemetry purpose.
// See NewHistogram()
func NewHistogramWithOpts(subsystem, name string, tags []string, help string, buckets []float64, opts telemetry.Options) Histogram {
	return GetCompatComponent().NewHistogramWithOpts(subsystem, name, tags, help, buckets, telemetry.Options(opts))
}

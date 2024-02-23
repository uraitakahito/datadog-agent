// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package telemetry

import "github.com/DataDog/datadog-agent/comp/core/telemetry"

type histogramNoOp struct{}

func (h histogramNoOp) Observe(_ float64, _ ...string)                           {}
func (h histogramNoOp) Delete(_ ...string)                                       {}
func (h histogramNoOp) WithValues(tagsValue ...string) telemetry.SimpleHistogram { return nil } //nolint:revive // TODO fix revive unused-parameter
func (h histogramNoOp) WithTags(tags map[string]string) telemetry.SimpleHistogram { //nolint:revive // TODO fix revive unused-parameter
	return nil
}

// NewHistogramNoOp creates a dummy Histogram
func NewHistogramNoOp() Histogram {
	return histogramNoOp{}
}

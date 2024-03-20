// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package collector implements the OTLP Collector component for non-OTLP builds.
package collectorimpl

/*
// Component represents the no-op Component interface.
type Component interface {
	Start() error
	Stop()
}
*/

type noOpComp struct{}

// Start is a no-op.
func (noOpComp) Start() error { return nil }

// Stop is a no-op.
func (noOpComp) Stop() {}

func NewAgentComponent() (collectordef.Component, error) {
	return noOpComp{}, nil
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

//go:build !serverless
// +build !serverless

package telemetryimpl

import (
	"github.com/DataDog/datadog-agent/comp/core/telemetry"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"go.uber.org/fx"

	"github.com/prometheus/client_golang/prometheus"
	sdk "go.opentelemetry.io/otel/sdk/metric"
)

// Mock implements mock-specific methods.
type Mock interface {
	prometheusComponent

	GetRegistry() *prometheus.Registry
	GetMeterProvider() *sdk.MeterProvider
}

type telemetryImplMock struct {
	telemetryImpl
}

func newMock() Mock {
	reg := prometheus.NewRegistry()
	provider := newProvider(reg)

	telemetry := &telemetryImplMock{
		telemetryImpl{
			mutex:         &mutex,
			registry:      reg,
			meterProvider: provider,
		},
	}

	return telemetry
}

func (t *telemetryImplMock) GetRegistry() *prometheus.Registry {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.registry
}

func (t *telemetryImplMock) GetMeterProvider() *sdk.MeterProvider {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.meterProvider
}

// MockModule defines the fx options for the mock component.
func MockModule() fxutil.Module {
	return fxutil.Component(
		fx.Provide(newMock),
		fx.Provide(func(m Mock) telemetry.Component { return m }))
}

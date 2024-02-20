// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

package serverlessimpl

import (
	"net/http"
	"sync"

	"github.com/DataDog/datadog-agent/comp/core/telemetry"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"go.uber.org/fx"
)

type serverlessComponent interface {
	telemetry.Component
}

type serverlessImpl struct {
	mutex *sync.Mutex
}

func newTelemetry() serverlessComponent {
	return &serverlessImpl{
		mutex: &sync.Mutex{},
	}
}

// GetCompatComponent returns a component wrapping telemetry global variables
// TODO (components): Remove this when all telemetry is migrated to the component
func GetCompatComponent() telemetry.Component {
	return newTelemetry()
}

func (t *serverlessImpl) Handler() http.Handler {
	//TODO
}

func (t *serverlessImpl) Reset() {
	mutex.Lock()
	defer mutex.Unlock()
	// TODO
}

func (t *serverlessImpl) NewCounter(subsystem, name string, tags []string, help string) telemetry.Counter {
	return t.NewCounterWithOpts(subsystem, name, tags, help, telemetry.DefaultOptions)
}

func (t *serverlessImpl) NewCounterWithOpts(subsystem, name string, tags []string, help string, opts telemetry.Options) telemetry.Counter {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	name = opts.NameWithSeparator(subsystem, name)

	// TODO: implementation
}

func (t *serverlessImpl) NewSimpleCounter(subsystem, name, help string) telemetry.SimpleCounter {
	return t.NewSimpleCounterWithOpts(subsystem, name, help, telemetry.DefaultOptions)
}

func (t *serverlessImpl) NewSimpleCounterWithOpts(subsystem, name, help string, opts telemetry.Options) telemetry.SimpleCounter {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	name = opts.NameWithSeparator(subsystem, name)

	// TODO: implementation

}

func (t *serverlessImpl) NewGauge(subsystem, name string, tags []string, help string) telemetry.Gauge {
	return t.NewGaugeWithOpts(subsystem, name, tags, help, telemetry.DefaultOptions)
}

func (t *serverlessImpl) NewGaugeWithOpts(subsystem, name string, tags []string, help string, opts telemetry.Options) telemetry.Gauge {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	name = opts.NameWithSeparator(subsystem, name)

	// TODO: implementation

}

func (t *serverlessImpl) NewSimpleGauge(subsystem, name, help string) telemetry.SimpleGauge {
	return t.NewSimpleGaugeWithOpts(subsystem, name, help, telemetry.DefaultOptions)
}

func (t *serverlessImpl) NewSimpleGaugeWithOpts(subsystem, name, help string, opts telemetry.Options) telemetry.SimpleGauge {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	name = opts.NameWithSeparator(subsystem, name)

	// TODO: implementation

}

func (t *serverlessImpl) NewHistogram(subsystem, name string, tags []string, help string, buckets []float64) telemetry.Histogram {
	return t.NewHistogramWithOpts(subsystem, name, tags, help, buckets, telemetry.DefaultOptions)
}

func (t *serverlessImpl) NewHistogramWithOpts(subsystem, name string, tags []string, help string, buckets []float64, opts telemetry.Options) telemetry.Histogram {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	name = opts.NameWithSeparator(subsystem, name)

	// TODO: implementation

}

func (t *serverlessImpl) NewSimpleHistogram(subsystem, name, help string, buckets []float64) telemetry.SimpleHistogram {
	return t.NewSimpleHistogramWithOpts(subsystem, name, help, buckets, telemetry.DefaultOptions)
}

func (t *serverlessImpl) NewSimpleHistogramWithOpts(subsystem, name, help string, buckets []float64, opts telemetry.Options) telemetry.SimpleHistogram {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	name = opts.NameWithSeparator(subsystem, name)

	// TODO: implementation

}

// Module defines the fx options for this component.
func Module() fxutil.Module {
	return fxutil.Component(
		fx.Provide(newTelemetry))
}

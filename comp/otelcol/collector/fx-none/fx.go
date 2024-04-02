package collectorfx

import (
	collectorimpl "github.com/DataDog/datadog-agent/comp/otelcol/collector/impl-none"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"go.uber.org/fx"
)

// Module specifies the fx module for non-OTLP builds.
func Module() fxutil.Module {
	return fxutil.Component(
		fx.Provide(collectorimpl.NewAgentComponent))
}

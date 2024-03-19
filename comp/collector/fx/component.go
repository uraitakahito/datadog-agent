package collectorfx

import (
	"github.com/DataDog/datadog-agent/comp/otelcol/collector"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"github.com/DataDog/datadog-agent/pkg/util/optional"
	"go.uber.org/fx"
)

// Module defines the fx options for this component.
func Module() fxutil.Module {
	return fxutil.Component(
		fx.Provide(collectorimpl.NewAgentComponents),
		fx.Provide(func(c collector.Component) optional.Option[collector.Component] {
			return optional.NewOption[collector.Component](c)
		}),
	)
}

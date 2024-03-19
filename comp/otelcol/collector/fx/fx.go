package collectorfx

import (
	collectorimpl "github.com/DataDog/datadog-agent/comp/otelcol/collector/impl"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"go.uber.org/fx"
)

type dependencies struct {
	fx.In
	collectorimpl.Dependencies

	/*
	   // Lc specifies the fx lifecycle settings, used for appending startup
	   // and shutdown hooks.
	   Lc fx.Lifecycle

	   // Config specifies the Datadog Agent's configuration component.
	   Config config.Component

	   // Log specifies the logging component.
	   Log corelog.Component

	   // Serializer specifies the metrics serializer that is used to export metrics
	   // to Datadog.
	   Serializer serializer.MetricSerializer

	   // LogsAgent specifies a logs agent
	   LogsAgent optional.Option[logsagent.Component]

	   // InventoryAgent require the inventory metadata payload, allowing otelcol to add data to it.
	   InventoryAgent inventoryagent.Component
	*/
}

type provides struct {
	fx.Out
	collectorimpl.Provides
	/*
		Comp           collectortype.Component
		StatusProvider status.InformationProvider
	*/
}

// Module specifies the Collector module bundle.
func Module() fxutil.Module {
	return fxutil.Component(
		fx.Provide(collectorimpl.NewAgentComponents))
}

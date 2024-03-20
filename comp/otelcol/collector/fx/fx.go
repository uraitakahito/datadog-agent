package collectorfx

import (
	"github.com/DataDog/datadog-agent/comp/core/config"
	corelog "github.com/DataDog/datadog-agent/comp/core/log"
	"github.com/DataDog/datadog-agent/comp/core/status"
	logsagent "github.com/DataDog/datadog-agent/comp/logs/agent"
	"github.com/DataDog/datadog-agent/comp/metadata/inventoryagent"
	collectordef "github.com/DataDog/datadog-agent/comp/otelcol/collector/def"
	collectorimpl "github.com/DataDog/datadog-agent/comp/otelcol/collector/impl"
	"github.com/DataDog/datadog-agent/pkg/serializer"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"github.com/DataDog/datadog-agent/pkg/util/optional"
	"go.uber.org/fx"
)

// TODO: It would be good if this struct could be anonymously generated using reflection in a
// helper function
type dependencies struct {
	fx.In

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
}

type provides struct {
	fx.Out
	Comp           collectordef.Component
	StatusProvider status.InformationProvider
}

// Module specifies the Collector module bundle.
func Module() fxutil.Module {
	return fxutil.Component(
		fx.Provide(func(deps dependencies) provides {
			inputs := collectorimpl.Inputs{
				Config:         deps.Config,
				Log:            deps.Log,
				Serializer:     deps.Serializer,
				LogsAgent:      deps.LogsAgent,
				InventoryAgent: deps.InventoryAgent,
			}
			outputs, _ := collectorimpl.NewAgentComponents(inputs)
			return provides{
				Comp:           outputs.Comp,
				StatusProvider: outputs.StatusProvider,
			}
		}))
}

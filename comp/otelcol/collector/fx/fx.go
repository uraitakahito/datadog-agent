package collectorfx

import (
	"github.com/DataDog/datadog-agent/comp/core/config"
	corelog "github.com/DataDog/datadog-agent/comp/core/log"
	"github.com/DataDog/datadog-agent/comp/core/status"
	logsagent "github.com/DataDog/datadog-agent/comp/logs/agent"
	"github.com/DataDog/datadog-agent/comp/metadata/inventoryagent"
	"github.com/DataDog/datadog-agent/comp/otelcol/collector"
	"github.com/DataDog/datadog-agent/pkg/serializer"
	"github.com/DataDog/datadog-agent/pkg/util/optional"
	"go.uber.org/fx"
)

type dependencies struct {
	fx.In

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
}

type provides struct {
	fx.Out

	Comp           collector.Component
	StatusProvider status.InformationProvider
}

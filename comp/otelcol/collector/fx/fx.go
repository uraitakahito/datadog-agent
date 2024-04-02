package collectorfx

import (
	collectorimpl "github.com/DataDog/datadog-agent/comp/otelcol/collector/impl"
	"github.com/DataDog/datadog-agent/pkg/util/fxhelper"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
)

// Module specifies the Collector module bundle.
func Module() fxutil.Module {
	return fxutil.Component(fxhelper.ProvideComponentConstructor(collectorimpl.NewAgentComponents))
}

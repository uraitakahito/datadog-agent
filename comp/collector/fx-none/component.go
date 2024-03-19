package collectorfx

import (
	"github.com/DataDog/datadog-agent/comp/otelcol/collector"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"github.com/DataDog/datadog-agent/pkg/util/optional"
	"go.uber.org/fx"
)

// NoneModule return a None optional type for Component.
// This helper allows code that needs a disabled Optional type for the collector to get it. The helper is split from
// the implementation to avoid linking with the implementation.
func NoneModule() fxutil.Module {
	return fxutil.Component(
		fx.Provide(func() optional.Option[collector.Component] {
			return optional.NewNoneOption[collector.Component]()
		}),
	)
}

package telemetry

import "sync"

var (
	once    sync.Once
	global  Component
	builder Builder
)

// TODO: everything in this file should be eventually removed as
//       telemetry is fully componentized and no longer used as
//       as a global.

// SetBuilder call from init hook to define telemetry builder
func SetBuilder(b Builder) {
	builder = b
}

// GetCompatComponent returns a component wrapping telemetry global variables
// TODO (components): Remove this when all telemetry is migrated to the component
func GetCompatComponent() Component {
	once.Do(func() {
		if builder != nil {
			global = builder()
		}
	})

	return global
}

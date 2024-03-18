module github.com/DataDog/datadog-agent/comp/serializer/compression/compressionimpl

go 1.21.8

replace github.com/DataDog/datadog-agent/comp/serializer/compression => ../.

require (
	github.com/DataDog/datadog-agent/comp/core/config v0.51.1
	github.com/DataDog/datadog-agent/comp/serializer/compression v0.0.0-00010101000000-000000000000
	github.com/DataDog/datadog-agent/pkg/util/fxutil v0.51.1
	github.com/DataDog/datadog-agent/pkg/util/log v0.51.1
	github.com/DataDog/zstd v1.5.5
	go.uber.org/fx v1.21.0
)

require (
	github.com/DataDog/datadog-agent/comp/core/secrets v0.51.1 // indirect
	github.com/DataDog/datadog-agent/pkg/collector/check/defaults v0.51.1 // indirect
	github.com/DataDog/datadog-agent/pkg/config/env v0.51.1 // indirect
	github.com/DataDog/datadog-agent/pkg/config/model v0.51.1 // indirect
	github.com/DataDog/datadog-agent/pkg/config/setup v0.51.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/executable v0.51.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/filesystem v0.51.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/hostname/validate v0.51.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/optional v0.51.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/scrubber v0.51.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/system/socket v0.51.1 // indirect
	github.com/DataDog/datadog-agent/pkg/util/winutil v0.51.1 // indirect
	github.com/DataDog/viper v1.12.0 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/shirou/gopsutil/v3 v3.23.9 // indirect
	github.com/spf13/afero v1.1.2 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/dig v1.17.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	golang.org/x/exp v0.0.0-20231214170342-aacd6d4b4611 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/tools v0.16.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

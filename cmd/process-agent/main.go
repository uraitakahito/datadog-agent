// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//nolint:revive // TODO(PROC) Fix revive linter
package main

import (
	_ "net/http/pprof"
	"os"

	"github.com/DataDog/datadog-agent/cmd/internal/runcmd"
	"github.com/DataDog/datadog-agent/cmd/process-agent/command"
	"github.com/DataDog/datadog-agent/cmd/process-agent/subcommands"
	"github.com/DataDog/datadog-agent/pkg/util/flavor"

	_ "github.com/DataDog/datadog-agent/comp/forwarder/orchestrator"
	_ "github.com/DataDog/datadog-agent/internal/third_party/kubernetes/pkg/kubelet/types"
	_ "github.com/DataDog/datadog-agent/pkg/collector/corechecks/cluster/orchestrator/processors"
	_ "github.com/DataDog/datadog-agent/pkg/collector/corechecks/cluster/orchestrator/processors/k8s"
	_ "github.com/DataDog/datadog-agent/pkg/collector/corechecks/cluster/orchestrator/transformers"
	_ "github.com/DataDog/datadog-agent/pkg/collector/corechecks/cluster/orchestrator/transformers/k8s"
	_ "github.com/DataDog/datadog-agent/pkg/obfuscate"
	_ "github.com/DataDog/datadog-agent/pkg/orchestrator"
	_ "github.com/DataDog/datadog-agent/pkg/proto/pbgo/trace"
	_ "github.com/DataDog/datadog-agent/pkg/status/render"
	_ "github.com/DataDog/datadog-agent/pkg/trace/config"
	_ "github.com/DataDog/datadog-agent/pkg/trace/log"
	_ "github.com/DataDog/datadog-agent/pkg/trace/telemetry"
	_ "github.com/DataDog/datadog-agent/pkg/trace/traceutil"
	_ "github.com/DataDog/go-sqllexer"
	_ "github.com/DataDog/opentelemetry-mapping-go/pkg/otlp/attributes"
	_ "github.com/DataDog/opentelemetry-mapping-go/pkg/otlp/attributes/azure"
	_ "github.com/DataDog/opentelemetry-mapping-go/pkg/otlp/attributes/ec2"
	_ "github.com/DataDog/opentelemetry-mapping-go/pkg/otlp/attributes/gcp"
	_ "github.com/DataDog/opentelemetry-mapping-go/pkg/otlp/attributes/source"
	_ "github.com/hashicorp/go-version"
	_ "github.com/knadh/koanf/maps"
	_ "github.com/knadh/koanf/providers/confmap"
	_ "github.com/knadh/koanf/v2"
	_ "github.com/mitchellh/copystructure"
	_ "github.com/mitchellh/reflectwalk"
	_ "github.com/outcaste-io/ristretto"
	_ "github.com/outcaste-io/ristretto/z"
	_ "github.com/outcaste-io/ristretto/z/simd"
	_ "github.com/pmezard/go-difflib/difflib"
	_ "github.com/stretchr/objx"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/mock"
	_ "github.com/ulikunitz/xz"
	_ "go.opentelemetry.io/collector/component"
	_ "go.opentelemetry.io/collector/config/configtelemetry"
	_ "go.opentelemetry.io/collector/confmap"
	_ "go.opentelemetry.io/collector/featuregate"
	_ "go.opentelemetry.io/collector/pdata/pcommon"
	_ "go.opentelemetry.io/collector/semconv/v1.6.1"
	_ "golang.org/x/sys/execabs"
)

// main is the main application entry point
func main() {
	flavor.SetFlavor(flavor.ProcessAgent)

	os.Args = command.FixDeprecatedFlags(os.Args, os.Stdout)

	rootCmd := command.MakeCommand(subcommands.ProcessAgentSubcommands(), command.UseWinParams, command.RootCmdRun)
	os.Exit(runcmd.Run(rootCmd))
}

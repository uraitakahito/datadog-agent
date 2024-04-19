// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build serverless

package config

import (
	corecompcfg "github.com/DataDog/datadog-agent/comp/core/config"
	coreconfig "github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/trace/config"
)

func setOtlpReceiver(c *config.AgentConfig, core corecompcfg.Component, grpcPort int) error {
	c.OTLPReceiver = &config.OTLP{
		BindHost:               c.ReceiverHost,
		GRPCPort:               grpcPort,
		MaxRequestBytes:        c.MaxRequestBytes,
		SpanNameRemappings:     coreconfig.Datadog.GetStringMapString("otlp_config.traces.span_name_remappings"),
		SpanNameAsResourceName: core.GetBool("otlp_config.traces.span_name_as_resource_name"),
		ProbabilisticSampling:  core.GetFloat64("otlp_config.traces.probabilistic_sampler.sampling_percentage"),
	}

	return nil
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver

// Package metrics defines the telemetry of the Admission Controller.
package metrics

import (
	"github.com/DataDog/datadog-agent/pkg/telemetry"

	"github.com/prometheus/client_golang/prometheus"
)

// Metric names
const (
	SecretControllerName   = "secrets"
	WebhooksControllerName = "webhooks"
)

// Mutation errors
const (
	InvalidInput         = "invalid_input"
	InternalError        = "internal_error"
	ConfigInjectionError = "config_injection_error"
)

// Status tags
const (
	StatusSuccess = "success"
	StatusError   = "error"
)

// Telemetry metrics
var (
	ReconcileSuccess = telemetry.NewGaugeWithOpts("admission_webhooks", "reconcile_success",
		[]string{"controller"}, "Number of reconcile success per controller.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	ReconcileErrors = telemetry.NewGaugeWithOpts("admission_webhooks", "reconcile_errors",
		[]string{"controller"}, "Number of reconcile errors per controller.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	CertificateDuration = telemetry.NewGaugeWithOpts("admission_webhooks", "certificate_expiry",
		[]string{}, "Time left before the certificate expires in hours.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	MutationAttempts = telemetry.NewGaugeWithOpts("admission_webhooks", "mutation_attempts",
		[]string{"mutation_type", "status", "injected", "error"}, "Number of pod mutation attempts by mutation type",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	WebhooksReceived = telemetry.NewCounterWithOpts("admission_webhooks", "webhooks_received",
		[]string{"mutation_type"}, "Number of mutation webhook requests received.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	GetOwnerCacheHit = telemetry.NewGaugeWithOpts("admission_webhooks", "owner_cache_hit",
		[]string{"resource"}, "Number of cache hits while getting pod's owner object.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	GetOwnerCacheMiss = telemetry.NewGaugeWithOpts("admission_webhooks", "owner_cache_miss",
		[]string{"resource"}, "Number of cache misses while getting pod's owner object.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	WebhooksResponseDuration = telemetry.NewHistogramWithOpts(
		"admission_webhooks",
		"response_duration",
		[]string{"mutation_type"},
		"Webhook response duration distribution (in seconds).",
		prometheus.DefBuckets, // The default prometheus buckets are adapted to measure response time
		telemetry.Options{NoDoubleUnderscoreSep: true},
	)
	LibInjectionAttempts = telemetry.NewCounterWithOpts("admission_webhooks", "library_injection_attempts",
		[]string{"language", "injected", "auto_detected", "injection_type"}, "Number of pod library injection attempts by language and injection type",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	LibInjectionErrors = telemetry.NewCounterWithOpts("admission_webhooks", "library_injection_errors",
		[]string{"language", "auto_detected", "injection_type"}, "Number of library injection failures by language and injection type",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	CWSExecInstrumentationAttempts = telemetry.NewHistogramWithOpts(
		"admission_webhooks",
		"cws_exec_instrumentation_attempts",
		[]string{"mode", "injected", "reason"},
		"Distribution of exec requests instrumentation attempts by CWS Instrumentation mode",
		prometheus.LinearBuckets(0, 1, 1),
		telemetry.Options{NoDoubleUnderscoreSep: true})
	CWSPodInstrumentationAttempts = telemetry.NewHistogramWithOpts(
		"admission_webhooks",
		"cws_pod_instrumentation_attempts",
		[]string{"mode", "injected", "reason"},
		"Distribution of pod requests instrumentation attempts by CWS Instrumentation mode",
		prometheus.LinearBuckets(0, 1, 1),
		telemetry.Options{NoDoubleUnderscoreSep: true})
	RemoteConfigs = telemetry.NewGaugeWithOpts("admission_webhooks", "rc_provider_configs",
		[]string{}, "Number of valid remote configurations.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	InvalidRemoteConfigs = telemetry.NewGaugeWithOpts("admission_webhooks", "rc_provider_configs_invalid",
		[]string{}, "Number of invalid remote configurations.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	DeleteRemoteConfigsAttempts = telemetry.NewCounterWithOpts("admission_webhooks", "rc_provider_configs_delete_attempts",
		[]string{}, "Number of deleted remote configurations.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	DeleteRemoteConfigsCompleted = telemetry.NewCounterWithOpts("admission_webhooks", "rc_provider_configs_delete_completed",
		[]string{}, "Number of deleted remote configurations.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	DeleteRemoteConfigsErrors = telemetry.NewCounterWithOpts("admission_webhooks", "rc_provider_configs_delete_errors",
		[]string{}, "Number of deleted remote configurations.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	PatchAttempts = telemetry.NewCounterWithOpts("admission_webhooks", "patcher_attempts",
		[]string{}, "Number of patch attempts.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	PatchCompleted = telemetry.NewCounterWithOpts("admission_webhooks", "patcher_completed",
		[]string{}, "Number of completed patch attempts.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
	PatchErrors = telemetry.NewCounterWithOpts("admission_webhooks", "patcher_errors",
		[]string{}, "Number of patch errors.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
)

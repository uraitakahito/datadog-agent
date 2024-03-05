// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

//go:build !serverless
// +build !serverless

// Package leaderelection provides functions related with the leader election
// mechanism offered in Kubernetes.
package telemetry

import (
	"github.com/DataDog/datadog-agent/comp/core/telemetry"
	"github.com/DataDog/datadog-agent/comp/core/telemetry/telemetryimpl"
)

// GetCompatComponent returns a component wrapping telemetry global variables
// TODO (components): Remove this when all telemetry is migrated to the component
func GetCompatComponent() telemetry.Component {
	return telemetryimpl.GetCompatComponent()
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver

package autoinstrumentation

import (
	"errors"

	"github.com/DataDog/datadog-agent/pkg/clusteragent/admission/common"
)

// TargetObjKind represents the supported k8s object kinds
type TargetObjKind string

const (
	KindCluster TargetObjKind = "cluster"
)

// Action is the action requested in the patch
type Action string

const (
	// StageConfig instructs the patcher to process the configuration without triggering a rolling update
	StageConfig Action = "stage"
	// EnableConfig instructs the patcher to apply the patch request
	EnableConfig Action = "enable"
	// DisableConfig instructs the patcher to disable library injection
	DisableConfig Action = "disable"
)

// Request holds the required data to target a k8s object and apply library configuration
type Request struct {
	ID            string `json:"id"`
	Revision      int64  `json:"revision"`
	RcVersion     uint64 `json:"rc_version"`
	SchemaVersion string `json:"schema_version"`
	Action        Action `json:"action"`

	// Library parameters
	LibConfig common.LibConfig `json:"lib_config"`

	K8sTargetV2 *K8sTargetV2 `json:"k8s_target_v2,omitempty"`
}

type K8sClusterTarget struct {
	ClusterName       string    `json:"cluster_name"`
	Enabled           *bool     `json:"enabled,omitempty"`
	EnabledNamespaces *[]string `json:"enabled_namespaces,omitempty"`
}

type K8sTargetV2 struct {
	ClusterTargets []K8sClusterTarget `json:"cluster_targets"`
}

// Validate returns whether a patch request is applicable
func (pr Request) Validate(clusterName string) error {
	if pr.LibConfig.Language == "" {
		return errors.New("library language is empty")
	}
	if pr.LibConfig.Version == "" {
		return errors.New("library version is empty")
	}
	return nil
}

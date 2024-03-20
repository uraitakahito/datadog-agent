// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux

// Package selector holds selector related files
package selector

import (
	proto "github.com/DataDog/agent-payload/v5/cws/dumpsv1"

	cgroupModel "github.com/DataDog/datadog-agent/pkg/security/resolvers/cgroup/model"
)

// WorkloadSelectorToProto incode a WorkloadSelector to its protobuf representation
func WorkloadSelectorToProto(input *cgroupModel.WorkloadSelector) *proto.ProfileSelector {
	if input == nil {
		return nil
	}

	return &proto.ProfileSelector{
		ImageName: input.Image,
		ImageTag:  input.Tag,
	}
}

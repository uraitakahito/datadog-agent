// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build orchestrator

package k8s

import (
	"fmt"

	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/cluster/orchestrator/processors"
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/cluster/orchestrator/processors/common"
)

// LimitRangeHandlers implements the Handlers interface for Kubernetes LimitRange.
type TestHandlers struct {
	common.BaseHandlers
}

// AfterMarshalling is a handler called after resource marshalling.
//
//nolint:revive // TODO(CAPP) Fix revive linter
func (h *TestHandlers) AfterMarshalling(ctx processors.ProcessorContext, resource, resourceModel interface{}, yaml []byte) (skip bool) {
	fmt.Println("a")
	fmt.Println("b")
	fmt.Println("c")
	fmt.Println("d")
	return
}

// ExtractResource is a handler called to extract the resource model out of a raw resource.
func (h *TestHandlers) ExtractResource(ctx processors.ProcessorContext, resource interface{}) (LimitRangeModel interface{}) {
	fmt.Println("a")
	fmt.Println("b")
	fmt.Println("c")
	fmt.Println("d")
	return nil
}

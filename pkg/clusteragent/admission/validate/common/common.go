// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver

// Package common provides functions used by several mutating webhooks
package common

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/wI2L/jsondiff"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"

	"github.com/DataDog/datadog-agent/pkg/clusteragent/admission/metrics"
)

// ValidationFunc is a function that mutates a pod
type ValidationFunc func(pod *corev1.Pod, ns string, cl dynamic.Interface) (bool, error)

// Validate handles validating pods and encoding and decoding admission
// requests and responses for the public validate functions.
func Validate(rawPod []byte, ns string, mutationType string, m ValidationFunc, dc dynamic.Interface) ([]byte, error) {
	var pod corev1.Pod
	if err := json.Unmarshal(rawPod, &pod); err != nil {
		return nil, fmt.Errorf("failed to decode raw object: %v", err)
	}

	validated, err := m(&pod, ns, dc)
	if err != nil {
		metrics.ValidationAttempts.Inc(mutationType, metrics.StatusError, strconv.FormatBool(false), err.Error())
		return nil, fmt.Errorf("failed to mutate pod: %v", err)
	}

	metrics.ValidationAttempts.Inc(mutationType, metrics.StatusSuccess, strconv.FormatBool(validated), "")

	bytes, err := json.Marshal(pod)
	if err != nil {
		return nil, fmt.Errorf("failed to encode the validated Pod object: %v", err)
	}

	patch, err := jsondiff.CompareJSON(rawPod, bytes) // TODO: Try to generate the patch at the MutationFunc
	if err != nil {
		return nil, fmt.Errorf("failed to prepare the JSON patch: %v", err)
	}

	return json.Marshal(patch)
}

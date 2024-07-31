// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

//go:build kubeapiserver

// Package always_admit implements a validation webhook that always admits the request.
// This is useful for testing purposes.
package always_admit

import (
	"github.com/DataDog/datadog-agent/cmd/cluster-agent/admission"
	"github.com/DataDog/datadog-agent/pkg/clusteragent/admission/validate/common"
	"github.com/DataDog/datadog-agent/pkg/clusteragent/autoscaling/workload"
	"github.com/DataDog/datadog-agent/pkg/config"

	admiv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

const (
	webhookName     = "alwaysadmit"
	webhookEndpoint = "/alwaysadmit"
	webhookType     = "validating"
)

// Webhook implements the MutatingWebhook interface
type Webhook struct {
	name        string
	isEnabled   bool
	webhookType string
	endpoint    string
	resources   []string
	operations  []admiv1.OperationType
	patcher     workload.PodPatcher
}

// NewWebhook returns a new Webhook
func NewWebhook() *Webhook {
	return &Webhook{
		name:        webhookName,
		isEnabled:   config.Datadog().GetBool("autoscaling.workload.enabled"),
		webhookType: webhookType,
		endpoint:    webhookEndpoint,
		resources:   []string{"pods"},
		operations:  []admiv1.OperationType{admiv1.Create},
	}
}

// Name returns the name of the webhook
func (w *Webhook) Name() string {
	return w.name
}

// IsEnabled returns whether the webhook is enabled
func (w *Webhook) IsEnabled() bool {
	return w.isEnabled
}

// Endpoint returns the endpoint of the webhook
func (w *Webhook) Endpoint() string {
	return w.endpoint
}

// Resources returns the kubernetes resources for which the webhook should
// be invoked
func (w *Webhook) Resources() []string {
	return w.resources
}

// Operations returns the operations on the resources specified for which
// the webhook should be invoked
func (w *Webhook) Operations() []admiv1.OperationType {
	return w.operations
}

// LabelSelectors returns the label selectors that specify when the webhook
// should be invoked
func (w *Webhook) LabelSelectors(useNamespaceSelector bool) (namespaceSelector *metav1.LabelSelector, objectSelector *metav1.LabelSelector) {
	return common.DefaultLabelSelectors(useNamespaceSelector)
}

// WebhookFunc returns the function that runs the webhook logic
func (w *Webhook) WebhookFunc() admission.WebhookFunc {
	return w.validate
}

// validate is the function that runs the webhook logic
func (w *Webhook) validate(request *admission.Request) ([]byte, error) {
	// TODO Add user information from request *admission.Request
	return common.Validate(request.Raw, request.Namespace, w.Name(), w.alwaysAdmit, request.DynamicClient)
}

// alwaysAdmit is the function that always admits the request
func (w *Webhook) alwaysAdmit(pod *corev1.Pod, _ string, _ dynamic.Interface) (bool, error) {
	return false, nil
}

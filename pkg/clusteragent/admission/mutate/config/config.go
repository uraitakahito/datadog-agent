// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver

// Package config implements the webhook that injects DD_AGENT_HOST and
// DD_ENTITY_ID into a pod template as needed
package config

import (
	"errors"
	"fmt"
	"strings"

	admiv1 "k8s.io/api/admission/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"github.com/DataDog/datadog-agent/cmd/cluster-agent/admission"
	workloadmeta "github.com/DataDog/datadog-agent/comp/core/workloadmeta/def"
	"github.com/DataDog/datadog-agent/pkg/clusteragent/admission/common"
	"github.com/DataDog/datadog-agent/pkg/clusteragent/admission/metrics"
	mutatecommon "github.com/DataDog/datadog-agent/pkg/clusteragent/admission/mutate/common"
	"github.com/DataDog/datadog-agent/pkg/config"
	apiCommon "github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver/common"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

const (
	// Env vars
	agentHostEnvVarName      = "DD_AGENT_HOST"
	ddEntityIDEnvVarName     = "DD_ENTITY_ID"
	ddExternalDataEnvVarName = "DD_EXTERNAL_ENV"
	traceURLEnvVarName       = "DD_TRACE_AGENT_URL"
	dogstatsdURLEnvVarName   = "DD_DOGSTATSD_URL"
	podUIDEnvVarName         = "DD_INTERNAL_POD_UID"

	// External Data Prefixes
	// These prefixes are used to build the External Data Environment Variable.
	// This variable is then used for Origin Detection.
	externalDataInitPrefix          = "it-"
	externalDataContainerNamePrefix = "cn-"
	externalDataPodUIDPrefix        = "pu-"

	// Config injection modes
	hostIP  = "hostip"
	socket  = "socket"
	service = "service"

	// DatadogVolumeName is the name of the volume used to mount the socket
	DatadogVolumeName = "datadog"

	webhookName = "agent_config"
)

var (
	agentHostIPEnvVar = corev1.EnvVar{
		Name:  agentHostEnvVarName,
		Value: "",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.hostIP",
			},
		},
	}

	agentHostServiceEnvVar = corev1.EnvVar{
		Name:  agentHostEnvVarName,
		Value: config.Datadog().GetString("admission_controller.inject_config.local_service_name") + "." + apiCommon.GetMyNamespace() + ".svc.cluster.local",
	}

	defaultDdEntityIDEnvVar = corev1.EnvVar{
		Name:  ddEntityIDEnvVarName,
		Value: "",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.uid",
			},
		},
	}

	traceURLSocketEnvVar = corev1.EnvVar{
		Name:  traceURLEnvVarName,
		Value: config.Datadog().GetString("admission_controller.inject_config.trace_agent_socket"),
	}

	dogstatsdURLSocketEnvVar = corev1.EnvVar{
		Name:  dogstatsdURLEnvVarName,
		Value: config.Datadog().GetString("admission_controller.inject_config.dogstatsd_socket"),
	}
)

// Webhook is the webhook that injects DD_AGENT_HOST and DD_ENTITY_ID into a pod
type Webhook struct {
	name            string
	webhookType     common.WebhookType
	isEnabled       bool
	endpoint        string
	resources       []string
	operations      []admissionregistrationv1.OperationType
	mode            string
	wmeta           workloadmeta.Component
	injectionFilter mutatecommon.InjectionFilter
}

// NewWebhook returns a new Webhook
func NewWebhook(wmeta workloadmeta.Component, injectionFilter mutatecommon.InjectionFilter) *Webhook {
	return &Webhook{
		name:            webhookName,
		webhookType:     common.MutatingWebhook,
		isEnabled:       config.Datadog().GetBool("admission_controller.inject_config.enabled"),
		endpoint:        config.Datadog().GetString("admission_controller.inject_config.endpoint"),
		resources:       []string{"pods"},
		operations:      []admissionregistrationv1.OperationType{admissionregistrationv1.Create},
		mode:            config.Datadog().GetString("admission_controller.inject_config.mode"),
		wmeta:           wmeta,
		injectionFilter: injectionFilter,
	}
}

// Name returns the name of the webhook
func (w *Webhook) Name() string {
	return w.name
}

// WebhookType returns the type of the webhook
func (w *Webhook) WebhookType() common.WebhookType {
	return w.webhookType
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
func (w *Webhook) Operations() []admissionregistrationv1.OperationType {
	return w.operations
}

// LabelSelectors returns the label selectors that specify when the webhook
// should be invoked
func (w *Webhook) LabelSelectors(useNamespaceSelector bool) (namespaceSelector *metav1.LabelSelector, objectSelector *metav1.LabelSelector) {
	return common.DefaultLabelSelectors(useNamespaceSelector)
}

// WebhookFunc returns the function that mutates the resources
func (w *Webhook) WebhookFunc() admission.WebhookFunc {
	return func(request *admission.Request) *admiv1.AdmissionResponse {
		return common.MutationResponse(mutatecommon.Mutate(request.Raw, request.Namespace, w.Name(), w.inject, request.DynamicClient))
	}
}

// inject injects the following environment variables into the pod template:
// - DD_AGENT_HOST: the host IP of the node
// - DD_ENTITY_ID: the entity ID of the pod
// - DD_EXTERNAL_ENV: the External Data Environment Variable
func (w *Webhook) inject(pod *corev1.Pod, _ string, _ dynamic.Interface) (bool, error) {
	var injectedConfig, injectedEntity, injectedExternalEnv bool

	if pod == nil {
		return false, errors.New(metrics.InvalidInput)
	}

	if !w.injectionFilter.ShouldMutatePod(pod) {
		return false, nil
	}

	// Inject DD_AGENT_HOST
	switch injectionMode(pod, w.mode) {
	case hostIP:
		injectedConfig = mutatecommon.InjectEnv(pod, agentHostIPEnvVar)
	case service:
		injectedConfig = mutatecommon.InjectEnv(pod, agentHostServiceEnvVar)
	case socket:
		volume, volumeMount := buildVolume(DatadogVolumeName, config.Datadog().GetString("admission_controller.inject_config.socket_path"), true)
		injectedVol := mutatecommon.InjectVolume(pod, volume, volumeMount)
		injectedEnv := mutatecommon.InjectEnv(pod, traceURLSocketEnvVar)
		injectedEnv = mutatecommon.InjectEnv(pod, dogstatsdURLSocketEnvVar) || injectedEnv
		injectedConfig = injectedEnv || injectedVol
	default:
		log.Errorf("invalid injection mode %q", w.mode)
		return false, errors.New(metrics.InvalidInput)
	}

	injectedEntity = mutatecommon.InjectEnv(pod, defaultDdEntityIDEnvVar)

	// Inject External Data Environment Variable
	injectedExternalEnv = injectExternalDataEnvVar(pod)

	return injectedConfig || injectedEntity || injectedExternalEnv, nil
}

// injectionMode returns the injection mode based on the global mode and pod labels
func injectionMode(pod *corev1.Pod, globalMode string) string {
	if val, found := pod.GetLabels()[common.InjectionModeLabelKey]; found {
		mode := strings.ToLower(val)
		switch mode {
		case hostIP, service, socket:
			return mode
		default:
			log.Warnf("Invalid label value '%s=%s' on pod %s should be either 'hostip', 'service' or 'socket', defaulting to %q", common.InjectionModeLabelKey, val, mutatecommon.PodString(pod), globalMode)
			return globalMode
		}
	}

	return globalMode
}

// buildExternalEnv generate an External Data environment variable.
func buildExternalEnv(container *corev1.Container, init bool) (corev1.EnvVar, error) {
	return corev1.EnvVar{
		Name:  ddExternalDataEnvVarName,
		Value: fmt.Sprintf("%s%t,%s%s,%s$(%s)", externalDataInitPrefix, init, externalDataContainerNamePrefix, container.Name, externalDataPodUIDPrefix, podUIDEnvVarName),
	}, nil
}

// injectExternalDataEnvVar injects the External Data environment variable.
// The format is: it-<init>,cn-<container_name>,pu-<pod_uid>
func injectExternalDataEnvVar(pod *corev1.Pod) (injected bool) {
	// Inject External Data Environment Variable for the pod
	injected = mutatecommon.InjectDynamicEnv(pod, buildExternalEnv)

	// Inject Internal Pod UID
	injected = mutatecommon.InjectEnv(pod, corev1.EnvVar{
		Name: podUIDEnvVarName,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.uid",
			},
		},
	}) || injected

	return
}

func buildVolume(volumeName, path string, readOnly bool) (corev1.Volume, corev1.VolumeMount) {
	pathType := corev1.HostPathDirectoryOrCreate
	volume := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: path,
				Type: &pathType,
			},
		},
	}

	volumeMount := corev1.VolumeMount{
		Name:      volumeName,
		MountPath: path,
		ReadOnly:  readOnly,
	}

	return volume, volumeMount
}

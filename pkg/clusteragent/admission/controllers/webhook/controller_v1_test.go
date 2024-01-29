// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver

package webhook

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admiv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/DataDog/datadog-agent/pkg/clusteragent/admission/mutate"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/certificate"
)

const waitFor = 5 * time.Second
const tick = 50 * time.Millisecond

var v1Cfg = NewConfig(true, false)

func TestSecretNotFoundV1(t *testing.T) {
	f := newFixtureV1(t)

	stopCh := make(chan struct{})
	defer close(stopCh)
	c := f.run(stopCh)

	_, err := c.webhooksLister.Get(v1Cfg.getWebhookName())
	if !errors.IsNotFound(err) {
		t.Fatal("Webhook shouldn't be created")
	}

	// The queue might not be closed yet because it's done asynchronously
	assert.Eventually(t, func() bool {
		return c.queue.Len() == 0
	}, waitFor, tick, "Work queue isn't empty")
}

func TestCreateWebhookV1(t *testing.T) {
	f := newFixtureV1(t)

	data, err := certificate.GenerateSecretData(time.Now(), time.Now().Add(365*24*time.Hour), []string{"my.svc.dns"})
	if err != nil {
		t.Fatalf("Failed to create the Secret: %v", err)
	}

	secret := buildSecret(data, v1Cfg)
	f.populateSecretsCache(secret)

	stopCh := make(chan struct{})
	defer close(stopCh)
	c := f.run(stopCh)

	var webhook *admiv1.MutatingWebhookConfiguration
	require.Eventually(t, func() bool {
		webhook, err = c.webhooksLister.Get(v1Cfg.getWebhookName())
		return err == nil
	}, waitFor, tick)

	if err := validateV1(webhook, secret); err != nil {
		t.Fatalf("Invalid Webhook: %v", err)
	}

	if c.queue.Len() != 0 {
		t.Fatal("Work queue isn't empty")
	}
}

func TestUpdateOutdatedWebhookV1(t *testing.T) {
	f := newFixtureV1(t)

	data, err := certificate.GenerateSecretData(time.Now(), time.Now().Add(365*24*time.Hour), []string{"my.svc.dns"})
	if err != nil {
		t.Fatalf("Failed to create the Secret: %v", err)
	}

	secret := buildSecret(data, v1Cfg)
	f.populateSecretsCache(secret)

	webhook := &admiv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: v1Cfg.getWebhookName(),
		},
		Webhooks: []admiv1.MutatingWebhook{
			{
				Name: "webhook-foo",
			},
			{
				Name: "webhook-bar",
			},
		},
	}

	f.populateWebhooksCache(webhook)

	stopCh := make(chan struct{})
	defer close(stopCh)
	c := f.run(stopCh)

	var newWebhook *admiv1.MutatingWebhookConfiguration
	require.Eventually(t, func() bool {
		newWebhook, err = c.webhooksLister.Get(v1Cfg.getWebhookName())
		return err == nil && !reflect.DeepEqual(webhook, newWebhook)
	}, waitFor, tick)

	if err := validateV1(newWebhook, secret); err != nil {
		t.Fatalf("Invalid Webhook: %v", err)
	}

	if c.queue.Len() != 0 {
		t.Fatal("Work queue isn't empty")
	}
}

func TestAdmissionControllerFailureModeIgnore(t *testing.T) {
	f := newFixtureV1(t)
	c, _ := f.createController()
	c.config = NewConfig(true, false)

	holdValue := config.Datadog.Get("admission_controller.failure_policy")
	defer config.Datadog.SetWithoutSource("admission_controller.failure_policy", holdValue)

	config.Datadog.SetWithoutSource("admission_controller.failure_policy", "Ignore")
	c.config = NewConfig(true, false)

	webhookSkeleton := c.getWebhookSkeleton("foo", "/bar", []admiv1.OperationType{admiv1.Create}, []string{"pods"}, nil, nil)
	assert.Equal(t, admiv1.Ignore, *webhookSkeleton.FailurePolicy)

	config.Datadog.SetWithoutSource("admission_controller.failure_policy", "ignore")
	c.config = NewConfig(true, false)

	webhookSkeleton = c.getWebhookSkeleton("foo", "/bar", []admiv1.OperationType{admiv1.Create}, []string{"pods"}, nil, nil)
	assert.Equal(t, admiv1.Ignore, *webhookSkeleton.FailurePolicy)

	config.Datadog.SetWithoutSource("admission_controller.failure_policy", "BadVal")
	c.config = NewConfig(true, false)

	webhookSkeleton = c.getWebhookSkeleton("foo", "/bar", []admiv1.OperationType{admiv1.Create}, []string{"pods"}, nil, nil)
	assert.Equal(t, admiv1.Ignore, *webhookSkeleton.FailurePolicy)

	config.Datadog.SetWithoutSource("admission_controller.failure_policy", "")
	c.config = NewConfig(true, false)

	webhookSkeleton = c.getWebhookSkeleton("foo", "/bar", []admiv1.OperationType{admiv1.Create}, []string{"pods"}, nil, nil)
	assert.Equal(t, admiv1.Ignore, *webhookSkeleton.FailurePolicy)
}

func TestAdmissionControllerFailureModeFail(t *testing.T) {
	holdValue := config.Datadog.Get("admission_controller.failure_policy")
	defer config.Datadog.SetWithoutSource("admission_controller.failure_policy", holdValue)

	f := newFixtureV1(t)
	c, _ := f.createController()

	config.Datadog.SetWithoutSource("admission_controller.failure_policy", "Fail")
	c.config = NewConfig(true, false)

	webhookSkeleton := c.getWebhookSkeleton("foo", "/bar", []admiv1.OperationType{admiv1.Create}, []string{"pods"}, nil, nil)
	assert.Equal(t, admiv1.Fail, *webhookSkeleton.FailurePolicy)

	config.Datadog.SetWithoutSource("admission_controller.failure_policy", "fail")
	c.config = NewConfig(true, false)

	webhookSkeleton = c.getWebhookSkeleton("foo", "/bar", []admiv1.OperationType{admiv1.Create}, []string{"pods"}, nil, nil)
	assert.Equal(t, admiv1.Fail, *webhookSkeleton.FailurePolicy)
}

func TestGenerateTemplatesV1(t *testing.T) {
	mockConfig := config.Mock(t)
	defaultReinvocationPolicy := admiv1.IfNeededReinvocationPolicy
	failurePolicy := admiv1.Ignore
	matchPolicy := admiv1.Exact
	sideEffects := admiv1.SideEffectClassNone
	port := int32(443)
	timeout := config.Datadog.GetInt32("admission_controller.timeout_seconds")
	webhook := func(name, path string, objSelector, nsSelector *metav1.LabelSelector, operations []admiv1.OperationType, resources []string) admiv1.MutatingWebhook {
		return admiv1.MutatingWebhook{
			Name: name,
			ClientConfig: admiv1.WebhookClientConfig{
				Service: &admiv1.ServiceReference{
					Namespace: "nsfoo",
					Name:      "datadog-admission-controller",
					Port:      &port,
					Path:      &path,
				},
			},
			Rules: []admiv1.RuleWithOperations{
				{
					Operations: operations,
					Rule: admiv1.Rule{
						APIGroups:   []string{""},
						APIVersions: []string{"v1"},
						Resources:   resources,
					},
				},
			},
			ReinvocationPolicy:      &defaultReinvocationPolicy,
			FailurePolicy:           &failurePolicy,
			MatchPolicy:             &matchPolicy,
			SideEffects:             &sideEffects,
			TimeoutSeconds:          &timeout,
			AdmissionReviewVersions: []string{"v1", "v1beta1"},
			ObjectSelector:          objSelector,
			NamespaceSelector:       nsSelector,
		}
	}
	tests := []struct {
		name        string
		setupConfig func()
		configFunc  func() Config
		want        func() []admiv1.MutatingWebhook
	}{
		{
			name: "config injection, mutate all",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.mutate_unlabelled", true)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.auto_instrumentation.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", false)
			},
			configFunc: func() Config { return NewConfig(false, false) },
			want: func() []admiv1.MutatingWebhook {
				webhook := webhook("datadog.webhook.config", "/injectconfig", &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "admission.datadoghq.com/enabled",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"false"},
						},
					},
				}, nil, []admiv1.OperationType{admiv1.Create}, []string{"pods"})
				return []admiv1.MutatingWebhook{webhook}
			},
		},
		{
			name: "config injection, mutate labelled",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.mutate_unlabelled", false)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.auto_instrumentation.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", false)
			},
			configFunc: func() Config { return NewConfig(false, false) },
			want: func() []admiv1.MutatingWebhook {
				webhook := webhook("datadog.webhook.config", "/injectconfig", &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"admission.datadoghq.com/enabled": "true",
					},
				}, nil, []admiv1.OperationType{admiv1.Create}, []string{"pods"})
				return []admiv1.MutatingWebhook{webhook}
			},
		},
		{
			name: "tags injection, mutate all",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.mutate_unlabelled", true)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.auto_instrumentation.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", false)
			},
			configFunc: func() Config { return NewConfig(false, false) },
			want: func() []admiv1.MutatingWebhook {
				webhook := webhook("datadog.webhook.tags", "/injecttags", &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "admission.datadoghq.com/enabled",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"false"},
						},
					},
				}, nil, []admiv1.OperationType{admiv1.Create}, []string{"pods"})
				return []admiv1.MutatingWebhook{webhook}
			},
		},
		{
			name: "tags injection, mutate labelled",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.mutate_unlabelled", false)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.auto_instrumentation.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", false)
			},
			configFunc: func() Config { return NewConfig(false, false) },
			want: func() []admiv1.MutatingWebhook {
				webhook := webhook("datadog.webhook.tags", "/injecttags", &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"admission.datadoghq.com/enabled": "true",
					},
				}, nil, []admiv1.OperationType{admiv1.Create}, []string{"pods"})
				return []admiv1.MutatingWebhook{webhook}
			},
		},
		{
			name: "lib injection, mutate all",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.mutate_unlabelled", true)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.auto_instrumentation.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", false)
			},
			configFunc: func() Config { return NewConfig(false, false) },
			want: func() []admiv1.MutatingWebhook {
				webhook := webhook("datadog.webhook.auto.instrumentation", "/injectlib", &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "admission.datadoghq.com/enabled",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"false"},
						},
					},
				}, nil, []admiv1.OperationType{admiv1.Create}, []string{"pods"})
				return []admiv1.MutatingWebhook{webhook}
			},
		},
		{
			name: "lib injection, mutate labelled",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.mutate_unlabelled", false)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.auto_instrumentation.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", false)
			},
			configFunc: func() Config { return NewConfig(false, false) },
			want: func() []admiv1.MutatingWebhook {
				webhook := webhook("datadog.webhook.auto.instrumentation", "/injectlib", &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"admission.datadoghq.com/enabled": "true",
					},
				}, nil, []admiv1.OperationType{admiv1.Create}, []string{"pods"})
				return []admiv1.MutatingWebhook{webhook}
			},
		},
		{
			name: "config and tags injection, mutate labelled",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.auto_instrumentation.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", false)
			},
			configFunc: func() Config { return NewConfig(false, false) },
			want: func() []admiv1.MutatingWebhook {
				webhookConfig := webhook("datadog.webhook.config", "/injectconfig", &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"admission.datadoghq.com/enabled": "true",
					},
				}, nil, []admiv1.OperationType{admiv1.Create}, []string{"pods"})
				webhookTags := webhook("datadog.webhook.tags", "/injecttags", &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"admission.datadoghq.com/enabled": "true",
					},
				}, nil, []admiv1.OperationType{admiv1.Create}, []string{"pods"})
				return []admiv1.MutatingWebhook{webhookConfig, webhookTags}
			},
		},
		{
			name: "config and tags injection, mutate all",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.mutate_unlabelled", true)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.auto_instrumentation.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", false)
			},
			configFunc: func() Config { return NewConfig(false, false) },
			want: func() []admiv1.MutatingWebhook {
				webhookConfig := webhook("datadog.webhook.config", "/injectconfig", &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "admission.datadoghq.com/enabled",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"false"},
						},
					},
				}, nil, []admiv1.OperationType{admiv1.Create}, []string{"pods"})
				webhookTags := webhook("datadog.webhook.tags", "/injecttags", &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "admission.datadoghq.com/enabled",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"false"},
						},
					},
				}, nil, []admiv1.OperationType{admiv1.Create}, []string{"pods"})
				return []admiv1.MutatingWebhook{webhookConfig, webhookTags}
			},
		},
		{
			name: "namespace selector enabled",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.mutate_unlabelled", false)
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.auto_instrumentation.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.namespace_selector_fallback", true)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", false)
			},
			configFunc: func() Config { return NewConfig(false, true) },
			want: func() []admiv1.MutatingWebhook {
				webhookConfig := webhook("datadog.webhook.config", "/injectconfig", nil, &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"admission.datadoghq.com/enabled": "true",
					},
				}, []admiv1.OperationType{admiv1.Create}, []string{"pods"})
				webhookTags := webhook("datadog.webhook.tags", "/injecttags", nil, &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"admission.datadoghq.com/enabled": "true",
					},
				}, []admiv1.OperationType{admiv1.Create}, []string{"pods"})
				return []admiv1.MutatingWebhook{webhookConfig, webhookTags}
			},
		},
		{
			name: "AKS-specific label selector without namespace selector enabled",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.add_aks_selectors", true)
				mockConfig.SetWithoutSource("admission_controller.namespace_selector_fallback", false)
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.mutate_unlabelled", true)
				mockConfig.SetWithoutSource("admission_controller.auto_instrumentation.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", false)
			},
			configFunc: func() Config { return NewConfig(false, false) },
			want: func() []admiv1.MutatingWebhook {
				webhook := webhook(
					"datadog.webhook.config",
					"/injectconfig",
					&metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "admission.datadoghq.com/enabled",
								Operator: metav1.LabelSelectorOpNotIn,
								Values:   []string{"false"},
							},
						},
					},
					&metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "control-plane",
								Operator: metav1.LabelSelectorOpDoesNotExist,
							},
							{
								Key:      "control-plane",
								Operator: metav1.LabelSelectorOpNotIn,
								Values:   []string{"true"},
							},
							{
								Key:      "kubernetes.azure.com/managedby",
								Operator: metav1.LabelSelectorOpNotIn,
								Values:   []string{"aks"},
							},
						},
					},
					[]admiv1.OperationType{admiv1.Create},
					[]string{"pods"},
				)
				return []admiv1.MutatingWebhook{webhook}
			},
		},
		{
			name: "AKS-specific label selector with namespace selector enabled",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.add_aks_selectors", true)
				mockConfig.SetWithoutSource("admission_controller.namespace_selector_fallback", true)
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.mutate_unlabelled", true)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", false)
			},
			configFunc: func() Config { return NewConfig(false, true) },
			want: func() []admiv1.MutatingWebhook {
				webhook := webhook(
					"datadog.webhook.config",
					"/injectconfig",
					nil,
					&metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "admission.datadoghq.com/enabled",
								Operator: metav1.LabelSelectorOpNotIn,
								Values:   []string{"false"},
							},
							{
								Key:      "control-plane",
								Operator: metav1.LabelSelectorOpDoesNotExist,
							},
							{
								Key:      "control-plane",
								Operator: metav1.LabelSelectorOpNotIn,
								Values:   []string{"true"},
							},
							{
								Key:      "kubernetes.azure.com/managedby",
								Operator: metav1.LabelSelectorOpNotIn,
								Values:   []string{"aks"},
							},
						},
					},
					[]admiv1.OperationType{admiv1.Create},
					[]string{"pods"},
				)
				return []admiv1.MutatingWebhook{webhook}
			},
		},
		{
			name: "cws instrumentation",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.mutate_unlabelled", false)
				mockConfig.SetWithoutSource("admission_controller.namespace_selector_fallback", false)
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.auto_instrumentation.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.mutate_unlabelled", true)
			},
			configFunc: func() Config { return NewConfig(true, false) },
			want: func() []admiv1.MutatingWebhook {
				podWebhook := webhook(
					"datadog.webhook.cws.pod.instrumentation",
					"/inject-pod-cws",
					&metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      mutate.CWSInstrumentationPodLabelEnabled,
								Operator: metav1.LabelSelectorOpNotIn,
								Values:   []string{"false"},
							},
						},
					},
					nil,
					[]admiv1.OperationType{admiv1.Create},
					[]string{"pods"},
				)
				execWebhook := webhook(
					"datadog.webhook.cws.command.instrumentation",
					"/inject-command-cws",
					nil,
					nil,
					[]admiv1.OperationType{admiv1.Connect},
					[]string{"pods/exec"},
				)
				return []admiv1.MutatingWebhook{podWebhook, execWebhook}
			},
		},
		{
			name: "cws instrumentation, mutate unlabelled",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.mutate_unlabelled", false)
				mockConfig.SetWithoutSource("admission_controller.namespace_selector_fallback", false)
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.auto_instrumentation.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.mutate_unlabelled", false)
			},
			configFunc: func() Config { return NewConfig(true, false) },
			want: func() []admiv1.MutatingWebhook {
				podWebhook := webhook(
					"datadog.webhook.cws.pod.instrumentation",
					"/inject-pod-cws",
					&metav1.LabelSelector{
						MatchLabels: map[string]string{
							mutate.CWSInstrumentationPodLabelEnabled: "true",
						},
					},
					nil,
					[]admiv1.OperationType{admiv1.Create},
					[]string{"pods"},
				)
				execWebhook := webhook(
					"datadog.webhook.cws.command.instrumentation",
					"/inject-command-cws",
					nil,
					nil,
					[]admiv1.OperationType{admiv1.Connect},
					[]string{"pods/exec"},
				)
				return []admiv1.MutatingWebhook{podWebhook, execWebhook}
			},
		},
		{
			name: "cws instrumentation, namespace selector",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.mutate_unlabelled", false)
				mockConfig.SetWithoutSource("admission_controller.namespace_selector_fallback", true)
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.auto_instrumentation.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.mutate_unlabelled", true)
			},
			configFunc: func() Config { return NewConfig(true, true) },
			want: func() []admiv1.MutatingWebhook {
				podWebhook := webhook(
					"datadog.webhook.cws.pod.instrumentation",
					"/inject-pod-cws",
					nil,
					&metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      mutate.CWSInstrumentationPodLabelEnabled,
								Operator: metav1.LabelSelectorOpNotIn,
								Values:   []string{"false"},
							},
						},
					},
					[]admiv1.OperationType{admiv1.Create},
					[]string{"pods"},
				)
				execWebhook := webhook(
					"datadog.webhook.cws.command.instrumentation",
					"/inject-command-cws",
					nil,
					nil,
					[]admiv1.OperationType{admiv1.Connect},
					[]string{"pods/exec"},
				)
				return []admiv1.MutatingWebhook{podWebhook, execWebhook}
			},
		},
		{
			name: "cws instrumentation, namespace selector, mutate unlabelled",
			setupConfig: func() {
				mockConfig.SetWithoutSource("admission_controller.mutate_unlabelled", false)
				mockConfig.SetWithoutSource("admission_controller.namespace_selector_fallback", true)
				mockConfig.SetWithoutSource("admission_controller.inject_config.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.inject_tags.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.auto_instrumentation.enabled", false)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.enabled", true)
				mockConfig.SetWithoutSource("admission_controller.cws_instrumentation.mutate_unlabelled", false)
			},
			configFunc: func() Config { return NewConfig(true, true) },
			want: func() []admiv1.MutatingWebhook {
				podWebhook := webhook(
					"datadog.webhook.cws.pod.instrumentation",
					"/inject-pod-cws",
					nil,
					&metav1.LabelSelector{
						MatchLabels: map[string]string{
							mutate.CWSInstrumentationPodLabelEnabled: "true",
						},
					},
					[]admiv1.OperationType{admiv1.Create},
					[]string{"pods"},
				)
				execWebhook := webhook(
					"datadog.webhook.cws.command.instrumentation",
					"/inject-command-cws",
					nil,
					nil,
					[]admiv1.OperationType{admiv1.Connect},
					[]string{"pods/exec"},
				)
				return []admiv1.MutatingWebhook{podWebhook, execWebhook}
			},
		},
	}

	mockConfig.SetWithoutSource("kube_resources_namespace", "nsfoo")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupConfig()
			defer resetMockConfig(mockConfig) // Reset to default

			c := &ControllerV1{}
			c.config = tt.configFunc()
			c.generateTemplates()

			assert.EqualValues(t, tt.want(), c.webhookTemplates)
		})
	}
}

func TestGetWebhookSkeletonV1(t *testing.T) {
	mockConfig := config.Mock(t)
	defaultReinvocationPolicy := admiv1.IfNeededReinvocationPolicy
	failurePolicy := admiv1.Ignore
	matchPolicy := admiv1.Exact
	sideEffects := admiv1.SideEffectClassNone
	port := int32(443)
	path := "/bar"
	defaultTimeout := config.Datadog.GetInt32("admission_controller.timeout_seconds")
	customTimeout := int32(2)
	namespaceSelector, _ := buildLabelSelectors(true)
	_, objectSelector := buildLabelSelectors(false)
	webhook := func(to *int32, objSelector, nsSelector *metav1.LabelSelector) admiv1.MutatingWebhook {
		return admiv1.MutatingWebhook{
			Name: "datadog.webhook.foo",
			ClientConfig: admiv1.WebhookClientConfig{
				Service: &admiv1.ServiceReference{
					Namespace: "nsfoo",
					Name:      "datadog-admission-controller",
					Port:      &port,
					Path:      &path,
				},
			},
			Rules: []admiv1.RuleWithOperations{
				{
					Operations: []admiv1.OperationType{
						admiv1.Create,
					},
					Rule: admiv1.Rule{
						APIGroups:   []string{""},
						APIVersions: []string{"v1"},
						Resources:   []string{"pods"},
					},
				},
			},
			ReinvocationPolicy:      &defaultReinvocationPolicy,
			FailurePolicy:           &failurePolicy,
			MatchPolicy:             &matchPolicy,
			SideEffects:             &sideEffects,
			TimeoutSeconds:          to,
			AdmissionReviewVersions: []string{"v1", "v1beta1"},
			ObjectSelector:          objSelector,
			NamespaceSelector:       nsSelector,
		}
	}
	type args struct {
		nameSuffix string
		path       string
	}
	tests := []struct {
		name              string
		args              args
		timeout           *int32
		namespaceSelector bool
		want              admiv1.MutatingWebhook
	}{
		{
			name: "nominal case",
			args: args{
				nameSuffix: "foo",
				path:       "/bar",
			},
			namespaceSelector: false,
			want:              webhook(&defaultTimeout, objectSelector, nil),
		},
		{
			name: "namespace selector",
			args: args{
				nameSuffix: "foo",
				path:       "/bar",
			},
			namespaceSelector: true,
			want:              webhook(&defaultTimeout, nil, namespaceSelector),
		},
		{
			name: "custom timeout",
			args: args{
				nameSuffix: "foo",
				path:       "/bar",
			},
			timeout:           &customTimeout,
			namespaceSelector: false,
			want:              webhook(&customTimeout, objectSelector, nil),
		},
	}

	mockConfig.SetWithoutSource("kube_resources_namespace", "nsfoo")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeout != nil {
				mockConfig.SetWithoutSource("admission_controller.timeout_seconds", *tt.timeout)
				defer mockConfig.SetDefault("admission_controller.timeout_seconds", defaultTimeout)
			}

			c := &ControllerV1{}
			c.config = NewConfig(false, tt.namespaceSelector)

			nsSelector, objSelector := buildLabelSelectors(tt.namespaceSelector)

			assert.EqualValues(t, tt.want, c.getWebhookSkeleton(tt.args.nameSuffix, tt.args.path, []admiv1.OperationType{admiv1.Create}, []string{"pods"}, nsSelector, objSelector))
		})
	}
}

type fixtureV1 struct {
	fixture
}

func newFixtureV1(t *testing.T) *fixtureV1 {
	f := &fixtureV1{}
	f.t = t
	f.client = fake.NewSimpleClientset()
	return f
}

func (f *fixtureV1) createController() (*ControllerV1, informers.SharedInformerFactory) {
	factory := informers.NewSharedInformerFactory(f.client, time.Duration(0))

	return NewControllerV1(
		f.client,
		factory.Core().V1().Secrets(),
		factory.Admissionregistration().V1().MutatingWebhookConfigurations(),
		func() bool { return true },
		make(chan struct{}),
		v1Cfg,
	), factory
}

func (f *fixtureV1) run(stopCh chan struct{}) *ControllerV1 {
	c, factory := f.createController()

	factory.Start(stopCh)
	go c.Run(stopCh)

	return c
}

func (f *fixtureV1) populateWebhooksCache(webhooks ...*admiv1.MutatingWebhookConfiguration) {
	for _, w := range webhooks {
		_, _ = f.client.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.TODO(), w, metav1.CreateOptions{})
	}
}

func validateV1(w *admiv1.MutatingWebhookConfiguration, s *corev1.Secret) error {
	if len(w.Webhooks) != 3 {
		return fmt.Errorf("Webhooks should contain 3 entries, got %d", len(w.Webhooks))
	}

	for i := 0; i < len(w.Webhooks); i++ {
		if !reflect.DeepEqual(w.Webhooks[i].ClientConfig.CABundle, certificate.GetCABundle(s.Data)) {
			return fmt.Errorf("The Webhook CABundle doesn't match the Secret: CABundle: %v, Secret: %v", w.Webhooks[i].ClientConfig.CABundle, s)
		}
	}

	return nil
}

func TestAdmissionControllerReinvocationPolicyV1(t *testing.T) {
	f := newFixtureV1(t)
	c, _ := f.createController()
	c.config = NewConfig(true, false)

	defaultValue := config.Datadog.Get("admission_controller.reinvocation_policy")
	defer config.Datadog.SetWithoutSource("admission_controller.reinvocation_policy", defaultValue)

	config.Datadog.SetWithoutSource("admission_controller.reinvocation_policy", "IfNeeded")
	c.config = NewConfig(true, false)
	webhook := c.getWebhookSkeleton("foo", "/bar", []admiv1.OperationType{admiv1.Create}, []string{"pods"}, nil, nil)
	assert.Equal(t, admiv1.IfNeededReinvocationPolicy, *webhook.ReinvocationPolicy)

	config.Datadog.SetWithoutSource("admission_controller.reinvocation_policy", "ifneeded")
	c.config = NewConfig(true, false)
	webhook = c.getWebhookSkeleton("foo", "/bar", []admiv1.OperationType{admiv1.Create}, []string{"pods"}, nil, nil)
	assert.Equal(t, admiv1.IfNeededReinvocationPolicy, *webhook.ReinvocationPolicy)

	config.Datadog.SetWithoutSource("admission_controller.reinvocation_policy", "Never")
	c.config = NewConfig(true, false)
	webhook = c.getWebhookSkeleton("foo", "/bar", []admiv1.OperationType{admiv1.Create}, []string{"pods"}, nil, nil)
	assert.Equal(t, admiv1.NeverReinvocationPolicy, *webhook.ReinvocationPolicy)

	config.Datadog.SetWithoutSource("admission_controller.reinvocation_policy", "never")
	c.config = NewConfig(true, false)
	webhook = c.getWebhookSkeleton("foo", "/bar", []admiv1.OperationType{admiv1.Create}, []string{"pods"}, nil, nil)
	assert.Equal(t, admiv1.NeverReinvocationPolicy, *webhook.ReinvocationPolicy)

	config.Datadog.SetWithoutSource("admission_controller.reinvocation_policy", "wrong")
	c.config = NewConfig(true, false)
	webhook = c.getWebhookSkeleton("foo", "/bar", []admiv1.OperationType{admiv1.Create}, []string{"pods"}, nil, nil)
	assert.Equal(t, admiv1.IfNeededReinvocationPolicy, *webhook.ReinvocationPolicy)

	config.Datadog.SetWithoutSource("admission_controller.reinvocation_policy", "")
	c.config = NewConfig(true, false)
	webhook = c.getWebhookSkeleton("foo", "/bar", []admiv1.OperationType{admiv1.Create}, []string{"pods"}, nil, nil)
	assert.Equal(t, admiv1.IfNeededReinvocationPolicy, *webhook.ReinvocationPolicy)
}

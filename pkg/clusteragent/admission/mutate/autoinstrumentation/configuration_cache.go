// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver

// Package autoinstrumentation implements the webhook that injects APM libraries
// into pods
package autoinstrumentation

import (
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

type instrumentationConfiguration struct {
	enabled            bool
	enabledNamespaces  []string
	disabledNamespaces []string
}

type instrumentationConfigurationCache struct {
	localConfiguration        *instrumentationConfiguration
	currentConfiguration      *instrumentationConfiguration
	configurationUpdatesQueue chan Request
	clusterName               string
}

//var c = newInstrumentationConfigurationCache()

func newInstrumentationConfigurationCache(
	provider rcProvider,
	localEnabled *bool,
	localEnabledNamespaces *[]string,
	localDisabledNamespaces *[]string,
	clusterName string,
) *instrumentationConfigurationCache {
	localConfig := newInstrumentationConfiguration(localEnabled, localEnabledNamespaces, localDisabledNamespaces)
	currentConfig := newInstrumentationConfiguration(localEnabled, localEnabledNamespaces, localDisabledNamespaces)
	reqChannel := make(chan Request, 10)
	if provider != nil {
		reqChannel = provider.subscribe("cluster")
	}

	return &instrumentationConfigurationCache{
		localConfiguration:        localConfig,
		currentConfiguration:      currentConfig,
		configurationUpdatesQueue: reqChannel,
		clusterName:               clusterName,
	}
}

func (c *instrumentationConfigurationCache) start(stopCh <-chan struct{}) {
	for {
		select {
		case req := <-c.configurationUpdatesQueue:
			// if err := c.updateConfiguration(nil, nil); err != nil {
			// 	log.Error(err.Error())
			// }
			c.update(req)
		case <-stopCh:
			log.Info("Shutting down patcher")
			return
		}
	}
}

func (c *instrumentationConfigurationCache) update(req Request) {
	if req.K8sTargetV2 == nil || req.K8sTargetV2.ClusterTargets == nil {
		log.Errorf("K8sTargetV2 is not set for config %s", req.ID)
	}
	k8sClusterTargets := req.K8sTargetV2.ClusterTargets
	//env := req.K8sTargetV2.Environment

	for _, target := range k8sClusterTargets {
		clusterName := target.ClusterName
		log.Debugf("APM Configuration cache: clusterName %s", clusterName)
		if c.clusterName == clusterName {
			newEnabled := target.Enabled
			newEnabledNamespaces := target.EnabledNamespaces
			c.updateConfiguration(*newEnabled, newEnabledNamespaces)
		}
	}
}

func (c *instrumentationConfigurationCache) readConfiguration() *instrumentationConfiguration {
	return c.currentConfiguration
}

func (c *instrumentationConfigurationCache) readLocalConfiguration() *instrumentationConfiguration {
	return c.localConfiguration
}

func (c *instrumentationConfigurationCache) updateConfiguration(enabled bool, enabledNamespaces *[]string) {
	log.Debugf("Updating current APM Instrumentation configuration")
	log.Debugf("Old APM Instrumentation configuration [enabled=%t enabledNamespaces=%v disabledNamespaces=%v]",
		c.currentConfiguration.enabled,
		c.currentConfiguration.enabledNamespaces,
		c.currentConfiguration.disabledNamespaces,
	)

	if c.currentConfiguration.enabled && !enabled {
		log.Errorf("Disabling APM instrumentation in the cluster remotly is not supported")
		return
	}

	if c.currentConfiguration.enabled {
		if len(c.currentConfiguration.enabledNamespaces) == 0 &&
			len(c.currentConfiguration.disabledNamespaces) == 0 &&
			enabledNamespaces != nil && len(*enabledNamespaces) > 0 {
			log.Errorf("Currently APM Insrtumentation is enabled in the whole cluster. Cannot enable it in specific namespaces.")
			return
		} else if len(c.currentConfiguration.enabledNamespaces) > 0 {
			log.Debugf("Appending new enabledNamespaces to the configuration...")
			// TODO: deduplicate enabledNamespaces and remove enabledNamespcaes from disabledNamespaces list
			for _, ns := range *enabledNamespaces {
				c.currentConfiguration.enabledNamespaces = append(c.currentConfiguration.enabledNamespaces, ns)
			}
		} else if len(c.currentConfiguration.disabledNamespaces) > 0 {
			log.Debugf("Removing new namespaces to enable from disabledNamespaces...")
			disabledNsMap := make(map[string]struct{})
			for _, ns := range c.currentConfiguration.disabledNamespaces {
				disabledNsMap[ns] = struct{}{}
			}
			for _, ns := range *enabledNamespaces {
				if _, ok := disabledNsMap[ns]; ok {
					delete(disabledNsMap, ns)
				}
			}
			disabledNs := make([]string, 0, len(disabledNsMap))
			for k := range disabledNsMap {
				disabledNs = append(disabledNs, k)
			}
			c.currentConfiguration.disabledNamespaces = disabledNs
		}
	} else {
		if enabled {
			c.currentConfiguration.enabled = enabled
			if enabledNamespaces != nil && len(*enabledNamespaces) > 0 {
				log.Debugf("Enabling APM instrumentation in namespaces [%v] ...", *enabledNamespaces)
				c.currentConfiguration.enabledNamespaces = *enabledNamespaces
			} else {
				log.Debugf("Enabling APM instrumentation in the whole cluster...")
			}
		} else {
			log.Errorf("Noop: APM Instrumentation is off")
			return
		}
	}

	log.Debugf("New APM Instrumentation configuration [enabled=%t enabledNamespaces=%v disabledNamespaces=%v]",
		c.currentConfiguration.enabled,
		c.currentConfiguration.enabledNamespaces,
		c.currentConfiguration.disabledNamespaces,
	)

}

func newInstrumentationConfiguration(
	enabled *bool,
	enabledNamespaces *[]string,
	disabledNamespaces *[]string,
) *instrumentationConfiguration {
	config := instrumentationConfiguration{
		enabled:            false,
		enabledNamespaces:  []string{},
		disabledNamespaces: []string{},
	}
	if enabled != nil {
		config.enabled = *enabled
	}
	if enabledNamespaces != nil {
		config.enabledNamespaces = *enabledNamespaces
	}
	if disabledNamespaces != nil {
		config.disabledNamespaces = *disabledNamespaces
	}

	return &config
}

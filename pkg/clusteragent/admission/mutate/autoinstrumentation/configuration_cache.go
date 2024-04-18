// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver

// Package autoinstrumentation implements the webhook that injects APM libraries
// into pods
package autoinstrumentation

import (
	"fmt"
	"sync"

	"github.com/DataDog/datadog-agent/pkg/remoteconfig/state"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

type instrumentationConfiguration struct {
	enabled            bool
	enabledNamespaces  []string
	disabledNamespaces []string
}

type enablementConfig struct {
	configID          string
	rcVersion         int
	rcAction          string
	env               *string
	enabled           *bool
	enabledNamespaces *[]string
}

type instrumentationConfigurationCache struct {
	localConfiguration           *instrumentationConfiguration
	currentConfiguration         *instrumentationConfiguration
	configurationUpdatesQueue    chan Request
	configurationUpdatesResponse chan Response
	clusterName                  string
	namespaceToConfigIdMap       map[string]string // maps the namespace with enabled instrumentation to Remote Enablement rule

	mu                  sync.RWMutex
	lastAppliedRevision int64
	orderedRevisions    []int64
	enabledConfigIDs    map[string]interface{}
	enabledRevisions    map[int64]enablementConfig
	deletedConfigIDs    map[string]interface{}
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
	respChannel := make(chan Response, 10)
	if provider != nil {
		reqChannel = provider.subscribe("cluster")
		respChannel = provider.getResponseChan()
	}
	nsToRules := make(map[string]string)
	if *localEnabled {
		for _, ns := range *localEnabledNamespaces {
			nsToRules[ns] = "local"
		}
	}

	return &instrumentationConfigurationCache{
		localConfiguration:           localConfig,
		currentConfiguration:         currentConfig,
		configurationUpdatesQueue:    reqChannel,
		configurationUpdatesResponse: respChannel,
		clusterName:                  clusterName,
		namespaceToConfigIdMap:       nsToRules,

		orderedRevisions: make([]int64, 0),
		enabledConfigIDs: map[string]interface{}{},
		enabledRevisions: map[int64]enablementConfig{},
		deletedConfigIDs: map[string]interface{}{},
	}
}

func (c *instrumentationConfigurationCache) start(stopCh <-chan struct{}) {
	for {
		select {
		case req := <-c.configurationUpdatesQueue:
			// if err := c.updateConfiguration(nil, nil); err != nil {
			// 	log.Error(err.Error())
			// }
			resp := c.update(req)
			c.configurationUpdatesResponse <- resp
		case <-stopCh:
			log.Info("Shutting down patcher")
			return
		}
	}
}

func (c *instrumentationConfigurationCache) update(req Request) Response {
	// if req.K8sTargetV2 == nil || req.K8sTargetV2.ClusterTargets == nil {
	// 	log.Errorf("K8sTargetV2 is not set for config %s", req.ID)
	// }
	k8sClusterTargets := req.K8sTargetV2.ClusterTargets
	var resp Response

	switch req.Action {
	case StageConfig:
		// Consume the config without triggering a rolling update.
		log.Debugf("Remote Config ID %q with revision %q has a \"stage\" action. The pod template won't be patched, only the deployment annotations", req.ID, req.Revision)
	case EnableConfig:
		for _, target := range k8sClusterTargets {
			clusterName := target.ClusterName
			log.Debugf("APM Configuration cache: clusterName %s", clusterName)
			log.Info("LILIYAB11")
			if c.clusterName == clusterName {
				log.Infof("Current configuration: %v, %v, %v",
					c.currentConfiguration.enabled, c.currentConfiguration.enabledNamespaces, c.currentConfiguration.disabledNamespaces)
				newEnabled := target.Enabled
				newEnabledNamespaces := target.EnabledNamespaces

				c.mu.Lock()
				resp = c.updateConfiguration(*newEnabled, newEnabledNamespaces, req.ID, int(req.RcVersion))
				log.Infof("Updated configuration: %v, %v, %v",
					c.currentConfiguration.enabled, c.currentConfiguration.enabledNamespaces, c.currentConfiguration.disabledNamespaces)

				c.orderedRevisions = append(c.orderedRevisions, req.Revision)
				c.enabledConfigIDs[req.ID] = struct{}{}
				c.enabledRevisions[req.Revision] = enablementConfig{
					configID:          req.ID,
					rcVersion:         int(req.RcVersion),
					rcAction:          string(req.Action),
					env:               req.LibConfig.Env,
					enabled:           target.Enabled,
					enabledNamespaces: target.EnabledNamespaces,
				}
				c.mu.Unlock()
			}
		}
	case DisableConfig:
		log.Info("LILIYAB00")
		log.Infof("ID: %s", req.ID)
		log.Infof("Revision: %v", req.Revision)
		log.Infof("RcVersion: %v", req.RcVersion)
		log.Infof("Env: %v", req.LibConfig.Env)
		log.Infof("K8sTargetV2.ClusterTargets: %v", req.K8sTargetV2.ClusterTargets)
		for _, target := range k8sClusterTargets {
			if c.clusterName == target.ClusterName {
				log.Infof("Current configuration: %v, %v, %v",
					c.currentConfiguration.enabled, c.currentConfiguration.enabledNamespaces, c.currentConfiguration.disabledNamespaces)
				//newEnabled := target.Enabled
				//newEnabledNamespaces := target.EnabledNamespaces
				resp = c.disableConfiguration(req.ID, req.Revision, req.RcVersion)
				//c.updateConfiguration(*newEnabled, newEnabledNamespaces, req.ID)
				log.Infof("Updated configuration: %v, %v, %v",
					c.currentConfiguration.enabled, c.currentConfiguration.enabledNamespaces, c.currentConfiguration.disabledNamespaces)

			}
		}
	default:
		log.Errorf("unknown action %q", req.Action)
	}

	return resp
}

func (c *instrumentationConfigurationCache) readConfiguration() *instrumentationConfiguration {
	return c.currentConfiguration
}

func (c *instrumentationConfigurationCache) readLocalConfiguration() *instrumentationConfiguration {
	return c.localConfiguration
}

func (c *instrumentationConfigurationCache) disableConfiguration(
	configID string, rcRevision int64, rcVersion uint64) Response {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.enabledConfigIDs, configID)
	// for r, conf := range c.enabledRevisions {
	// 	if conf.configID == configID {
	// 		delete(c.enabledRevisions, r)
	// 		break
	// 	}
	// }

	for i, rev := range c.orderedRevisions {
		confId, ok := c.enabledRevisions[rev]
		if !ok {
			log.Error("Revision was not found")
		}
		if confId.configID == configID {
			delete(c.enabledRevisions, rev)
			c.orderedRevisions = append(c.orderedRevisions[:i], c.orderedRevisions[i+1:]...)
			break
		}
	}
	c.resetConfiguration()
	return Response{
		ID:        configID,
		Revision:  rcRevision,
		RcVersion: rcVersion,
		Status:    state.ApplyStatus{State: state.ApplyStateAcknowledged},
	}
}

func (c *instrumentationConfigurationCache) resetConfiguration() {
	c.currentConfiguration = c.localConfiguration
	for _, rev := range c.orderedRevisions {
		conf := c.enabledRevisions[rev]
		c.updateConfiguration(*conf.enabled, conf.enabledNamespaces, conf.configID, conf.rcVersion)
	}
}

func (c *instrumentationConfigurationCache) updateConfiguration(enabled bool, enabledNamespaces *[]string, rcID string, rcVersion int) Response {
	log.Debugf("Updating current APM Instrumentation configuration")
	log.Debugf("Old APM Instrumentation configuration [enabled=%t enabledNamespaces=%v disabledNamespaces=%v]",
		c.currentConfiguration.enabled,
		c.currentConfiguration.enabledNamespaces,
		c.currentConfiguration.disabledNamespaces,
	)
	resp := Response{
		ID:        rcID,
		RcVersion: uint64(rcVersion),
		Status:    state.ApplyStatus{State: state.ApplyStateAcknowledged},
	}

	if c.currentConfiguration.enabled && !enabled {
		log.Errorf("Disabling APM instrumentation in the cluster remotly is not supported")
		resp.Status.State = state.ApplyStateError
		resp.Status.Error = "Disabling APM instrumentation in the cluster remotly is not supported"
		return resp
	}

	if c.currentConfiguration.enabled {
		if len(c.currentConfiguration.enabledNamespaces) == 0 &&
			len(c.currentConfiguration.disabledNamespaces) == 0 &&
			enabledNamespaces != nil && len(*enabledNamespaces) > 0 {
			log.Errorf("Currently APM Insrtumentation is enabled in the whole cluster. Cannot enable it in specific namespaces.")
			resp.Status.State = state.ApplyStateError
			resp.Status.Error = "Failing policy because SSI is enabled in the whole cluster"
			return resp
		} else if len(c.currentConfiguration.enabledNamespaces) > 0 {
			log.Debugf("Appending new enabledNamespaces to the configuration...")
			// TODO: deduplicate enabledNamespaces and remove enabledNamespcaes from disabledNamespaces list
			alreadyEnabledNamespaces := []string{}
			for _, ns := range *enabledNamespaces {
				if _, ok := c.namespaceToConfigIdMap[ns]; ok {
					alreadyEnabledNamespaces = append(alreadyEnabledNamespaces, ns)
				} else {
					c.currentConfiguration.enabledNamespaces = append(c.currentConfiguration.enabledNamespaces, ns)
					c.namespaceToConfigIdMap[ns] = fmt.Sprintf("%s-%d", rcID, rcVersion)
				}
			}
			if len(alreadyEnabledNamespaces) > 0 {
				resp.Status.State = state.ApplyStateError
				resp.Status.Error = fmt.Sprintf("Failing policy because SSI in namespaces %v is already enabled", alreadyEnabledNamespaces)
				return resp
			}
		} else if len(c.currentConfiguration.enabledNamespaces) > 0 && (enabledNamespaces == nil || len(*enabledNamespaces) == 0) {
			// TODO: implement scenario local config enabling a single ns, remote config enabling whole cluster
		} else if len(c.currentConfiguration.disabledNamespaces) > 0 {
			log.Debugf("Removing new namespaces to enable from disabledNamespaces...")
			disabledNsMap := make(map[string]struct{})
			for _, ns := range c.currentConfiguration.disabledNamespaces {
				disabledNsMap[ns] = struct{}{}
			}
			for _, ns := range *enabledNamespaces {
				if _, ok := disabledNsMap[ns]; ok {
					delete(disabledNsMap, ns)
					c.namespaceToConfigIdMap[ns] = fmt.Sprintf("%s-%d", rcID, rcVersion)
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
				for _, ns := range *enabledNamespaces {
					c.currentConfiguration.enabledNamespaces = append(c.currentConfiguration.enabledNamespaces, ns)
					c.namespaceToConfigIdMap[ns] = fmt.Sprintf("%s-%d", rcID, rcVersion)
				}
			} else {
				log.Debugf("Enabling APM instrumentation in the whole cluster...")
				c.namespaceToConfigIdMap["cluster"] = fmt.Sprintf("%s-%d", rcID, rcVersion)
			}
		} else {
			log.Errorf("Noop: APM Instrumentation is off")
			resp.Status.State = state.ApplyStateError
			resp.Status.Error = "Noop: APM Instrumentation is off"
			return resp
		}
	}

	log.Debugf("New APM Instrumentation configuration [enabled=%t enabledNamespaces=%v disabledNamespaces=%v]",
		c.currentConfiguration.enabled,
		c.currentConfiguration.enabledNamespaces,
		c.currentConfiguration.disabledNamespaces,
	)
	return resp
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

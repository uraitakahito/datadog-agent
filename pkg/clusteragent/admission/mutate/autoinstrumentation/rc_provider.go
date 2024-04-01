// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver

package autoinstrumentation

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/DataDog/datadog-agent/pkg/clusteragent/telemetry"
	rcclient "github.com/DataDog/datadog-agent/pkg/config/remote/client"
	"github.com/DataDog/datadog-agent/pkg/remoteconfig/state"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// remoteConfigProvider consumes tracing configs from RC and delivers them to the patcher
type remoteConfigProvider struct {
	client                    *rcclient.Client
	isLeaderNotif             <-chan struct{}
	subscribers               map[TargetObjKind]chan Request
	clusterName               string
	telemetryCollector        telemetry.TelemetryCollector
	apmInstrumentationState   *instrumentationConfigurationCache
	lastProcessedRCRevision   int64
	currentlyAppliedConfigIDs map[string]interface{}
	deletedConfigIDs          map[string]interface{}
}

type rcProvider interface {
	start(stopCh <-chan struct{})
	subscribe(kind TargetObjKind) chan Request
}

var _ rcProvider = &remoteConfigProvider{}

func newRemoteConfigProvider(
	client *rcclient.Client,
	isLeaderNotif <-chan struct{},
	telemetryCollector telemetry.TelemetryCollector,
	clusterName string,
) (*remoteConfigProvider, error) {
	if client == nil {
		return nil, errors.New("remote config client not initialized")
	}
	return &remoteConfigProvider{
		client:                    client,
		isLeaderNotif:             isLeaderNotif,
		subscribers:               make(map[TargetObjKind]chan Request),
		clusterName:               clusterName,
		telemetryCollector:        telemetryCollector,
		lastProcessedRCRevision:   0,
		currentlyAppliedConfigIDs: make(map[string]interface{}),
		deletedConfigIDs:          make(map[string]interface{}),
	}, nil
}

func (rcp *remoteConfigProvider) start(stopCh <-chan struct{}) {
	log.Info("Starting remote-config provider")
	rcp.client.Subscribe(state.ProductAPMTracing, rcp.process)
	rcp.client.Start()

	for {
		select {
		case <-rcp.isLeaderNotif:
			log.Info("Got a leader notification, polling from remote-config")
			rcp.process(rcp.client.GetConfigs(state.ProductAPMTracing), rcp.client.UpdateApplyStatus)
		case <-stopCh:
			log.Info("Shutting down remote-config patch provider")
			rcp.client.Close()
			return
		}
	}
}

func (rcp *remoteConfigProvider) subscribe(kind TargetObjKind) chan Request {
	ch := make(chan Request, 10)
	rcp.subscribers[kind] = ch
	return ch
}

// process is the event handler called by the RC client on config updates
func (rcp *remoteConfigProvider) process(update map[string]state.RawConfig, applyStateCallback func(string, state.ApplyStatus)) {
	log.Infof("Got %d updates from remote-config", len(update))
	var valid, invalid float64
	for path, config := range update {
		log.Debugf("Parsing config %s from path %s", config.Config, path)
		var req Request
		err := json.Unmarshal(config.Config, &req)
		if err != nil {
			invalid++
			//rcp.telemetryCollector.SendRemoteConfigPatchEvent(req.getApmRemoteConfigEvent(err, telemetry.ConfigParseFailure))
			log.Errorf("Error while parsing config: %v", err)
			continue
		}

		if shouldSkipConfig(req, rcp.lastProcessedRCRevision, rcp.clusterName) {
			log.Info("LILIYAB: skipping config")
			continue
		}

		req.RcVersion = config.Metadata.Version
		log.Debugf("Patch request parsed %+v", req)
		if ch, found := rcp.subscribers["cluster"]; found {
			valid++
			ch <- req
			applyStateCallback(path, state.ApplyStatus{State: state.ApplyStateAcknowledged})
			rcp.lastProcessedRCRevision = req.Revision
			rcp.currentlyAppliedConfigIDs[req.ID] = struct{}{}
		}
	}
	//metrics.RemoteConfigs.Set(valid)
	//metrics.InvalidRemoteConfigs.Set(invalid)
}

func shouldSkipConfig(req Request, lastAppliedRevision int64, clusterName string) bool {
	// check if config should be applied based on presence K8sTargetV2 object
	if req.K8sTargetV2 == nil || len(req.K8sTargetV2.ClusterTargets) == 0 {
		log.Infof("Skipping config %s because K8sTargetV2 is not set", req.ID)
		return true
	}

	// check if config should be applied based on RC revision
	lastAppliedTime := time.UnixMilli(lastAppliedRevision)
	requestTime := time.UnixMilli(req.Revision)

	if requestTime.Before(lastAppliedTime) || requestTime.Equal(lastAppliedTime) {
		log.Infof("Skipping config %s because it has already been applied: revision %v, last applied revision %v", req.ID, requestTime, lastAppliedTime)
		return true
	}

	isTargetingCluster := false
	for _, target := range req.K8sTargetV2.ClusterTargets {
		if target.ClusterName == clusterName {
			isTargetingCluster = true
			break
		}
	}
	if !isTargetingCluster {
		log.Infof("Skipping config %s because it's not targeting current cluster %s", req.ID, req.K8sTargetV2.ClusterTargets[0].ClusterName)
	}
	return !isTargetingCluster

}

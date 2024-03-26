// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver

package autoinstrumentation

import (
	"encoding/json"
	"errors"

	"github.com/DataDog/datadog-agent/pkg/clusteragent/admission/metrics"
	"github.com/DataDog/datadog-agent/pkg/clusteragent/telemetry"
	rcclient "github.com/DataDog/datadog-agent/pkg/config/remote/client"
	"github.com/DataDog/datadog-agent/pkg/remoteconfig/state"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// remoteConfigProvider consumes tracing configs from RC and delivers them to the patcher
type remoteConfigProvider struct {
	client                  *rcclient.Client
	isLeaderNotif           <-chan struct{}
	subscribers             map[TargetObjKind]chan Request
	clusterName             string
	telemetryCollector      telemetry.TelemetryCollector
	apmInstrumentationState *instrumentationConfigurationCache
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
		client:             client,
		isLeaderNotif:      isLeaderNotif,
		subscribers:        make(map[TargetObjKind]chan Request),
		clusterName:        clusterName,
		telemetryCollector: telemetryCollector,
	}, nil
}

func (rcp *remoteConfigProvider) start(stopCh <-chan struct{}) {
	log.Info("Starting remote-config patch provider")
	log.Debugf("4444444444444444444")
	rcp.client.Subscribe(state.ProductAPMTracing, rcp.process)
	rcp.client.Start()

	for {
		select {
		case <-rcp.isLeaderNotif:
			log.Info("Got a leader notification, polling from remote-config")
			log.Debugf("55555555555555555555555555")
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
		if req.K8sTargetV2 == nil || len(req.K8sTargetV2.ClusterTargets) == 0 {
			log.Infof("Ignoring update with configId %s, because K8sTargetV2 is not set", req.ID)
			continue
		}
		hasUpdateForCluster := false
		for _, target := range req.K8sTargetV2.ClusterTargets {
			if target.ClusterName == rcp.clusterName {
				hasUpdateForCluster = true
				break
			}
		}
		if !hasUpdateForCluster {
			log.Infof("Ignoring update with configId %s, because K8sTargetV2 doesn't have current cluster as a target", req.ID)
			continue
		}
		req.RcVersion = config.Metadata.Version
		log.Debugf("Patch request parsed %+v", req)
		if ch, found := rcp.subscribers["cluster"]; found {
			valid++
			ch <- req
			applyStateCallback(path, state.ApplyStatus{State: state.ApplyStateAcknowledged})
		}
	}
	metrics.RemoteConfigs.Set(valid)
	metrics.InvalidRemoteConfigs.Set(invalid)
}

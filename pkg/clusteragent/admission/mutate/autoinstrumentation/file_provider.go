// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver

package autoinstrumentation

import (
	"encoding/json"
	"os"
	"time"

	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// fileProvider this is a stub and will be used for e2e testing only
type fileProvider struct {
	file                  string
	isLeaderNotif         <-chan struct{}
	pollInterval          time.Duration
	subscribers           map[TargetObjKind]chan Request
	responseChan          chan Response
	lastSuccessfulRefresh time.Time
	clusterName           string
	// *instrumentationConfigurationCache
}

var _ rcProvider = &fileProvider{}

func newfileProvider(file string, isLeaderNotif <-chan struct{}, clusterName string) *fileProvider {
	return &fileProvider{
		file:          file,
		isLeaderNotif: isLeaderNotif,
		pollInterval:  5 * time.Second,
		subscribers:   make(map[TargetObjKind]chan Request),
		responseChan:  make(chan Response, 10),
		clusterName:   clusterName,
	}
}

func (fpp *fileProvider) subscribe(kind TargetObjKind) chan Request {
	ch := make(chan Request, 10)
	fpp.subscribers[kind] = ch

	return ch
}

func (fpp *fileProvider) getResponseChan() chan Response {
	return fpp.responseChan
}

func (fpp *fileProvider) start(stopCh <-chan struct{}) {
	log.Infof("Starting file patch provider: watching %s", fpp.file)
	ticker := time.NewTicker(fpp.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-fpp.isLeaderNotif:
			log.Info("Got a leader notification, polling from file")
			fpp.process(true)
		case <-ticker.C:
			fpp.process(false)
		case <-stopCh:
			log.Info("Shutting down file patch provider")
			return
		}
	}
}

func (fpp *fileProvider) process(forcePoll bool) {
	requests, err := fpp.poll(forcePoll)
	if err != nil {
		log.Errorf("Error refreshing patch requests: %v", err)
		return
	}
	if len(requests) == 0 {
		return
	}
	log.Infof("Got %d updates from local file", len(requests))
	for _, req := range requests {
		if err := req.Validate(fpp.clusterName); err != nil {
			log.Errorf("Skipping invalid patch request: %s", err)
			continue
		}
		if ch, found := fpp.subscribers["cluster"]; found {
			log.Infof("Publishing patch request for target %s", req.K8sTargetV2)
			ch <- req
		}
	}
	fpp.lastSuccessfulRefresh = time.Now()
}

func (fpp *fileProvider) poll(forcePoll bool) ([]Request, error) {
	info, err := os.Stat(fpp.file)
	if err != nil {
		return nil, err
	}
	if !forcePoll && fpp.lastSuccessfulRefresh.After(info.ModTime()) {
		log.Debugf("File %q hasn't changed since the last Successful refresh at %v", fpp.file, fpp.lastSuccessfulRefresh)
		return []Request{}, nil
	}
	content, err := os.ReadFile(fpp.file)
	if err != nil {
		return nil, err
	}
	var requests []Request
	err = json.Unmarshal(content, &requests)
	return requests, err
}

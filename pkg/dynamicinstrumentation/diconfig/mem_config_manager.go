// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux_bpf

package diconfig

import (
	"encoding/json"
	"io"
	"reflect"
	"sync"

	"github.com/DataDog/datadog-agent/pkg/dynamicinstrumentation/ditypes"
	"github.com/DataDog/datadog-agent/pkg/dynamicinstrumentation/proctracker"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// ReaderConfigManager is used to track updates to configurations
// which are read from memory
type ReaderConfigManager struct {
	sync.Mutex
	configReader *ConfigReader
	procTracker  *proctracker.ProcessTracker

	callback configUpdateCallback
	configs  configsByService
	state    ditypes.DIProcs
}

type readerConfigCallback func(configsByService)

func NewReaderConfigManager(reader *ConfigReader) (*ReaderConfigManager, error) {
	cm := &ReaderConfigManager{
		callback:     applyConfigUpdate,
		configReader: reader,
	}

	cm.procTracker = proctracker.NewProcessTracker(cm.updateProcessInfo)
	err := cm.procTracker.Start()
	if err != nil {
		return nil, err
	}

	err = reader.Start()
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func (cm *ReaderConfigManager) update() error {
	var updatedState = ditypes.NewDIProcs()
	for serviceName, configsByID := range cm.configs {
		for pid, proc := range cm.configReader.Processes {
			// If a config exists relevant to this proc
			if proc.ServiceName == serviceName {
				procCopy := *proc
				updatedState[pid] = &procCopy
				updatedState[pid].ProbesByID = convert(serviceName, configsByID)
			}
		}
	}

	if !reflect.DeepEqual(cm.state, updatedState) {
		err := inspectGoBinaries(updatedState)
		if err != nil {
			return err
		}

		for pid, procInfo := range cm.state {
			// cleanup dead procs
			if _, running := updatedState[pid]; !running {
				procInfo.CloseAllUprobeLinks()
				delete(cm.state, pid)
			}
		}

		for pid, procInfo := range updatedState {
			if _, tracked := cm.state[pid]; !tracked {
				for _, probe := range procInfo.GetProbes() {
					// install all probes from new process
					cm.callback(procInfo, probe)
				}
			} else {
				for _, existingProbe := range cm.state[pid].GetProbes() {
					updatedProbe := procInfo.GetProbe(existingProbe.ID)
					if updatedProbe == nil {
						// delete old probes
						cm.state[pid].DeleteProbe(existingProbe.ID)
					}
				}
				for _, updatedProbe := range procInfo.GetProbes() {
					existingProbe := cm.state[pid].GetProbe(updatedProbe.ID)
					if !reflect.DeepEqual(existingProbe, updatedProbe) {
						// update existing probes that changed
						cm.callback(procInfo, updatedProbe)
					}
				}
			}
		}
		cm.state = updatedState
	}
	return nil
}

func (cm *ReaderConfigManager) updateProcessInfo(procs ditypes.DIProcs) {
	cm.Lock()
	defer cm.Unlock()
	log.Info("Updating procs", procs)
	cm.configReader.UpdateProcesses(procs)
	err := cm.update()
	if err != nil {
		log.Info(err)
	}
}

func (cm *ReaderConfigManager) updateServiceConfigs(configs configsByService) {
	log.Info("Updating config from file:", configs)
	cm.configs = configs
	err := cm.update()
	if err != nil {
		log.Info(err)
	}
}

type ConfigReader struct {
	io.Reader
	updateChannel  chan ([]byte)
	Processes      map[ditypes.PID]*ditypes.ProcessInfo
	configCallback configReaderCallback
	stopChannel    chan (bool)
}

type configReaderCallback func(configsByService)

func NewConfigReader(onConfigUpdate configReaderCallback) *ConfigReader {
	return &ConfigReader{
		updateChannel:  make(chan []byte),
		configCallback: onConfigUpdate,
	}
}

func (r *ConfigReader) Read(p []byte) (n int, e error) {
	go func() {
		r.updateChannel <- p
	}()
	return 0, nil
}

func (r *ConfigReader) Start() error {
	go func() {
	configUpdateLoop:
		for {
			select {
			case rawConfigBytes := <-r.updateChannel:
				conf := map[string]map[string]rcConfig{}
				err := json.Unmarshal(rawConfigBytes, &conf)
				if err != nil {
					log.Errorf("invalid config read from reader: %v", err)
					continue
				}
				r.configCallback(conf)
			case <-r.stopChannel:
				break configUpdateLoop
			}
		}
	}()
	return nil
}

func (cu *ConfigReader) Stop() {
	cu.stopChannel <- true
}

// UpdateProcesses is the callback interface that ConfigReader uses to consume the map of ProcessInfo's
// such that it's used whenever there's an update to the state of known service processes on the machine.
// It simply overwrites the previous state of known service processes with the new one
func (cu *ConfigReader) UpdateProcesses(procs ditypes.DIProcs) {
	current := procs
	old := cu.Processes
	if !reflect.DeepEqual(current, old) {
		cu.Processes = current
	}
}

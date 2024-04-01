// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux_bpf

package usm

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/DataDog/datadog-agent/pkg/network/config"
	"github.com/DataDog/datadog-agent/pkg/network/usm/utils"
	"github.com/DataDog/datadog-agent/pkg/process/monitor"
	"github.com/DataDog/datadog-agent/pkg/util/kernel"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	manager "github.com/DataDog/ebpf-manager"
)

const (
	nodejsSslReadRetprobe = "nodejs_uretprobe__SSL_read"
	//nodejsSslReadExRetprobe  = "nodejs_uretprobe__SSL_read_ex"
	nodejsSslWriteRetprobe = "nodejs_uretprobe__SSL_write"
	//nodejsSslWriteExRetprobe = "nodejs_uretprobe__SSL_write_ex"
)

var nodejsProbes = []manager.ProbesSelector{
	&manager.AllOf{
		Selectors: []manager.ProbesSelector{
			&manager.ProbeSelector{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: ssDoHandshakeProbe,
				},
			},
			&manager.ProbeSelector{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: ssDoHandshakeRetprobe,
				},
			},
			&manager.ProbeSelector{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: sslSetBioProbe,
				},
			},
			&manager.ProbeSelector{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: sslReadProbe,
				},
			},
			&manager.ProbeSelector{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: nodejsSslReadRetprobe,
				},
			},
			//&manager.ProbeSelector{
			//	ProbeIdentificationPair: manager.ProbeIdentificationPair{
			//		EBPFFuncName: nodejsSslReadExRetprobe,
			//	},
			//},
			&manager.ProbeSelector{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: sslWriteProbe,
				},
			},
			&manager.ProbeSelector{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: nodejsSslWriteRetprobe,
				},
			},
			//&manager.ProbeSelector{
			//	ProbeIdentificationPair: manager.ProbeIdentificationPair{
			//		EBPFFuncName: nodejsSslWriteExRetprobe,
			//	},
			//},
			&manager.ProbeSelector{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: sslShutdownProbe,
				},
			},
		},
	},
}

var nodeJSCmd = []byte("/bin/node")

type nodejsMonitor struct {
	registry *utils.FileRegistry
	procRoot string

	// `utils.FileRegistry` callbacks
	registerCB   func(utils.FilePath) error
	unregisterCB func(utils.FilePath) error

	// Termination
	wg   sync.WaitGroup
	done chan struct{}
}

func newNodeJSMonitor(c *config.Config, mgr *manager.Manager) *nodejsMonitor {
	//if !c.EnableNodeMonitoring {
	//	return nil
	//}

	procRoot := kernel.ProcFSRoot()
	return &nodejsMonitor{
		registry: utils.NewFileRegistry("nodejs"),
		procRoot: procRoot,
		done:     make(chan struct{}),

		// Callbacks
		registerCB:   addHooks(mgr, procRoot, nodejsProbes),
		unregisterCB: removeHooks(mgr, nodejsProbes),
	}
}

func (m *nodejsMonitor) Start() {
	if m == nil {
		return
	}

	processMonitor := monitor.GetProcessMonitor()

	// Subscribe to process events
	doneExec := processMonitor.SubscribeExec(m.handleProcessExec)
	doneExit := processMonitor.SubscribeExit(m.handleProcessExit)

	// Attach to existing processes
	m.sync()

	m.wg.Add(1)
	go func() {
		// This ticker is responsible for controlling the rate at which
		// we scrape the whole procFS again in order to ensure that we
		// terminate any dangling uprobes and register new processes
		// missed by the process monitor stream
		processSync := time.NewTicker(scanTerminatedProcessesInterval)

		defer func() {
			processSync.Stop()
			// Execute process monitor callback termination functions
			doneExec()
			doneExit()
			// Stopping the process monitor (if we're the last instance)
			processMonitor.Stop()
			// Cleaning up all active hooks
			m.registry.Clear()
			// marking we're finished.
			m.wg.Done()
		}()

		for {
			select {
			case <-m.done:
				return
			case <-processSync.C:
				m.sync()
				m.registry.Log()
			}
		}
	}()

	log.Debug("NodeJS monitoring enabled")
}

func (m *nodejsMonitor) Stop() {
	if m == nil {
		return
	}

	close(m.done)
	m.wg.Wait()
}

// sync state of nodejsMonitor with the current state of procFS
// the purpose of this method is two-fold:
// 1) register processes for which we missed exec events (targeted mostly at startup)
// 2) unregister processes for which we missed exit events
func (m *nodejsMonitor) sync() {
	deletionCandidates := m.registry.GetRegisteredProcesses()

	_ = kernel.WithAllProcs(m.procRoot, func(pid int) error {
		if _, ok := deletionCandidates[uint32(pid)]; ok {
			// We have previously hooked into this process and it remains active,
			// so we remove it from the deletionCandidates list, and move on to the next PID
			delete(deletionCandidates, uint32(pid))
			return nil
		}

		// This is a new PID so we attempt to attach SSL probes to it
		m.handleProcessExec(uint32(pid))
		return nil
	})

	// At this point all entries from deletionCandidates are no longer alive, so
	// we should dettach our SSL probes from them
	for pid := range deletionCandidates {
		m.handleProcessExit(pid)
	}
}

func (m *nodejsMonitor) handleProcessExec(pid uint32) {
	path := m.getNodeJS(pid)
	if path == "" {
		return
	}

	m.registry.Register(
		path,
		pid,
		m.registerCB,
		m.unregisterCB,
	)
}

func (m *nodejsMonitor) handleProcessExit(pid uint32) {
	// We avoid filtering PIDs here because it's cheaper to simply do a registry lookup
	// instead of fetching a process name in order to determine whether it is an
	// envoy process or not (which at the very minimum involves syscalls)
	m.registry.Unregister(pid)
}

func (m *nodejsMonitor) getNodeJS(pid uint32) string {
	cmdlinePath := fmt.Sprintf("%s/%d/cmdline", m.procRoot, pid)

	f, err := os.Open(cmdlinePath)
	if err != nil {
		// This can happen often in the context of ephemeral processes
		return ""
	}
	defer f.Close()

	// From here on we shouldn't allocate for the common case
	// (eg., a process is *not* envoy)
	bufferPtr := readBufferPool.Get().(*[]byte)
	defer func() {
		readBufferPool.Put(bufferPtr)
	}()

	buffer := *bufferPtr
	n, _ := f.Read(buffer)
	if n == 0 {
		return ""
	}

	buffer = buffer[:n]
	i := bytes.Index(buffer, nodeJSCmd)
	if i < 0 {
		return ""
	}

	executable := buffer[:i+len(nodeJSCmd)]
	return string(executable)
}

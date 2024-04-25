// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux_bpf

package usm

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/atomic"
	"io"
	"net/http"
	"syscall"
	"time"

	"github.com/cilium/ebpf"

	manager "github.com/DataDog/ebpf-manager"

	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/ebpf/probe/ebpfcheck"
	"github.com/DataDog/datadog-agent/pkg/network/config"
	filterpkg "github.com/DataDog/datadog-agent/pkg/network/filter"
	"github.com/DataDog/datadog-agent/pkg/network/protocols"
	"github.com/DataDog/datadog-agent/pkg/network/protocols/telemetry"
	"github.com/DataDog/datadog-agent/pkg/process/monitor"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

type monitorState = string

const (
	disabled   monitorState = "disabled"
	running    monitorState = "running"
	notRunning monitorState = "Not running"
)

var (
	state        = disabled
	startupError error
)

// Monitor is responsible for:
// * Creating a raw socket and attaching an eBPF filter to it;
// * Consuming HTTP transaction "events" that are sent from Kernel space;
// * Aggregating and emitting metrics based on the received HTTP transactions;
type Monitor struct {
	cfg *config.Config

	ebpfProgram *ebpfProgram

	processMonitor *monitor.ProcessMonitor

	// termination
	closeFilterFn func()

	lastUpdateTime *atomic.Int64
}

var (
	mon *Monitor
)

type Request struct {
	Protocol string `json:"protocol"`
}

func StartModule(w http.ResponseWriter, r *http.Request) {
	if mon == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "USM is not enabled")
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed")
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid request: %s", err)
		return
	}

	if req.Protocol == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid request: protocol is required")
		return
	}

	var protocol *protocols.ProtocolSpec
	var index int
	for i, p := range mon.ebpfProgram.disabledProtocols {
		if p.Name == req.Protocol {
			protocol = p
			index = i
			break
		}
	}
	if protocol == nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid request: protocol %s is not disabled", req.Protocol)
		return
	}
	protocol.ChangeProtocolConfig(mon.cfg, true)
	if protocol.Instance == nil {
		instance, err := protocol.Factory(mon.cfg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error creating protocol instance: %s", err)
			return
		}
		protocol.Instance = instance
		mm, _, _ := mon.ebpfProgram.GetMap("a")
		mon.ebpfProgram.CloneMap()
	}
	for _, p := range protocol.Probes {
		probe, ok := mon.ebpfProgram.GetProbe(p.ProbeIdentificationPair)
		if !ok || probe == nil {
			continue
		}
		if err := probe.Attach(); err != nil {
			log.Errorf("error starting probe: %s", err)
		}
	}
	for _, tc := range protocol.TailCalls {
		if err := mon.ebpfProgram.UpdateTailCallRoutes(tc); err != nil {
			log.Errorf("error stopping tc: %s", err)
		}
	}

	mon.ebpfProgram.enabledProtocols = append(mon.ebpfProgram.enabledProtocols, protocol)
	mon.ebpfProgram.disabledProtocols = append(mon.ebpfProgram.disabledProtocols[:index], mon.ebpfProgram.disabledProtocols[index+1:]...)
	w.WriteHeader(200)
}

func StopModule(w http.ResponseWriter, r *http.Request) {
	if mon == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "USM is not enabled")
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed")
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid request: %s", err)
		return
	}

	if req.Protocol == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid request: protocol is required")
		return
	}

	var protocol *protocols.ProtocolSpec
	var index int
	for i, p := range mon.ebpfProgram.enabledProtocols {
		if p.Name == req.Protocol {
			protocol = p
			index = i
			break
		}
	}
	if protocol == nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid request: protocol %s is not enabled", req.Protocol)
		return
	}
	for _, p := range protocol.Probes {
		if err := mon.ebpfProgram.DetachHook(p.ProbeIdentificationPair); err != nil {
			log.Errorf("error stopping probe: %s", err)
		}
	}
	for _, tc := range protocol.TailCalls {
		if err := mon.ebpfProgram.DetachTailCall(tc); err != nil {
			log.Errorf("error stopping tc: %s", err)
		}
	}

	mon.ebpfProgram.disabledProtocols = append(mon.ebpfProgram.disabledProtocols, protocol)
	mon.ebpfProgram.enabledProtocols = append(mon.ebpfProgram.enabledProtocols[:index], mon.ebpfProgram.enabledProtocols[index+1:]...)
	w.WriteHeader(200)
}

// NewMonitor returns a new Monitor instance
func NewMonitor(c *config.Config, connectionProtocolMap *ebpf.Map) (m *Monitor, err error) {
	defer func() {
		// capture error and wrap it
		if err != nil {
			state = notRunning
			err = fmt.Errorf("could not initialize USM: %w", err)
			startupError = err
		}
	}()

	mgr, err := newEBPFProgram(c, connectionProtocolMap)
	if err != nil {
		return nil, fmt.Errorf("error setting up ebpf program: %w", err)
	}

	if len(mgr.enabledProtocols) == 0 {
		state = disabled
		log.Debug("not enabling USM as no protocols monitoring were enabled.")
		return nil, nil
	}

	if err := mgr.Init(); err != nil {
		return nil, fmt.Errorf("error initializing ebpf program: %w", err)
	}

	filter, _ := mgr.GetProbe(manager.ProbeIdentificationPair{EBPFFuncName: protocolDispatcherSocketFilterFunction, UID: probeUID})
	if filter == nil {
		return nil, fmt.Errorf("error retrieving socket filter")
	}
	ebpfcheck.AddNameMappings(mgr.Manager.Manager, "usm_monitor")

	closeFilterFn, err := filterpkg.HeadlessSocketFilter(c, filter)
	if err != nil {
		return nil, fmt.Errorf("error enabling traffic inspection: %s", err)
	}

	processMonitor := monitor.GetProcessMonitor()

	state = running

	usmMonitor := &Monitor{
		cfg:            c,
		ebpfProgram:    mgr,
		closeFilterFn:  closeFilterFn,
		processMonitor: processMonitor,
	}

	usmMonitor.lastUpdateTime = atomic.NewInt64(time.Now().Unix())

	return usmMonitor, nil
}

// Start USM monitor.
func (m *Monitor) Start() error {
	if m == nil {
		return nil
	}

	var err error

	defer func() {
		if err != nil {
			if errors.Is(err, syscall.ENOMEM) {
				err = fmt.Errorf("could not enable usm monitoring: not enough memory to attach http ebpf socket filter. please consider raising the limit via sysctl -w net.core.optmem_max=<LIMIT>")
			} else {
				err = fmt.Errorf("could not enable USM: %s", err)
			}

			m.Stop()

			startupError = err
		} else {
			mon = m
		}
	}()

	err = m.ebpfProgram.Start()
	if err != nil {
		return err
	}

	// Need to explicitly save the error in `err` so the defer function could save the startup error.
	if m.cfg.EnableNativeTLSMonitoring || m.cfg.EnableGoTLSSupport || m.cfg.EnableJavaTLSSupport || m.cfg.EnableIstioMonitoring || m.cfg.EnableNodeJSMonitoring {
		err = m.processMonitor.Initialize()
	}

	return err
}

// GetUSMStats returns the current state of the USM monitor
func (m *Monitor) GetUSMStats() map[string]interface{} {
	response := map[string]interface{}{
		"state": state,
	}

	if startupError != nil {
		response["error"] = startupError.Error()
	}

	if m != nil {
		response["last_check"] = m.lastUpdateTime
	}
	return response
}

// GetProtocolStats returns the current stats for all protocols
func (m *Monitor) GetProtocolStats() map[protocols.ProtocolType]interface{} {
	if m == nil {
		return nil
	}

	defer func() {
		// Update update time
		now := time.Now().Unix()
		m.lastUpdateTime.Swap(now)
		telemetry.ReportPrometheus()
	}()

	return m.ebpfProgram.getProtocolStats()
}

// Stop HTTP monitoring
func (m *Monitor) Stop() {
	if m == nil {
		return
	}

	m.processMonitor.Stop()

	ebpfcheck.RemoveNameMappings(m.ebpfProgram.Manager.Manager)

	m.ebpfProgram.Close()
	m.closeFilterFn()
}

// DumpMaps dumps the maps associated with the monitor
func (m *Monitor) DumpMaps(w io.Writer, maps ...string) error {
	return m.ebpfProgram.DumpMaps(w, maps...)
}

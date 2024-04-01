// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux_bpf

package tracer

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cihub/seelog"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/DataDog/datadog-agent/pkg/network/events"
	"github.com/DataDog/datadog-agent/pkg/telemetry"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

var defaultFilteredEnvs = []string{
	"DD_ENV",
	"DD_VERSION",
	"DD_SERVICE",
}

const (
	maxProcessQueueLen = 1000
	// maxProcessListSize is the max size of a processList
	maxProcessListSize     = 3
	processCacheModuleName = "network_tracer__process_cache"
	defaultExpiry          = 2 * time.Minute
)

var processCacheTelemetry = struct {
	cacheEvicts   telemetry.Counter
	cacheLength   *prometheus.Desc
	eventsDropped telemetry.Counter
	eventsSkipped telemetry.Counter
}{
	telemetry.NewCounter(processCacheModuleName, "cache_evicts", []string{}, "Counter measuring the number of evictions in the process cache"),
	prometheus.NewDesc(processCacheModuleName+"__cache_length", "Gauge measuring the current size of the process cache", nil, nil),
	telemetry.NewCounter(processCacheModuleName, "events_dropped", []string{}, "Counter measuring the number of dropped process events"),
	telemetry.NewCounter(processCacheModuleName, "events_skipped", []string{}, "Counter measuring the number of skipped process events"),
}

type processList []*events.Process

type processCacheValue struct {
	procs  processList
	exited bool
}

type processCache struct {
	mu sync.RWMutex

	// cache of pid -> list of processes holds a list of processes
	// with the same pid but differing start times up to a max of
	// maxProcessListSize. this is used to determine the closest
	// match to a connection's timestamp
	cacheByPid map[uint32]processCacheValue
	numProcs   int
	maxProcs   int
	// filteredEnvs contains environment variable names
	// that a process in the cache must have; empty filteredEnvs
	// means no filter, and any process can be inserted the cache
	filteredEnvs map[string]struct{}

	in      chan *events.Process
	stopped chan struct{}
	stop    sync.Once
}

type processCacheKey struct {
	pid       uint32
	startTime int64
}

func newProcessCache(maxProcs int, filteredEnvs []string) (*processCache, error) {
	pc := &processCache{
		filteredEnvs: make(map[string]struct{}, len(filteredEnvs)),
		cacheByPid:   map[uint32]processCacheValue{},
		maxProcs:     maxProcs,
		in:           make(chan *events.Process, maxProcessQueueLen),
		stopped:      make(chan struct{}),
	}

	for _, e := range filteredEnvs {
		pc.filteredEnvs[e] = struct{}{}
	}

	go func() {
		for {
			select {
			case <-pc.stopped:
				return
			case p := <-pc.in:
				if p.Exited {
					pc.remove(p)
					continue
				}

				pc.add(p)
			}
		}
	}()

	return pc, nil
}

func (pc *processCache) HandleProcessEvent(entry *events.Process) {

	select {
	case <-pc.stopped:
		return
	default:
	}

	p := pc.processEvent(entry)
	if p == nil {
		processCacheTelemetry.eventsSkipped.Inc()
		return
	}

	select {
	case pc.in <- p:
	default:
		// dropped
		processCacheTelemetry.eventsDropped.Inc()
	}
}

func (pc *processCache) processEvent(entry *events.Process) *events.Process {
	if entry.Exited {
		return entry
	}

	envs := entry.Envs[:0]
	for _, e := range entry.Envs {
		k, _, _ := strings.Cut(e, "=")
		if len(pc.filteredEnvs) > 0 {
			if _, found := pc.filteredEnvs[k]; !found {
				continue
			}
		}

		envs = append(envs, e)

		if len(pc.filteredEnvs) > 0 && len(pc.filteredEnvs) == len(envs) {
			break
		}
	}

	// entry is acceptable if:
	// 1. it has a container ID; or
	// 2. we are not filtering on env vars or at least one of the env vars is present
	if (entry.ContainerID != nil && entry.ContainerID.Get().(string) != "") ||
		len(pc.filteredEnvs) == 0 ||
		len(envs) > 0 {
		entry.Envs = envs
		return entry
	}

	return nil
}

func (pc *processCache) Trim() {
	if pc == nil {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	now := time.Now().Unix()
	trimmed := 0
	for pid, v := range pc.cacheByPid {
		if v.exited {
			trimmed++
			delete(pc.cacheByPid, pid)
			continue
		}

		pl := v.procs
		for p := 0; p < len(pl); {
			if now < pl[p].Expiry {
				p++
				continue
			}

			log.TraceFunc(func() string {
				return fmt.Sprintf("trimming process %d", pid)
			})
			trimmed++
			pl[p], pl[len(pl)-1] = pl[len(pl)-1], pl[p]
			pl = pl[:len(pl)-1]
		}

		if len(pl) == 0 {
			delete(pc.cacheByPid, pid)
			continue
		}

		v.procs = pl
		pc.cacheByPid[pid] = v
	}

	if trimmed > 0 {
		log.Debugf("Trimmed %d process cache entries", trimmed)
	}

	pc.numProcs -= trimmed
}

func (pc *processCache) Stop() {
	if pc == nil {
		return
	}

	pc.stop.Do(func() { close(pc.stopped) })
}

func (pc *processCache) remove(p *events.Process) {
	if pc == nil {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	v := pc.cacheByPid[p.Pid]
	v.exited = true
	pc.cacheByPid[p.Pid] = v
}

func (pc *processCache) add(p *events.Process) {
	if pc == nil {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.numProcs >= pc.maxProcs {
		processCacheTelemetry.eventsDropped.Inc()
		return
	}

	if log.ShouldLog(seelog.TraceLvl) {
		log.Tracef("adding process %+v to process cache", p)
	}

	p.Expiry = time.Now().Add(defaultExpiry).Unix()
	v := pc.cacheByPid[p.Pid]
	pl := v.procs
	var added bool
	if pl, added = pl.update(p); added {
		pc.numProcs++
	}

	v.procs = pl
	v.exited = false
	pc.cacheByPid[p.Pid] = v
}

func (pc *processCache) Get(pid uint32, ts int64) (*events.Process, bool) {
	if pc == nil {
		return nil, false
	}

	pc.mu.RLock()
	defer pc.mu.RUnlock()

	log.TraceFunc(func() string { return fmt.Sprintf("looking up pid %d", pid) })

	v := pc.cacheByPid[pid]
	pl := v.procs
	if closest := pl.closest(ts); closest != nil {
		closest.Expiry = time.Now().Add(defaultExpiry).Unix()
		log.TraceFunc(func() string { return fmt.Sprintf("found entry for pid %d: %+v", pid, closest) })
		return closest, true
	}

	log.TraceFunc(func() string { return fmt.Sprintf("entry not found for process %d", pid) })
	return nil, false
}

func (pc *processCache) Dump() (interface{}, error) {
	res := map[uint32]interface{}{}
	if pc == nil {
		return res, nil
	}

	pc.mu.RLock()
	defer pc.mu.RUnlock()

	for pid, pl := range pc.cacheByPid {
		res[pid] = pl
	}

	return res, nil
}

// Describe returns all descriptions of the collector.
func (pc *processCache) Describe(ch chan<- *prometheus.Desc) {
	ch <- processCacheTelemetry.cacheLength
}

// Collect returns the current state of all metrics of the collector.
func (pc *processCache) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(processCacheTelemetry.cacheLength, prometheus.GaugeValue, float64(pc.numProcs))
}

func (pl processList) update(p *events.Process) (processList, bool) {
	for i := range pl {
		if pl[i].StartTime == p.StartTime {
			pl[i] = p
			return pl, false
		}
	}

	added := len(pl) < maxProcessListSize
	if len(pl) == maxProcessListSize {
		copy(pl, pl[1:])
		pl = pl[:len(pl)-1]
	}

	if pl == nil {
		pl = make(processList, 0, maxProcessListSize)
	}

	return append(pl, p), added
}

func (pl processList) remove(p *events.Process) processList {
	for i := range pl {
		if pl[i] == p {
			return append(pl[:i], pl[i+1:]...)
		}
	}

	return pl
}

func abs(i int64) int64 {
	if i < 0 {
		return -i
	}

	return i
}

func (pl processList) closest(ts int64) *events.Process {
	var closest *events.Process
	for i := range pl {
		if ts >= pl[i].StartTime &&
			(closest == nil ||
				abs(closest.StartTime-ts) > abs(pl[i].StartTime-ts)) {
			closest = pl[i]
		}
	}

	return closest
}

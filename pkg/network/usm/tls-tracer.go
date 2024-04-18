// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux

package usm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-agent/pkg/network/protocols"
)

type RequestBody struct {
	PID  uint32 `json:"pid"`
	Type string `json:"type"`
}

var (
	ebpfMgr *ebpfProgram
)

type callbackType uint8

const (
	attach callbackType = iota
	detach
)

func (m callbackType) String() string {
	switch m {
	case attach:
		return "attach"
	case detach:
		return "detach"
	default:
		return "unknown"
	}
}

func runTLSCallback(w http.ResponseWriter, r *http.Request, mode callbackType) {
	if ebpfMgr == nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Monitor not initialized")
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Only POST requests are allowed")
		return
	}

	var reqBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error decoding request body: %v", err)
		return
	}

	// Validate the type field
	switch reqBody.Type {
	case "go-tls", "native", "nodejs", "istio":
		var found bool
		var moduleTLS protocols.ProtocolTLS
		for _, module := range ebpfMgr.enabledProtocols {
			if module.Raw.(protocols.Protocol).Name() == reqBody.Type {
				moduleTLS, found = module.Raw.(protocols.ProtocolTLS)
				break
			}
		}
		if !found {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Module %q is not enabled", reqBody.Type)
			return
		}
		cb := moduleTLS.AttachPID
		if mode == detach {
			cb = moduleTLS.DetachPID
		}
		if err := cb(reqBody.PID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error %sing PID: %v", mode.String(), err)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s successfully %sed PID %d", reqBody.Type, mode.String(), reqBody.PID)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid 'type' value provided")
		return
	}
}

func AttachPIDEndpoint(w http.ResponseWriter, r *http.Request) {
	runTLSCallback(w, r, attach)
}

func DetachPIDEndpoint(w http.ResponseWriter, r *http.Request) {
	runTLSCallback(w, r, detach)
}

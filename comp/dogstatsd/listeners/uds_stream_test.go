// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build !windows

// UDS won't work in windows

package listeners

import (
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/DataDog/datadog-agent/comp/core/config"
	"github.com/DataDog/datadog-agent/comp/core/telemetry"
	workloadmeta "github.com/DataDog/datadog-agent/comp/core/workloadmeta/def"
	"github.com/DataDog/datadog-agent/comp/dogstatsd/packets"
	"github.com/DataDog/datadog-agent/comp/dogstatsd/pidmap"
	"github.com/DataDog/datadog-agent/pkg/util/optional"
)

func udsStreamListenerFactory(packetOut chan packets.Packets, manager *packets.PoolManager[packets.Packet], cfg config.Component, pidMap pidmap.Component, telemetryStore *TelemetryStore, packetsTelemetryStore *packets.TelemetryStore, telemetry telemetry.Component) (StatsdListener, error) {
	return NewUDSStreamListener(packetOut, manager, nil, cfg, nil, optional.NewNoneOption[workloadmeta.Component](), pidMap, telemetryStore, packetsTelemetryStore, telemetry)
}

func TestNewUDSStreamListener(t *testing.T) {
	testNewUDSListener(t, udsStreamListenerFactory, "unix")
}

func TestStartStopUDSStreamListener(t *testing.T) {
	testStartStopUDSListener(t, udsStreamListenerFactory, "unix")
}

func TestUDSStreamReceive(t *testing.T) {
	socketPath := testSocketPath(t)

	mockConfig := map[string]interface{}{}
	mockConfig[socketPathConfKey("unix")] = socketPath
	mockConfig["dogstatsd_origin_detection"] = false

	var contents0 = []byte("daemon:666|g|#sometag1:somevalue1,sometag2:somevalue2")
	var contents1 = []byte("daemon:999|g|#sometag1:somevalue1")

	packetsChannel := make(chan packets.Packets)

	deps := fulfillDepsWithConfig(t, mockConfig)
	telemetryStore := NewTelemetryStore(nil, deps.Telemetry)
	packetsTelemetryStore := packets.NewTelemetryStore(nil, deps.Telemetry)
	s, err := udsStreamListenerFactory(packetsChannel, newPacketPoolManagerUDS(deps.Config, packetsTelemetryStore), deps.Config, deps.PidMap, telemetryStore, packetsTelemetryStore, deps.Telemetry)
	assert.Nil(t, err)
	assert.NotNil(t, s)

	s.Listen()
	defer s.Stop()
	conn, err := net.Dial("unix", socketPath)
	assert.Nil(t, err)
	defer conn.Close()

	binary.Write(conn, binary.LittleEndian, int32(len(contents0)))
	conn.Write(contents0)

	binary.Write(conn, binary.LittleEndian, int32(len(contents1)))
	conn.Write(contents1)

	select {
	case pkts := <-packetsChannel:
		assert.Equal(t, 2, len(pkts))

		packet := pkts[0]
		assert.NotNil(t, packet)
		assert.Equal(t, packet.Contents, contents0)
		assert.Equal(t, packet.Origin, "")
		assert.Equal(t, packet.Source, packets.UDS)

		packet = pkts[1]
		assert.NotNil(t, packet)
		assert.Equal(t, packet.Contents, contents1)
		assert.Equal(t, packet.Origin, "")
		assert.Equal(t, packet.Source, packets.UDS)

	case <-time.After(2 * time.Second):
		assert.FailNow(t, "Timeout on receive channel")
	}

}

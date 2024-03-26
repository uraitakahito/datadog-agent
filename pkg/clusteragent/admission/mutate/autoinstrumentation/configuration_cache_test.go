// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver

package autoinstrumentation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateConfiguration(t *testing.T) {
	enabled := false
	c := newInstrumentationConfigurationCache(nil, &enabled, nil, nil, "")

	c.updateConfiguration(true, nil, "")
	require.Equal(t, true, c.currentConfiguration.enabled)
}

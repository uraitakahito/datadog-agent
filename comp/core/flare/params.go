// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package flare

// Params defines the parameters for the flare component.
type Params struct {
	// local is set to true when we could not contact a running Agent and the flare is created directly from the
	// CLI.
	local bool
}

// NewLocalParams returns parameters for to initialize a local flare component. Local flares are meant to be created by
// the CLI process instead of the main Agent one.
func NewLocalParams() Params {
	return Params{
		local: true,
	}
}

// NewParams returns parameters for to initialize a non local flare component
func NewParams() Params {
	return Params{
		local: false,
	}
}

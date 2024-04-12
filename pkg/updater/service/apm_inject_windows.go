// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build windows

// Package service provides a way to interact with os services
package service

// SetupAPMInjector noop
func SetupAPMInjector() error {
	return nil
}

// RemoveAPMInjector noop
func RemoveAPMInjector() error {
	return nil
}

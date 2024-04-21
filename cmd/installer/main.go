// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

// Package main implements 'installer'.
package main

import (
	"log"
	"os"

	"github.com/DataDog/datadog-agent/cmd/installer/command"
	"github.com/DataDog/datadog-agent/cmd/installer/subcommands"
	"github.com/DataDog/datadog-agent/cmd/internal/runcmd"
)

func main() {
	err := dropPrivileges()
	if err != nil {
		log.Printf("error dropping privileges: %v", err)
	}
	os.Exit(runcmd.Run(command.MakeCommand(subcommands.InstallerSubcommands())))
}

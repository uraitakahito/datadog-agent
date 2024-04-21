// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

//go:build !windows

// Package main implements 'installer'.
package main

import (
	"fmt"
	"os/user"
	"strconv"
	"syscall"
)

func dropPrivileges() error {
	fmt.Println("dropping privileges from root to dd-agent")
	userID := syscall.Getuid()
	if userID != 0 {
		return fmt.Errorf("the installer requires root privileges to manage datadog packages")
	}
	usr, err := user.Lookup("dd-agent")
	if err != nil {
		return fmt.Errorf("could not find dd-agent user: %v", err)
	}
	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		return fmt.Errorf("could not convert dd-agent user id to int: %v", err)
	}
	grp, err := user.LookupGroup("dd-agent")
	if err != nil {
		return fmt.Errorf("could not find dd-agent group: %v", err)
	}
	gid, err := strconv.Atoi(grp.Gid)
	if err != nil {
		return fmt.Errorf("could not convert dd-agent group id to int: %v", err)
	}
	err = syscall.Seteuid(uid)
	if err != nil {
		return fmt.Errorf("could not set uid to dd-agent user: %v", err)
	}
	err = syscall.Setegid(gid)
	if err != nil {
		return fmt.Errorf("could not set gid to dd-agent group: %v", err)
	}
	return nil
}

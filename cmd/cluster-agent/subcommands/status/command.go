// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build !windows && kubeapiserver

// Package status implements 'cluster-agent status'.
package status

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/DataDog/datadog-agent/cmd/cluster-agent/command"
	"github.com/DataDog/datadog-agent/comp/core"
	"github.com/DataDog/datadog-agent/comp/core/config"
	log "github.com/DataDog/datadog-agent/comp/core/log/def"
	"github.com/DataDog/datadog-agent/comp/core/secrets"
	"github.com/DataDog/datadog-agent/pkg/api/util"
	pkgconfig "github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
)

type cliParams struct {
	jsonStatus      bool
	prettyPrintJSON bool
	statusFilePath  string
}

// Commands returns a slice of subcommands for the 'cluster-agent' command.
func Commands(globalParams *command.GlobalParams) []*cobra.Command {
	cliParams := &cliParams{}
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print the current status",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fxutil.OneShot(run,
				fx.Supply(cliParams),
				fx.Supply(core.BundleParams{
					ConfigParams: config.NewClusterAgentParams(globalParams.ConfFilePath),
					SecretParams: secrets.NewEnabledParams(),
					LogParams:    log.ForOneShot(command.LoggerName, command.DefaultLogLevel, true),
				}),
				core.Bundle(),
			)
		},
	}

	cmd.Flags().BoolVarP(&cliParams.jsonStatus, "json", "j", false, "print out raw json")
	cmd.Flags().BoolVarP(&cliParams.prettyPrintJSON, "pretty-json", "p", false, "pretty print JSON")
	cmd.Flags().StringVarP(&cliParams.statusFilePath, "file", "o", "", "Output the status command to a file")

	return []*cobra.Command{cmd}
}

//nolint:revive // TODO(CINT) Fix revive linter
func run(log log.Component, config config.Component, cliParams *cliParams) error {
	fmt.Printf("Getting the status from the agent.\n")
	var e error
	var s string
	c := util.GetClient(false) // FIX: get certificates right then make this true

	v := url.Values{}
	if cliParams.prettyPrintJSON || cliParams.jsonStatus {
		v.Set("format", "json")
	} else {
		v.Set("format", "text")
	}

	url := url.URL{
		Scheme:   "https",
		Host:     fmt.Sprintf("localhost:%v", pkgconfig.Datadog().GetInt("cluster_agent.cmd_port")),
		Path:     "/status",
		RawQuery: v.Encode(),
	}

	// Set session token
	e = util.SetAuthToken(config)
	if e != nil {
		return e
	}

	r, e := util.DoGet(c, url.String(), util.LeaveConnectionOpen)
	if e != nil {
		var errMap = make(map[string]string)
		json.Unmarshal(r, &errMap) //nolint:errcheck
		// If the error has been marshalled into a json object, check it and return it properly
		if err, found := errMap["error"]; found {
			e = fmt.Errorf(err)
		}

		fmt.Printf(`
		Could not reach agent: %v
		Make sure the agent is running before requesting the status.
		Contact support if you continue having issues.`, e)
		return e
	}

	// The rendering is done in the client so that the agent has less work to do
	if cliParams.prettyPrintJSON {
		var prettyJSON bytes.Buffer
		json.Indent(&prettyJSON, r, "", "  ") //nolint:errcheck
		s = prettyJSON.String()
	} else {
		s = string(r)
	}

	if cliParams.statusFilePath != "" {
		os.WriteFile(cliParams.statusFilePath, []byte(s), 0644) //nolint:errcheck
	} else {
		fmt.Println(s)
	}

	return nil
}

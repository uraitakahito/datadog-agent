// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package examples

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/environments"
	awsdocker "github.com/DataDog/datadog-agent/test/new-e2e/pkg/environments/aws/docker"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/utils/e2e/client/agentclient"
	"github.com/DataDog/test-infra-definitions/components/datadog/dockeragentparams"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type dockerSuiteExtra struct {
	e2e.BaseSuite[environments.DockerHost]
}

const extraComposeContent = `version: '2'
services:
  redis:
    image: 'bitnami/redis:latest'
    environment:
      - ALLOW_EMPTY_PASSWORD=yes
`

func TestDockerExtra(t *testing.T) {
	e2e.Run(t, &dockerSuiteExtra{}, e2e.WithProvisioner(awsdocker.Provisioner(awsdocker.WithoutFakeIntake(), awsdocker.WithAgentOptions(dockeragentparams.WithExtraComposeManifest("docker-compose.extra.yaml", pulumi.String(extraComposeContent))))))
}

func (v *dockerSuiteExtra) TestExecuteCommand() {
	agentVersion := v.Env().Agent.Client.Version()
	regexpVersion := regexp.MustCompile(`.*Agent .* - Commit: .* - Serialization version: .* - Go version: .*`)

	v.Require().Truef(regexpVersion.MatchString(agentVersion), fmt.Sprintf("%v doesn't match %v", agentVersion, regexpVersion))
	// args is used to test client.WithArgs. The values of the arguments are not relevant.
	args := agentclient.WithArgs([]string{"-n", "-c", "."})
	version := v.Env().Agent.Client.Version(args)

	v.Require().Truef(regexpVersion.MatchString(version), fmt.Sprintf("%v do esn't match %v", version, regexpVersion))
}

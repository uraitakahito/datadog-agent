$ErrorActionPreference = "Stop"

$Env:Python2_ROOT_DIR=$Env:TEST_EMBEDDED_PY2
$Env:Python3_ROOT_DIR=$Env:TEST_EMBEDDED_PY3
$tmpfile = [System.IO.Path]::GetTempFileName()
& "C:\mnt\tools\ci\aws_ssm_get_wrapper.ps1" "$Env:DOCKER_REGISTRY_LOGIN_SSM_KEY" > $tmpfile
If ($LASTEXITCODE -ne "0") {
    exit $LASTEXITCODE
}
$DOCKER_REGISTRY_LOGIN = $(cat $tmpfile)

& "C:\mnt\tools\ci\aws_ssm_get_wrapper.ps1" "$Env:AGENT_GITHUB_APP_ID_SSM_NAME" > $tmpfile
If ($LASTEXITCODE -ne "0") {
    exit $LASTEXITCODE
}
$DOCKER_REGISTRY_PWD = $(cat $tmpfile)
echo "docker login --username \"${DOCKER_REGISTRY_LOGIN}\" --password \"${DOCKER_REGISTRY_PWD}\" \"docker.io\""
If ($LASTEXITCODE -ne "0") {
    throw "Previous command returned $LASTEXITCODE"
}
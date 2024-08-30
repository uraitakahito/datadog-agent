if not exist c:\mnt\ goto nomntdir

@echo c:\mnt found, continuing

@echo PARAMS %*
@echo PY_RUNTIMES %PY_RUNTIMES%

if NOT DEFINED PY_RUNTIMES set PY_RUNTIMES=%~1

set TEST_ROOT=c:\test-root
mkdir %TEST_ROOT%\datadog-agent
cd %TEST_ROOT%\datadog-agent
xcopy /e/s/h/q c:\mnt\*.*

Powershell -C "%TEST_ROOT%\datadog-agent\tasks\winbuildscripts\test.ps1" || exit /b 2

goto :EOF

:nomntdir
@echo directory not mounted, parameters incorrect
exit /b 1

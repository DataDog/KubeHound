@echo off

REM Lightweight wrapper script to run KubeHound from a release archive

set KUBEHOUND_ENV=release
set DOCKER_CMD=docker
set DOCKER_COMPOSE_FILE_PATH=-f deployments\kubehound\docker-compose.yaml
set DOCKER_COMPOSE_FILE_PATH=%DOCKER_COMPOSE_FILE_PATH% -f deployments\kubehound\docker-compose.release.yaml
if not "%DD_API_KEY%"=="" (
    set DOCKER_COMPOSE_FILE_PATH=%DOCKER_COMPOSE_FILE_PATH% -f deployments\kubehound\docker-compose.datadog.yaml
)

set DOCKER_COMPOSE_PROFILE=--profile infra

:run
 ./kubehound -c config.yaml
goto :eof

:backend-down
%DOCKER_CMD% compose %DOCKER_COMPOSE_FILE_PATH% %DOCKER_COMPOSE_PROFILE% rm -fvs
goto :eof

:backend-up
%DOCKER_CMD% compose %DOCKER_COMPOSE_FILE_PATH% %DOCKER_COMPOSE_PROFILE% up --force-recreate --build -d
goto :eof

:backend-reset
%DOCKER_CMD% compose %DOCKER_COMPOSE_FILE_PATH% %DOCKER_COMPOSE_PROFILE% rm -fvs
%DOCKER_CMD% compose %DOCKER_COMPOSE_FILE_PATH% %DOCKER_COMPOSE_PROFILE% up --force-recreate --build -d
goto :eof

:backend-reset-hard
call :backend-down
%DOCKER_CMD% volume rm kubehound-%KUBEHOUND_ENV%_mongodb_data
%DOCKER_CMD% volume rm kubehound-%KUBEHOUND_ENV%_kubegraph_data
call :backend-up
goto :eof

if "%1"=="" (
    echo Usage: %0 {run^|backend-up^|backend-reset^|backend-reset-hard^|backend-down}
    exit /b 1
)

setlocal enabledelayedexpansion
for %%i in (run backend-up backend-reset backend-reset-hard backend-down) do (
    if "%%i"=="%1" (
        endlocal
        goto %%i
    )
)
endlocal

echo Usage: %0 {run^|backend-up^|backend-reset^|backend-reset-hard^|backend-down}
exit /b 1

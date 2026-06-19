#!/usr/bin/env bash

# Script generado por Gemini Flash usando de referencia testStdinArguments.sh

# Directorio del script actual
scriptDir=$(cd "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

# Comando a testear
cmdTest="go run ${scriptDir}/../../../cmd/cli/main.go show"
awkRealTime='{ print $2 }'

StdinPipeTest() {
    local limit=5
    local expOutput=$'1\n2\n3\n4\n5'
    local logFile

    # Usamos un archivo temporal real
    logFile=$(mktemp)

    (
        for i in $(seq 1 "$limit"); do
            printf "SELECT '%s'\n" "$i"
            sleep 0.15
        done
    ) | ${cmdTest} -p 2>&1 | awk -F"'" "${awkRealTime}" > "$logFile"

    local cmdOutput
    cmdOutput=$(cat "$logFile")
    rm -f "$logFile"

    if [[ "${cmdOutput}" != "${expOutput}" ]]; then
        echo -e "\033[0;31m[testStdinPipe]\033[0m\noutput:\n${cmdOutput}\nExpected:\n${expOutput}" >&2
        exit 1
    else
        echo -e "\033[0;32m[testStdinPipe] PASS\033[0m\noutput:\n${cmdOutput}"
    fi
}

StdinPipeTest

exit 0

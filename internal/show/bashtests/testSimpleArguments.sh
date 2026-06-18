#!/usr/bin/env bash

declare -A dummy
dummy=(
  ["well.sql"]="SELECT '1-Well'" 
  ["done.sql"]="SELECT '2-Done!'" 
  ["friendo.sql"]="SELECT '3-Friendo'"
)

cmdTest="go run ../../cmd/cli/main.go show"

# Writing dummy files
for path in "${!dummy[@]}"; do
  content="${dummy[$path]}"
  echo "${content}" > "${path}"
done

# Capturing the Stdout from the command we want to test
cmdStdout=$(${cmdTest} "${!dummy[@]}" -p)
awkScript='{  sub(/^[0-9]-/,"", $2); printf "%s ",$2 }'
cmdStdout=$(echo "${cmdStdout}"| sort | awk -F"'" "${awkScript}" | xargs)

# Cleaning
for path in "${!dummy[@]}"; do
  rm "${path}"
done

# Body of the test
expOutput="Well Done! Friendo"
if [[ "${cmdStdout}" !=  "${expOutput}" ]]; then
  echo -e "[testSimpleArguments.OneDummyArgument]\noutput:${cmdStdout}\nExpected:${expOutput}" >&2
  exit 1
fi

exit 0

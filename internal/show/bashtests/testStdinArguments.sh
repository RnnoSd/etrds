#!/usr/bin/env bash

scriptDir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

# Command to be tested
cmdTest="go run ${scriptDir}/../../../cmd/cli/main.go show"
awkRealTime='{ sub(/^[0-9]-/,"", $2); printf "%s\n", $2 }'

# First Test:
OnlyStdinTest(){
  local limit=5
  local expOutput=$'1\n2\n3\n4\n5'
  local signal=1

  # Open a channel to see the interaction of the command as if it were done 
  # by the user (First In First Out) a kind of file used in linx for these tasks
  exec 3> >(
    stdbuf -oL ${cmdTest} - -p |\
    awk -F"'" "${awkRealTime}" > output.log 2>&1
  )
  local pid=$!

  while [[ "${signal}" -le "${limit}" ]]; do
    # send a signal to the channel 3 which was created
    echo "SELECT '${signal}'" >&3
    sleep 0.15

    signal=$((signal + 1))
    # Use kill -0 to see if the process is alive
    if ! kill -0 $pid 2>/dev/null; then
      break
    fi
  done

  # Close the channel
  exec 3>&-
  # Collect all the remainings of the process left by the Kernel
  wait $pid 2>/dev/null
  local cmdOutput=$(cat output.log)
  rm output.log
  if [[ "${cmdOutput}" != "${expOutput}" ]]; then
    echo -e "\033[0;31m[StdinArguments.OnlyStdin]\033[0m\noutput:\n${cmdOutput}\nExpected:\n${expOutput}" >&2
    exit 1
  else
    echo -e "\033[0;32m[testStdinArguments.OnlyStdin] PASS\033[0m\noutput:\n${cmdOutput}"
  fi
}

# Second Test
ArgumentsAndStdinTest() {
  local StdinStatic="SELECT 'This is a static Stdin'"
  local StdinDynamic=(
    "1-This" 
    "2-is"
    "3-dynamic"
  )
	declare -A dummy
  echo "SELECT 'This is an argument'" > congratulations.sql
  local expOutput=$'This is a static Stdin\nThis is an argument\nThis\nis\ndynamic'

  # Open a channel to see the interaction of the command as if it were done 
  # by the user (First In First Out) a kind of file used in linx for these tasks
  exec 3> >(
    stdbuf -oL ${cmdTest} <(echo "${StdinStatic}") - congratulations.sql  -p |\
    awk -F"'" "${awkRealTime}" > output.log 2>&1
  )
  local pid=$!

  for StdinSignal in $(printf '%s\n' "${StdinDynamic[@]}" | sort); do
    echo "SELECT '${StdinSignal}'" >&3
    sleep 0.15

    if ! kill -0 $pid 2>/dev/null; then
      break
    fi
  done

  # Close the channel
  exec 3>&-
  # Collect all the remainings of the process left by the Kernel
  wait $pid 2>/dev/null
  local cmdOutput=$(cat output.log)
  rm congratulations.sql output.log

  if [[ "${cmdOutput}" != "${expOutput}" ]]; then
    echo -e "\033[0;31m[StdinArguments.OnlyStdin]\033[0m\noutput:\n${cmdOutput}\nExpected:\n${expOutput}" >&2
  else
    echo -e "\033[0;32m[testStdinArguments.OnlyStdin] PASS\033[0m\noutput:\n${cmdOutput}"
    exit 1
  fi
}

OnlyStdinTest
ArgumentsAndStdinTest

exit 0

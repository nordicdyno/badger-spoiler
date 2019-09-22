#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

declare -a BadgerVersions=(
  "v1.6.0"
  "v2.0.0-rc2"
  "master"
)

# if not empty test should start generator in slow mode and kill it with provided signal number.
declare -a KillGenSignal=(
  "9"
  ""
)

for ver in "${BadgerVersions[@]}"; do
  for killVal in "${KillGenSignal[@]}"; do
    export KILL_GEN_SIGNAL=$killVal
    set +e
    go run make.go -v=$ver -- "$@"
    echo "###### KILL_GEN_SIGNAL=$KILL_GEN_SIGNAL BADGER=$ver test exit code: $?"
    set -e
  done
done

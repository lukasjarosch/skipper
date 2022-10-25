#!/usr/bin/env bash

examplesDir="$1"
exitCode=0

for dir in ${examplesDir}/*; do

  cd $dir

  if ! go mod tidy; then
    echo "FAILED to 'go mod tidy' $dir"
    exitCode=1
    continue
  fi

  if ! go build -buildvcs=false; then
    echo "FAILED to 'go build' $dir"
    exitCode=1
    continue
  fi

done

exit $exitCode

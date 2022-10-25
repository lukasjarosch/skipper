#!/usr/bin/env bash

examplesDir="$1"
exitCode=0

for dir in ${examplesDir}/*; do
  echo "Entering $dir"
  cd $dir

  if ! go mod tidy; then
    echo "Error: failed to 'go mod tidy' $dir"
    exitCode=1
    continue
  fi

  if ! go build -buildvcs=false; then
    echo "Error: failed to 'go build' $dir"
    exitCode=1
    continue
  else
    echo "::notice::Compiled successfully: $dir"
  fi

done

exit $exitCode

#!/usr/bin/env bash

examplesDir="$1"
testdataDir="testdata"
testFile="main_test.go"
exitCode=0

for dir in ${examplesDir}/*; do
  echo "Entering $dir"
  cd $dir

  # examples usually only have a 'main.go', hence the test 'main_test.go' must exist
  if [[ ! -f $testFile ]]; then
    echo "MISSING test file '$testFile'"
    exitCode=1
    continue
  fi

  # testdata directory must exist
  if [[ ! -d $testdataDir ]]; then
    echo "MISSING '$testdataDir' directory"
    exitCode=1
    continue
  fi

  # at least one golden file fixture needs to be present
  fileCount=$(ls $testdataDir | wc -l)
  if [[ ! $fileCount -gt 0 ]]; then
    echo "MISSING golden files in '$testdataDir'"
    exitCode=1
    continue
  fi

  # TODO: execute test

done

exit $exitCode

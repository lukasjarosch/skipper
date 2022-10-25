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
    echo "::error:: Missing test file '$testFile'"
    exitCode=1
    continue
  fi

  # testdata directory must exist
  if [[ ! -d $testdataDir ]]; then
    echo "::error:: Missing '$testdataDir' directory"
    exitCode=1
    continue
  fi

  # at least one golden file fixture needs to be present
  fileCount=$(ls $testdataDir | wc -l)
  if [[ ! $fileCount -gt 0 ]]; then
    echo "::error:: no golden files in '$testdataDir'"
    continue
  fi

  # run actual test
  go test -v

done

exit $exitCode

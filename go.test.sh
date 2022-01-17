#!/usr/bin/env bash

set -e
echo "" > coverage.txt

function test {
  for d in $(go list ./... | grep -v vendor); do
      go test -v -race -coverprofile=profile.out -covermode=atomic "$d"
      if [ -f profile.out ]; then
          cat profile.out >> $coverage
          rm profile.out
      fi
  done
}

coverage=../coverage.txt
cd xgorm && test && cd ..
cd xgormv2 && test && cd ..
cd xneo4j && test && cd ..
cd xredis && test && cd ..

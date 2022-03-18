#!/bin/bash

set -e
function test_module {
    for package in $(go list ./...); do
        go test $package -v -count=1 -race -cover -covermode=atomic -coverprofile=$profile
        test -f $profile && cat $profile >> $coverage && rm $profile
    done
}

rm -f coverage.txt
coverage=../coverage.txt
profile=../profile.out

cd xgorm   && test_module && cd ..
cd xgormv2 && test_module && cd ..
cd xneo4j  && test_module && cd ..
cd xredis  && test_module && cd ..

#!/bin/sh

cd /go/src/github.com/mdouchement/iosupport/

echo '##### Installing system dependencies'
apk upgrade
apk add --update --no-cache git

echo '##### Installing dependencies'
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega
go get github.com/golang/mock/gomock
go get -d -t -v ./...

echo '##### Running specs'
ginkgo -r

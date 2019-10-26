#!/usr/bin/env bash
cd $GOPATH/src/github.com/pingcap/tidb-operator
go build -o operator/tidb-controller-manager cmd/controller-manager/main.go
go build -o operator/tidb-scheduler cmd/scheduler/main.go
go build -o operator/tidb-discovery cmd/discovery/main.go
go build -o operator/tidb-admission-controller  cmd/admission-controller/main.go



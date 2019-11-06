#!/usr/bin/env bash
kubectl exec tidb-cluster-pd-0 -- ./pd-ctl stores show -u http://127.0.0.1:2379
FROM alpine:3.5
RUN apk add tzdata --no-cache
ADD tidb-controller-manager /usr/local/bin/tidb-controller-manager
ADD tidb-scheduler /usr/local/bin/tidb-scheduler
ADD tidb-discovery /usr/local/bin/tidb-discovery
ADD tidb-admission-controller /usr/local/bin/tidb-admission-controller


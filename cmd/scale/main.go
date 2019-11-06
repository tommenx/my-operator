package main

import (
	"flag"
	"fmt"
	"github.com/pingcap/tidb-operator/pkg/cdapi"
	"github.com/pingcap/tidb-operator/pkg/kubeutil"
	"github.com/pingcap/tidb-operator/pkg/scaler"
	"github.com/pingcap/tidb-operator/pkg/store"
	"sync"
	"time"
)

var (
	coordinator    string
	configPath     string
	pdURL          string
	defaultTimeout = time.Second * 5
	kind           string
	wg             sync.WaitGroup
	scaleOutStep   = int32(2)
	replica        = int32(4)
	resourceLimit  = 60
	etcdPath       string
)

func init() {
	flag.Set("logtostderr", "true")
	flag.StringVar(&kind, "kind", "old", "specify auto scale kind")
	flag.StringVar(&coordinator, "coordinator", "10.77.30.147:8888", "specify coordinator path")
	flag.StringVar(&configPath, "config", "/root/.kube/config", "specify config path")
	flag.StringVar(&etcdPath, "etcd", "127.0.0.1:2389", "specify etcd url")
}

func main() {
	flag.Parse()
	coordinatorAPI := fmt.Sprintf("http://%s", coordinator)
	cd := cdapi.NewCDClient(coordinatorAPI, defaultTimeout)
	kube := kubeutil.NewKubeClient(configPath)
	scale := scaler.NewScaleController(cd, kube)
	db := store.NewEtcdHandler([]string{etcdPath})
	stopCh := make(chan struct{})
	var autoScale scaler.AutoScale
	if kind == "old" {
		autoScale = scaler.NewOldAutoScale(cd, scale, db, scaleOutStep, replica, resourceLimit)
	} else if kind == "fancy" {
		autoScale = scaler.NewFancyAutoScale(cd, scale, db, scaleOutStep, replica, resourceLimit)
	}
	wg.Add(1)
	go autoScale.Run(stopCh)
	wg.Wait()
}

package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/pingcap/tidb-operator/pkg/cdapi"
	"github.com/pingcap/tidb-operator/pkg/kubeutil"
	scaler "github.com/pingcap/tidb-operator/pkg/scaler"
	"strconv"
	"time"
)

var (
	upBound     = 10
	originCount = 2
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {
	flag.Parse()
	coordinatorAPI := "http://10.77.30.147:8888"
	configPath := "/Users/tommenx/.kube/config"
	timeout := time.Second * 5
	period := time.Second * 10
	cd := cdapi.NewCDClient(coordinatorAPI, timeout)
	kube := kubeutil.NewFakeKubeClient(configPath)
	scale := scaler.NewScaleController(cd, kube)
	ticker := time.NewTicker(period)
	for {
		select {
		case <-ticker.C:
			go checkAndScale(cd, scale)
		}
	}
}

func checkAndScale(cd cdapi.CDClient, scale scaler.ScaleController) error {
	instances, err := cd.GetTiKVStatus()
	if err != nil {
		glog.Errorf("get tikve status error")
	}
	write := float64(0)
	read := float64(0)
	for _, instance := range instances {
		r, _ := strconv.ParseFloat(instance.Read, 64)
		w, _ := strconv.ParseFloat(instance.Write, 64)
		write += w
		read += r
		glog.Infof("%s read:%s write:%s\n", instance.Name, instance.Read, instance.Write)
	}
	avgwrite := int(write) / len(instances)
	avgread := int(read) / len(instances)
	glog.Infof("avg read:%d write:%d\n", avgread, avgwrite)
	if avgwrite > upBound || avgread > upBound {
		glog.Infof("got to warning bound, start to scale out, write:%d,read=%d\n", write, read)
		err = scale.ScaleOut("default", "ubuntu", 2)
		if err != nil {
			glog.Errorf("scale out error, err=%+v", err)
			return err
		}
	}
	return nil
}

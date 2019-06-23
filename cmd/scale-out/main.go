package main

import (
	"context"
	"flag"
	"github.com/golang/glog"
	"github.com/pingcap/tidb-operator/pkg/kubeutils"
	"github.com/pingcap/tidb-operator/pkg/monitor"
	"github.com/robfig/cron"
	"log"
	"math"
)

var (
	address    = "http://10.77.110.140:30764"
	GB         = int64(1 << 30)
	One        = 10
	masterUrl  string
	configPath string
	namespace  string
	name       string
)

func init() {
	flag.Set("logtostderr", "true")
	flag.StringVar(&masterUrl, "materUrl", "", "use to get master url")
	flag.StringVar(&configPath, "kubeconfig", "/root/.kube/config", "kubernetes config path")
	flag.StringVar(&namespace, "ns", "tidb", "get tidb cluster namespace")
	flag.StringVar(&name, "name", "tidb-cluster", "name of tidbcluster")
}

func main() {
	flag.Parse()
	cron := cron.New()
	spec := "*/5 * * * * ?"
	cron.AddFunc(spec, check)
	cron.Start()
	select {}
	//fmt.Println(GB)

}

func check() {
	ctx := context.TODO()
	monitor := monitor.NewMonitor(address)
	client := kubeutils.NewKubeClient(masterUrl, configPath)
	info, err := monitor.GetStoreSize(ctx)
	if err != nil {
		glog.Errorf("get storage info error, err=%+v", err)
	}
	glog.V(4).Infof("storage info is %+v", info)
	sumSize := float64(0)
	count := 0
	for _, v := range info {
		log.Printf("name=%v,size=%v", v.Name, v.Size*1.0/float64(GB))
		sumSize += v.Size
		count++
	}
	//如果超过了总使用量超过了总容量的80%，扩容25%
	warnStorageSize := float64(int64(count)*int64(One)*GB) * 0.8
	if sumSize >= warnStorageSize {
		count, err := resize(client, namespace, name)
		if err != nil {
			glog.Errorf("resize count error,err=%+v", err)
		}
		glog.Infof("resize success,now is %+v", count)

	}
	glog.V(4).Infof("count should be %d", count)

}

func resize(client kubeutils.KubeClient, ns, name string) (int32, error) {
	cluster, err := client.Get(ns, name)
	if err != nil {
		glog.Errorf("get tidb cluster error, err=%+v", err)
		return 0, err
	}
	count := cluster.Spec.TiKV.Replicas
	count += int32(math.Ceil(float64(count) * float64(0.25)))
	cluster.Spec.TiKV.Replicas = count
	newCluster, err := client.Update(ns, name, cluster)
	return newCluster.Spec.TiKV.Replicas, nil
}

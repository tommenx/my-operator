package kubeutil

import (
	"github.com/golang/glog"
	"github.com/pingcap/tidb-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/tidb-operator/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type kubeClient struct {
	tidbClient *versioned.Clientset
}

type KubeClient interface {
	ScaleOutTiKV(ns, name string, replica int32) error
}

func NewKubeClient(path string) KubeClient {
	var cfg *rest.Config
	var err error
	if path != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			glog.Errorf("Failed to get cluster config with error: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			glog.Errorf("Failed to get cluster config with error: %v\n", err)
			os.Exit(1)
		}
	}
	client, err := versioned.NewForConfig(cfg)
	if err != nil {
		glog.Errorf("Failed to create client with error: %v\n", err)
		os.Exit(1)
	}
	return &kubeClient{client}
}

func (c *kubeClient) Get(ns, name string) (*v1alpha1.TidbCluster, error) {
	tidbCluster, err := c.tidbClient.PingcapV1alpha1().TidbClusters(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("get TiDB cluster error,err=%+v", err)
		return nil, err
	}
	return tidbCluster, nil
}
func (c *kubeClient) Update(ns, name string, cluster *v1alpha1.TidbCluster) (*v1alpha1.TidbCluster, error) {
	tidbCluster, err := c.tidbClient.PingcapV1alpha1().TidbClusters(ns).Update(cluster)
	return tidbCluster, err
}

func (c *kubeClient) ScaleOutTiKV(ns, name string, replica int32) error {
	tidbCluster, err := c.tidbClient.PingcapV1alpha1().TidbClusters(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	origin := tidbCluster.Spec.TiKV.Replicas
	//if origin > replica {
	//	glog.Errorf("origin replica %d, now %d, abort it", origin, replica)
	//	return nil
	//}
	count := origin + replica

	tidbCluster.Spec.TiKV.Replicas = count
	_, err = c.tidbClient.PingcapV1alpha1().TidbClusters(ns).Update(tidbCluster)
	glog.Infof("scale out tikv from %d to %d\n", origin, count)
	return err
}

type fakeKubeClient struct {
	cli kubernetes.Interface
}

func NewFakeKubeClient(path string) KubeClient {
	cfg, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		glog.Errorf("Failed to get cluster config with error: %v\n", err)
		os.Exit(1)
	}
	cli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("failed to create Clientset: %v", err)
	}
	return &fakeKubeClient{cli: cli}
}

func (fake *fakeKubeClient) ScaleOutTiKV(ns, name string, replica int32) error {
	instance, err := fake.cli.AppsV1().StatefulSets(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("get statefulset error, err=%+v", err)
		return err
	}
	count := *instance.Spec.Replicas + replica
	instance.Spec.Replicas = &count
	_, err = fake.cli.AppsV1().StatefulSets(ns).Update(instance)
	if err != nil {
		glog.Errorf("update error, err=%+v", err)
		return err
	}
	return nil

}

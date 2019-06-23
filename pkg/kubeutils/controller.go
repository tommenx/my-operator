package kubeutils

import (
	"github.com/golang/glog"
	"github.com/pingcap/tidb-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/tidb-operator/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type kubeClient struct {
	tidbClient *versioned.Clientset
}

type KubeClient interface {
	Get(ns, name string) (*v1alpha1.TidbCluster, error)
	Update(ns, name string, cluster *v1alpha1.TidbCluster) (*v1alpha1.TidbCluster, error)
}

func NewKubeClient(masterUrl, path string) KubeClient {
	var cfg *rest.Config
	var err error
	if path != "" {
		cfg, err = clientcmd.BuildConfigFromFlags(masterUrl, path)
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

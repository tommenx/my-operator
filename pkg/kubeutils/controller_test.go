package kubeutils

import "testing"

func TestGet(t *testing.T) {
	materUrl := ""
	path := "/root/.kube/config"
	client := NewKubeClient(materUrl, path)
	tidbCluster, err := client.Get("tidb", "tidb-cluster")
	if err != nil {
		t.Errorf("err=%+v", err)
	}
	t.Logf("%+v", *tidbCluster)
}

package scaler

import (
	"github.com/golang/glog"
	"github.com/pingcap/tidb-operator/pkg/cdapi"
	"github.com/pingcap/tidb-operator/pkg/kubeutil"
)

type scaleController struct {
	cd       cdapi.CDClient
	kubeUtil kubeutil.KubeClient
}

type IsolationLimit struct {
	ReadLimit  string
	WriteLimit string
}

type ScaleController interface {
	ScaleOut(ns, name string, count int32) error
	ScaleUpOne(ns string, limit map[string]*IsolationLimit) error
	ScaleUpAll(ns, tag, val string, limit *IsolationLimit) error
}

func NewScaleController(cdClient cdapi.CDClient, kubeUtil kubeutil.KubeClient) ScaleController {

	return &scaleController{
		cd:       cdClient,
		kubeUtil: kubeUtil,
	}
}

func (s *scaleController) ScaleOut(ns, name string, count int32) error {
	return s.kubeUtil.ScaleOutTiKV(ns, name, count)
}

func (s *scaleController) ScaleUpOne(ns string, limit map[string]*IsolationLimit) error {
	instances := make([]*cdapi.Instance, 0)
	for name, isolation := range limit {
		instances = append(instances, &cdapi.Instance{
			Name:  name,
			Read:  isolation.ReadLimit,
			Write: isolation.WriteLimit,
		})
	}
	err := s.cd.SetOneLimit(ns, instances)
	if err != nil {
		glog.Errorf("set one pod isolation error, err=%+v", err)
		return err
	}
	glog.Infof("scale up one pod success")
	return nil
}

func (s *scaleController) ScaleUpAll(ns, tag, val string, limit *IsolationLimit) error {
	err := s.cd.SetBatchLimit(ns, tag, val, limit.ReadLimit, limit.WriteLimit)
	if err != nil {
		glog.Errorf("scale up %s:%s error, err=%+v", tag, val)
		return err
	}
	glog.Infof("scale up batch pod success")
	return nil
}

type AutoScale interface {
	Run(stopCh chan struct{})
}

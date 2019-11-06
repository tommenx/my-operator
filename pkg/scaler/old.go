package scaler

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/pingcap/tidb-operator/pkg/cdapi"
	"github.com/pingcap/tidb-operator/pkg/store"
	"math"
	"strconv"
	"time"
)

type State int32

// key /storage/show/old
// val resourceAllocation - resourceTime
var (
	upBound                  = 25
	lowBound                 = 25
	epoch                    = 10
	defaultPeriod            = time.Second * time.Duration(epoch)
	checkBalanceWindow       = 6
	checkScaleOutWindow      = 3
	prefixResourceTime       = "/storage/show/old/resourceTime"
	prefixResourceAllocation = "/storage/show/old/resourceAllocation"
)

const (
	ACTIVE State = 0
	FROZEN State = 1
)

//resourceTime 资源时，当前使用资源 + 资源分配量 * 时间
//resourceAllocation 当前资源的分配量
type oldAutoScaler struct {
	checkTicker        *time.Ticker
	state              State
	frozenOp           int
	activeOp           int
	step               int32
	resourceTime       int
	resourceAllocation int
	resourceLimit      int
	replica            int32
	cd                 cdapi.CDClient
	scaler             ScaleController
	db                 store.DB
}

func NewOldAutoScale(cd cdapi.CDClient, scaler ScaleController, db store.DB, step int32, replica int32, resourceLimit int) AutoScale {
	prefixResourceTime = "/storage/show/old/resourceTime"
	prefixResourceAllocation = "/storage/show/old/resourceAllocation"
	return &oldAutoScaler{
		checkTicker:        time.NewTicker(defaultPeriod),
		state:              ACTIVE,
		frozenOp:           0,
		activeOp:           0,
		step:               step,
		scaler:             scaler,
		cd:                 cd,
		replica:            replica,
		resourceLimit:      resourceLimit,
		resourceAllocation: resourceLimit * int(replica),
		db:                 db,
	}
}

func (s *oldAutoScaler) Run(stopCh chan struct{}) {
	for {
		select {
		case <-s.checkTicker.C:
			go s.CheckAndScale()
			go s.ReportResource()
		case <-stopCh:
			break
		}
	}
}

func (s *oldAutoScaler) CheckAndScale() {
	//进入函数计算资源时
	//每次横向扩容的时候修改replica的数量
	s.resourceTime += int(s.replica) * s.resourceLimit * epoch
	instances, err := s.cd.GetTiKVStatus()
	if err != nil {
		glog.Errorf("get tikv status error")
	}
	write := float64(0)
	read := float64(0)
	stateStr := "ACTIVE"
	if s.state == 1 {
		stateStr = "FROZEN"
	}
	glog.Infof("-----------------%s:[%d]-----------------", stateStr, s.frozenOp)
	for _, instance := range instances {
		r, _ := strconv.ParseFloat(instance.Read, 64)
		w, _ := strconv.ParseFloat(instance.Write, 64)
		if w > 1000 || r > 1000 {
			break
		}
		write += w
		read += r
		glog.Infof("%s read:%s write:%s\n", instance.Name, instance.Read, instance.Write)
	}
	resourceTime := math.Ceil(float64(s.resourceTime) / 1024)
	glog.Infof("allocation: %d MB/s, resource-time: %.1f GB", s.resourceAllocation, resourceTime)
	avgwrite := int(write) / len(instances)
	avgread := int(read) / len(instances)
	glog.Infof("avg read:%d write:%d\n", avgread, avgwrite)
	if s.state == FROZEN {
		//如果平均读写都小于lowBound
		if avgread <= lowBound && avgwrite <= lowBound {
			s.frozenOp++
		} else {
			//否则计数归零
			s.frozenOp = 0
		}
		if s.frozenOp > checkBalanceWindow && s.CheckRegionBalance() {
			glog.Infof("frozen operation is count is %d, state change to ACTIVE", s.frozenOp)
			s.state = ACTIVE
			s.frozenOp = 0
		} else {
			s.frozenOp = 0
		}
		return
	}
	if avgwrite > upBound || avgread > upBound {
		s.activeOp++
		if s.activeOp >= checkScaleOutWindow {
			glog.Infof("got to warning bound, start to scale out, write:%d,read=%d\n", avgwrite, avgread)
			err = s.scaler.ScaleOut("default", "tidb-cluster", s.step)
			if err != nil {
				glog.Errorf("scale out error, err=%+v", err)
				return
			}
			s.activeOp = 0
			s.state = FROZEN
			s.replica += s.step
		}
	} else {
		s.frozenOp = 0
	}
	s.resourceAllocation = s.resourceLimit * int(s.replica)
}

func (s *oldAutoScaler) CheckRegionBalance() bool {
	regions, err := cdapi.GetRegionStatus()
	if err != nil {
		glog.Errorf("get store region count error, err=%+v", err)
		return false
	}
	avgRegionCount := 0
	for _, count := range regions {
		avgRegionCount += count
	}
	avgRegionCount /= len(regions)
	diffCount := int(math.Ceil(float64(avgRegionCount) * 0.1))
	for _, count := range regions {
		if int(math.Abs(float64(count-avgRegionCount))) > diffCount {
			fmt.Printf("one count is %d, avg count is %d\n", count, avgRegionCount)
			return false
		}
	}
	return true
}

func (s *oldAutoScaler) ReportResource() {
	//单位为MB/秒
	resourceAllocation := math.Ceil(float64(s.resourceAllocation))
	err := s.db.Put(context.Background(), prefixResourceAllocation, strconv.Itoa(int(resourceAllocation)))
	if err != nil {
		glog.Errorf("etcd put %s error,err=%+v", prefixResourceAllocation, err)
	}
	//单位为GB
	resourceTime := math.Ceil(float64(s.resourceTime) / 1024)
	err = s.db.Put(context.Background(), prefixResourceTime, strconv.Itoa(int(resourceTime)))
	if err != nil {
		glog.Errorf("etcd put %s error,err=%+v", prefixResourceAllocation, err)
	}
	return
}

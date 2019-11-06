package scaler

import (
	"container/list"
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/pingcap/tidb-operator/pkg/cdapi"
	"github.com/pingcap/tidb-operator/pkg/store"
	"math"
	"strconv"
	"time"
)

var (
	reserveBW         = 20
	upBWFactor        = 3.0
	upBWCheckFactor   = 2.0
	downBWCheckFactor = 4.0
	downBWFactor      = 1.5
	reserveBWLimit    = 120
	scaleBWLimit      = 120
	round             = 3
	historyOp         = 5
)

type fancyAutoScaler struct {
	checkTicker        *time.Ticker
	state              State
	frozenOp           int
	activeOp           int
	step               int32
	resourceTime       int
	resourceAllocation int
	resourceLimit      int
	curLimit           float64
	replica            int32
	cd                 cdapi.CDClient
	scaler             ScaleController
	db                 store.DB
	history            *list.List
}

func NewFancyAutoScale(cd cdapi.CDClient, scaler ScaleController, db store.DB, step int32, replica int32, resourceLimit int) AutoScale {
	prefixResourceTime = "/storage/show/fancy/resourceTime"
	prefixResourceAllocation = "/storage/show/fancy/resourceAllocation"
	lowBound = 30
	return &fancyAutoScaler{
		checkTicker:        time.NewTicker(defaultPeriod),
		state:              ACTIVE,
		frozenOp:           0,
		activeOp:           0,
		step:               step,
		scaler:             scaler,
		cd:                 cd,
		replica:            replica,
		resourceLimit:      reserveBW,
		curLimit:           float64(reserveBW),
		resourceAllocation: reserveBW * int(replica),
		db:                 db,
		history:            list.New(),
	}
}

func (s *fancyAutoScaler) Run(stopCh chan struct{}) {
	for {
		select {
		case <-s.checkTicker.C:
			go s.checkAndScale()
			go s.reportResource()
		case <-stopCh:
			break
		}
	}
}

func (s *fancyAutoScaler) addHistory(val float64) {
	s.history.PushBack(val)
	if s.history.Len() > historyOp {
		s.history.Remove(s.history.Front())
	}
}

func (s *fancyAutoScaler) getAvgHistory() float64 {
	sum := float64(0)
	for e := s.history.Front(); e != nil; e = e.Next() {
		switch e.Value.(type) {
		case float64:
			sum += e.Value.(float64)
		}
	}
	return sum / float64(s.history.Len())
}

func (s *fancyAutoScaler) checkAndScale() {
	resourceTime := math.Ceil(float64(s.resourceTime) / 1024)
	s.resourceTime += int(s.replica) * s.resourceLimit * epoch
	avgWrite, avgRead, _ := s.getAvgBandwidth()
	glog.Infof("avg-write %.1f, avg-read` %.1f allocation %d resource-time %d", avgWrite, avgRead, s.resourceAllocation, int(resourceTime))
	maxOfOne := math.Max(avgWrite, avgRead)
	s.addHistory(maxOfOne)
	stateStr := "ACTIVE"
	if s.state == 1 {
		stateStr = "FROZEN"
	}
	avg := s.getAvgHistory()
	glog.Infof("get history val %.1f. BW_Limit: %.1f. Tolerated Region:[%.1f <--> %.1f]", avg, s.curLimit, s.curLimit/downBWCheckFactor, s.curLimit/upBWCheckFactor)
	if s.state == ACTIVE {
		if avg*upBWCheckFactor > s.curLimit {
			//到达纵向扩容临界点
			if avg*upBWCheckFactor >= float64(reserveBWLimit) {
				glog.Infof("state is %s, max BW is %.1f, scale out", stateStr, avg)
				//go s.scaleOut()
				s.state = FROZEN
			} else {
				upBW := avg * upBWFactor
				s.scaleUpAll(int(upBW))
				s.curLimit = upBW
				glog.Infof("state is %s, max BW is %.1f, scale up to %d", stateStr, avg, int(upBW))
			}
		} else if avg*downBWCheckFactor < s.curLimit {
			downBW := avg * downBWFactor
			s.scaleUpAll(int(downBW))
			s.curLimit = downBW
			glog.Infof("state is %s, max BW is %.1f, scale down to %d", stateStr, avg, int(downBW))
		}
	} else if s.state == FROZEN {
		if avg <= float64(lowBound) {
			s.frozenOp++
		} else {
			s.frozenOp = 0
		}
		if s.frozenOp > checkBalanceWindow && s.checkRegionBalance() {
			glog.Infof("frozen operation is count is %d, state change to ACTIVE", s.frozenOp)
			s.state = ACTIVE
			s.frozenOp = 0
		} else {
			s.frozenOp = 0
		}
	}
	s.resourceAllocation = s.resourceLimit * int(s.replica)

}

//write read error
func (s *fancyAutoScaler) getAvgBandwidth() (float64, float64, error) {
	instances, err := s.cd.GetTiKVStatus()
	if err != nil {
		return 0, 0, err
	}
	sumWrite := float64(0)
	sumRead := float64(0)
	for _, instance := range instances {
		r, _ := strconv.ParseFloat(instance.Read, 64)
		w, _ := strconv.ParseFloat(instance.Write, 64)
		if w > 1000 || r > 1000 {
			break
		}
		sumWrite += w
		sumRead += r
	}
	count := float64(len(instances))
	avgWrite := sumWrite / count
	avgRead := sumRead / count
	return avgWrite, avgRead, nil
}

func (s *fancyAutoScaler) scaleUpAll(num int) {
	ns := "default"
	key := "app.kubernetes.io/component"
	val := "tikv"
	read := strconv.Itoa(num)
	write := strconv.Itoa(num)
	err := s.scaler.ScaleUpAll(ns, key, val, &IsolationLimit{read, write})
	if err != nil {
		glog.Errorf("scale up all %s:%s error, err=%+v", err)
	}
	s.resourceLimit = num
}
func (s *fancyAutoScaler) scaleOut() {
	ns := "default"
	name := "tidb-cluster"
	err := s.scaler.ScaleOut(ns, name, s.step)
	if err != nil {
		glog.Errorf("scale out error, err=%+v", err)
	}
	//调用扩容命令，需要延迟5秒，等新的实例启动之后，设置限速值
	time.Sleep(5 * time.Second)
	s.scaleUpAll(scaleBWLimit)
	s.replica += s.step
	s.resourceLimit = scaleBWLimit
	glog.Infof("scale out and up success, replica is %d BW limit is %d", s.replica, s.resourceLimit)
}

func (s *fancyAutoScaler) checkRegionBalance() bool {
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

func (s *fancyAutoScaler) reportResource() {
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

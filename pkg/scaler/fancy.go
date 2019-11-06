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
	downBWCheckFactor = 3.5
	downBWFactor      = 1.5
	reserveBWLimit    = 30
	scaleBWLimit      = 120
	round             = 3
	downHistoryOp     = 5
	upHistoryOp       = 3
)

type fancyAutoScaler struct {
	checkTicker        *time.Ticker
	state              State
	frozenOp           int
	activeOp           int
	step               int32
	resourceTime       int
	resourceAllocation int
	curLimit           float64
	replica            int32
	cd                 cdapi.CDClient
	scaler             ScaleController
	db                 store.DB
	upHistory          *list.List
	downHistory        *list.List
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
		curLimit:           float64(reserveBW),
		resourceAllocation: reserveBW * int(replica),
		db:                 db,
		upHistory:          list.New(),
		downHistory:        list.New(),
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

func (s *fancyAutoScaler) addUpHistory(val float64) {
	s.upHistory.PushBack(val)
	if s.upHistory.Len() > upHistoryOp {
		s.upHistory.Remove(s.upHistory.Front())
	}
}

func (s *fancyAutoScaler) addDownHistory(val float64) {
	s.downHistory.PushBack(val)
	if s.downHistory.Len() > downHistoryOp {
		s.downHistory.Remove(s.downHistory.Front())
	}
}

func (s *fancyAutoScaler) getUpAvgHistory() float64 {
	sum := float64(0)
	for e := s.upHistory.Front(); e != nil; e = e.Next() {
		switch e.Value.(type) {
		case float64:
			sum += e.Value.(float64)
		}
	}
	return sum / float64(s.upHistory.Len())
}

func (s *fancyAutoScaler) getDownAvgHistory() float64 {
	sum := float64(0)
	for e := s.downHistory.Front(); e != nil; e = e.Next() {
		switch e.Value.(type) {
		case float64:
			sum += e.Value.(float64)
		}
	}
	return sum / float64(s.downHistory.Len())
}

func (s *fancyAutoScaler) checkAndScale() {
	resourceTime := math.Ceil(float64(s.resourceTime) / 1024)
	s.resourceTime += int(s.replica) * int(s.curLimit) * epoch
	avgWrite, avgRead, _ := s.getAvgBandwidth()
	glog.Infof("avg-write %.1f, avg-read` %.1f allocation %d resource-time %d", avgWrite, avgRead, s.resourceAllocation, int(resourceTime))
	maxOfOne := math.Max(avgWrite, avgRead)
	s.addUpHistory(maxOfOne)
	s.addDownHistory(maxOfOne)
	stateStr := "ACTIVE"
	if s.state == 1 {
		stateStr = "FROZEN"
	}
	avgUp := s.getUpAvgHistory()
	avgDown := s.getDownAvgHistory()
	glog.Infof("Avg Up: %.1f, Avg Down: %.1f. BW_Limit: %.1f. Tolerated Region:[%.1f <--> %.1f]", avgUp, avgDown, s.curLimit, s.curLimit/downBWCheckFactor, s.curLimit/upBWCheckFactor)
	if s.state == ACTIVE {
		if avgDown >= float64(reserveBWLimit) {
			// Scale OUT
			glog.Infof("state is %s, max BW is %.1f, scale out", stateStr, avgUp)
			// sleep 20s,等待新的实例启动
			s.state = FROZEN

			upBWFactor = 4.0
			upBWCheckFactor = 3.5
			downBWCheckFactor = 6
			downBWFactor = 1.5

			s.scaleOut()
		}
	} else if s.state == FROZEN {
		glog.Infof("avg-up is %.1f, low bound is %d", avgUp, lowBound)
		if avgUp <= float64(lowBound) {
			s.frozenOp += 1
			glog.Infof("get to low bound, op count is %d", s.frozenOp)
		} else {
			s.frozenOp = 0
		}
		//连续三次上限低于
		if s.frozenOp > 3 && s.checkRegionBalance() {
			glog.Infof("frozen operation is count is %d, state change to ACTIVE", s.frozenOp)
			s.state = ACTIVE

			upBWFactor = 3.0
			upBWCheckFactor = 2.0
			downBWCheckFactor = 3.5
			downBWFactor = 1.5
			s.frozenOp = 0
		} else if s.frozenOp > 3 && !s.checkRegionBalance() {
			s.frozenOp = 0
		}
		glog.Infof("state is FROZEN,op is %d", s.frozenOp)
	}

	if avgUp*upBWCheckFactor > s.curLimit {
		//到达纵向扩容临界点 Scale UP
		upBW := avgUp * upBWFactor
		if upBW > float64(scaleBWLimit) {
			upBW = float64(scaleBWLimit)
		}
		s.scaleUpAll(int(upBW))
		s.curLimit = upBW
		glog.Infof("state is %s, max BW is %.1f, scale up to %d", stateStr, avgUp, int(upBW))
	} else if avgDown*downBWCheckFactor < s.curLimit {
		// Scale Down
		downBW := avgDown * downBWFactor
		s.scaleUpAll(int(downBW))
		s.curLimit = downBW
		glog.Infof("state is %s, max BW is %.1f, scale down to %d", stateStr, avgDown, int(downBW))
	}

	s.resourceAllocation = int(s.curLimit) * int(s.replica)

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
}
func (s *fancyAutoScaler) scaleOut() {
	//TODO 暂时不加数
	ns := "default"
	name := "tidb-cluster"
	err := s.scaler.ScaleOut(ns, name, s.step)
	if err != nil {
		glog.Errorf("scale out error, err=%+v", err)
	}
	//调用扩容命令，需要延迟20秒，等新的实例启动之后，设置限速值
	time.Sleep(8 * time.Second)
	s.replica += s.step

	s.scaleUpAll(int(s.curLimit))
	glog.Infof("scale out and up success, replica is %d BW limit is %.1f", s.replica, s.curLimit)
	glog.Infof("scaling OUT..., replica is to %d", s.replica)
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

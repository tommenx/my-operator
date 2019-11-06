package cdapi

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"os/exec"
)

//type pdcli struct {
//
//}

type Store struct {
	Id      int    `json:"id"`
	Address string `json:"address"`
	Labels  []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"labels"`
	Version   string `json:"version"`
	StateName string `json:"state_name"`
}

type Status struct {
	Capacity    string `json:"capacity"`
	Available   string `json:"available"`
	LeaderCount int    `json:"leader_count"`
	RegionCount int    `json:"region_count"`
}

type StoreStatus struct {
	Count  int `json:"count"`
	Stores []struct {
		Store  Store  `json:"store"`
		Status Status `json:"status"`
	} `json:"stores"`
}

func GetRegionStatus() (map[int]int, error) {
	cmd := "./get_region_count.sh"
	data, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		glog.Errorf("get store region count error, err=%+v", err)
		return nil, err
	}
	return ParseData(data)
}

func ParseData(data []byte) (map[int]int, error) {
	storeStatus := &StoreStatus{}
	err := json.Unmarshal(data, storeStatus)
	if err != nil {
		glog.Errorf("parse data error,err=%+v", err)
		return nil, err
	}
	res := make(map[int]int)
	for _, store := range storeStatus.Stores {
		res[store.Store.Id] = store.Status.RegionCount
	}
	fmt.Printf("%+v", res)
	return res, nil
}

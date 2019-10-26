package cdapi

import (
	"testing"
	"time"
)

func TestCdClient_GetTiKVStatus(t *testing.T) {
	cd := NewCDClient("http://10.77.30.147:8888", time.Second*5)
	//GET TIKV STATUS
	status, err := cd.GetTiKVStatus()
	if err != nil {
		t.Errorf("get tikve status error, err=%+v", err)
		return
	}
	for _, one := range status {
		t.Logf("%+v", *one)
	}
	////set batch
	//err = cd.SetBatchLimit("default", "app", "ubuntu", "25", "25")
	//if err != nil {
	//	t.Errorf("set batch limit error, err=%+v", err)
	//}
}

func TestCdClient_SetOneLimit(t *testing.T) {
	cd := NewCDClient("http://10.77.30.147:8888", time.Second*5)
	instances := []*Instance{
		{"ubuntu-0", "10", "20"},
		{"ubuntu-1", "20", "30"},
	}
	err := cd.SetOneLimit("default", instances)
	if err != nil {
		t.Errorf("set one instance limit error, err=%+v", err)
	}
}

func TestCdClient_SetBatchLimit(t *testing.T) {
	cd := NewCDClient("http://10.77.30.147:8888", time.Second*5)
	err := cd.SetBatchLimit("default", "app", "ubuntu", "25", "25")
	if err != nil {
		t.Errorf("set batch limit error, err=%+v", err)
	}
}

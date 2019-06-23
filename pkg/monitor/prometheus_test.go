package monitor

import (
	"context"
	"testing"
)

//
func TestGetSource(t *testing.T) {
	address := "http://10.77.110.140:30764"
	monitor := NewMonitor(address)
	ctx := context.Background()
	list, err := monitor.GetStoreSize(ctx)
	if err != nil {
		t.Errorf("err=%+v", err)
	}
	t.Logf("list is %+v", list)
}

//func TestType(t *testing.T) {
//	vector := model.Vector{}
//	var val model.Value
//	val = vector
//	p := val.(model.Vector)
//	log.Println(p.Type())
//	log.Printf("%v", val.Type())
//}

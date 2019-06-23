package monitor

import (
	"golang.org/x/net/context"
	"testing"
)

func TestGetSource(t *testing.T){
	address := "http://10.77.110.140:30764"
	monitor := NewMonitor(address)
	ctx := context.Background()
	monitor.GetStoreSize(ctx)
}
package monitor

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"time"
)

type monitor struct {
	address string
	api     v1.API
}

type Storage struct {
	Name string
	Size float64
}

type Monitor interface {
	GetStoreSize(ctx context.Context) ([]*Storage, error)
}

func NewMonitor(address string) Monitor {
	cfg := api.Config{
		Address:      address,
		RoundTripper: api.DefaultRoundTripper,
	}
	client, err := api.NewClient(cfg)
	if err != nil {
		glog.Errorf("create prometheus client error,err=%+v", err)
		panic(err)
	}
	api := v1.NewAPI(client)
	return &monitor{
		address: address,
		api:     api,
	}
}

func (m *monitor) GetStoreSize(ctx context.Context) ([]*Storage, error) {
	queryStr := "sum(tikv_engine_size_bytes) by (instance)"
	storeSize, _, err := m.api.Query(ctx, queryStr, time.Now())
	if err != nil {
		glog.Errorf("err is %+v", err)
		return nil, err
	}
	//val, ok := storeSize.
	//if !ok {
	//	glog.Errorf("parse store size error")
	//	return nil, errors.New("store size is not vector type")
	//}
	//list := []*Storage{}
	//for _, v := range val {
	//	key := model.LabelName("instance")
	//	list = append(list, &Storage{
	//		Name: string(v.Metric[key]),
	//		Size: float64(v.Value),
	//	})
	//}
	fmt.Println(storeSize)
	return nil , nil
}

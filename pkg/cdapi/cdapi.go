package cdapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pingcap/tidb-operator/pkg/httputil"
	"io/ioutil"
	"net/http"
	"time"
)

type cdClient struct {
	url        string
	httpClient *http.Client
}

func NewCDClient(url string, timeout time.Duration) CDClient {
	httpClient := &http.Client{Timeout: timeout}
	return &cdClient{
		url:        url,
		httpClient: httpClient,
	}
}

type Instance struct {
	Name  string `json:"name"`
	Read  string `json:"read"`
	Write string `json:"write"`
}

type SetOnePodArgs struct {
	Namespace string      `json:"namespace"`
	Requests  []*Instance `json:"requests"`
}

type SetBatchPodArgs struct {
	Tag   string `json:"tag"`
	Val   string `json:"val"`
	Read  string `json:"read"`
	Write string `json:"write"`
}

type TiKVStatus struct {
	Instances []*Instance `json:"instances"`
	Code      int32       `json:"code"`
	Message   string      `json:"message"`
}

type SetLimitResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type CDClient interface {
	GetTiKVStatus() ([]*Instance, error)
	SetOneLimit(namespace string, instances []*Instance) error
	SetBatchLimit(namespace, tag, val, read, write string) error
}

var (
	tikvStatusPrefix    = "/util"
	setOneLimitPrefix   = "/setonepod"
	setBatchLimitPrefix = "/setbatchpod"
)

func (cc *cdClient) GetTiKVStatus() ([]*Instance, error) {
	apiURL := fmt.Sprintf("%s%s", cc.url, tikvStatusPrefix)
	body, err := httputil.GetBodyOK(cc.httpClient, apiURL)
	if err != nil {
		return nil, err
	}
	status := &TiKVStatus{}
	err = json.Unmarshal(body, status)
	if err != nil {
		return nil, err
	}
	return status.Instances, nil
}

func (cc *cdClient) SetOneLimit(namespace string, instances []*Instance) error {
	apiURL := fmt.Sprintf("%s%s", cc.url, setOneLimitPrefix)
	data := &SetOnePodArgs{
		Namespace: namespace,
		Requests:  instances,
	}
	reqData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(reqData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := cc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer httputil.DeferClose(req.Body)
	body, err := ioutil.ReadAll(res.Body)
	setLimitResponse := &SetLimitResponse{}
	err = json.Unmarshal(body, setLimitResponse)
	if err != nil || setLimitResponse.Code != 0 {
		return fmt.Errorf("set one instance limit error")
	}
	return nil
}

func (cc *cdClient) SetBatchLimit(namespace, tag, val, read, write string) error {
	apiURL := fmt.Sprintf("%s%s", cc.url, setBatchLimitPrefix)
	data := &SetBatchPodArgs{
		Tag:   tag,
		Val:   val,
		Read:  read,
		Write: write,
	}
	reqData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(reqData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := cc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer httputil.DeferClose(req.Body)
	body, err := ioutil.ReadAll(res.Body)
	setLimitResponse := &SetLimitResponse{}
	err = json.Unmarshal(body, setLimitResponse)
	if err != nil || setLimitResponse.Code != 0 {
		return fmt.Errorf("set batch instance limit error,err=%+v,msg=%+s", err, setLimitResponse.Message)
	}
	return nil
}

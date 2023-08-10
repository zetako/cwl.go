package slex

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/lijiang2014/cwl.go/intergration/client"
	"net/http"
	"starlight/common/httpclient"
	"strings"
)

const (
	testLoginAPI string = "http://uat.starlight-dev.nscc-gz.cn/api/keystone/short_term_token/name"
	testUsername string = "nscc-gz_yfb_2"
	testPassword string = "UHcyMDIyUkQ="
	testBaseDir  string = "/GPUFS/nscc-gz_yfb_2/"
)

var (
	globalExecutor *Executor
	testAllocModel *JobAllocationModel = &JobAllocationModel{
		Default: &SingleJobAllocationModel{
			Cluster:   "k8s_uat",
			Partition: "ln15",
			Cpu:       getIntPointer(1),
			Gpu:       getIntPointer(0),
			Memory:    getIntPointer(4 * 1024),
			WorkDir:   client.Volume{},
		},
		Diff: map[string]*SingleJobAllocationModel{},
	}
)

func getIntPointer(i int) *int {
	return &i
}

func getToken() (string, error) {
	jsonBody := fmt.Sprintf("{\"username\":\"%v\",\"password\":\"%v\"}", testUsername, testPassword)
	resp, err := http.Post(testLoginAPI, "application/json;charset=UTF-8", strings.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http code not 200, but %d", resp.StatusCode)
	}
	var token string
	_, err = httpclient.GetSpecResponse(resp.Body, &token)
	if err != nil {
		return "", err
	}
	return token, nil
}

func init() {
	// 1. get client
	ctx := context.TODO()
	c, err := client.New(ctx, client.Config{
		Username: testUsername,
		Password: testPassword,
		BaseURL:  "http://uat.starlight-dev.nscc-gz.cn",
	})
	if err != nil {
		panic(err)
	}
	// 2. get slex
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	globalExecutor, err = New(ctx, id.String(), c, testUsername, testAllocModel, client.BaseDir{Default: testBaseDir})
	if err != nil {
		panic(err)
	}
}

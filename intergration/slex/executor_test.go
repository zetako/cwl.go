package slex

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"starlight/common/httpclient"
	"starlight/common/model"
	"strings"
)

const (
	testLoginAPI string = "http://uat.starlight-dev.nscc-gz.cn/api/keystone/short_term_token/name"
	testUsername string = "nscc-gz_yfb_2"
	testPassword string = "UHcyMDIyUkQ="
)

var (
	globalExecutor *StarlightExecutor
	testAllocModel *JobAllocationModel = &JobAllocationModel{
		Default: &SingleJobAllocationModel{
			Cluster:   "k8s_uat",
			Partition: "ln15",
			Cpu:       getIntPointer(1),
			Gpu:       getIntPointer(0),
			Memory:    getIntPointer(4 * 1024),
			WorkDir:   model.Volume{},
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
	// 1. get token
	token, err := getToken()
	if err != nil {
		panic(err)
	}
	// 2. get slex
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	globalExecutor, err = New(context.TODO(), id.String(), token, testUsername, testAllocModel)
	if err != nil {
		panic(err)
	}
}

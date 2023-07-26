package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	LoginAPI = "/api/keystone/short_term_token/name"

	DefaultBaseURL     = "starlight.nscc-gz.cn"
	DefaultTimeout     = 10
	DefaultDialTimeout = 5
	DefaultRetry       = 3
	DefaultContentType = "text/plain"
)

var (
	ErrorNoToken = fmt.Errorf("token or username/password is not provided")
)

// StarlightClient 星光简单http客户端
type StarlightClient struct {
	ctx         context.Context // 上下文
	token       string          // Bihu-Token
	baseURL     string          // 星光服务基础URL
	contentType string          // 请求的Content-Type字段
	timeout     int             // 响应超时限制
	dialTimeout int             // 链接超时限制
	retry       int             // 最大重试次数
}

// StarlightClientConfig 星光http客户端的配置
type StarlightClientConfig struct {
	Token string `yaml:"token" json:"token"` // Bihu-Token
	// 👇 如果不设置Token,需要这两个
	Username string `yaml:"username" json:"username"` // 用户名
	Password string `yaml:"password" json:"password"` // 密码

	BaseURL     string `yaml:"base_url" json:"base_url"`         // 基础URL
	Timeout     int    `yaml:"timeout" json:"timeout"`           // 响应超时限制
	DialTimeout int    `yaml:"dial_timeout" json:"dial_timeout"` // 响应超时限制
	Retry       int    `yaml:"retry" json:"retry"`               // 最大重试次数

	ContentType string `yaml:"content_type" json:"content_type"` // 数据类型
}

// SetDefault 填充默认配置
func (conf *StarlightClientConfig) SetDefault() error {
	// 1. 基础信息
	if conf.BaseURL == "" {
		conf.BaseURL = DefaultBaseURL
	}
	if conf.Timeout == 0 {
		conf.Timeout = DefaultTimeout
	}
	if conf.DialTimeout == 0 {
		conf.DialTimeout = DefaultDialTimeout
	}
	if conf.Retry == 0 {
		conf.Retry = DefaultRetry
	}
	if conf.ContentType == "" {
		conf.ContentType = DefaultContentType
	}

	// 2. 登录信息
	if conf.Token != "" {
		if conf.Username == "" || conf.Password == "" {
			return ErrorNoToken
		}
		encodedPasswd := base64.StdEncoding.EncodeToString([]byte(conf.Password))
		jsonBody := fmt.Sprintf("{\"username\":\"%s\",\"password\":\"%s\"}", conf.Username, encodedPasswd)
		resp, err := http.Post(LoginAPI, "application/json;charset=UTF-8", strings.NewReader(jsonBody))
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("login request failed")
		}
		_, err = ResponseSpec(resp.Body, &conf.Token)
		if err != nil {
			return fmt.Errorf("login request resolve failed")
		}
	}
	return nil
}

func (c StarlightClient) PostSpec(api string, dataBean interface{}, specBean interface{}) (*ResponseWrap, error) {

}

func (c StarlightClient) GetSpec(api string, specBean interface{}) (*ResponseWrap, error) {

}

// ResponseSpec 获取请求中的Spec字段，并反序列化为传入的结构体
func ResponseSpec(reader io.Reader, specBean interface{}) (*ResponseWrap, error) {
	var wrap ResponseWrap
	err := json.NewDecoder(reader).Decode(&wrap)
	if err != nil {
		return nil, err
	}
	if wrap.Code != 200 {
		return nil, fmt.Errorf("got error code %d", wrap.Code)
	}
	if wrap.Spec != nil && specBean != nil {
		err = json.Unmarshal(wrap.Spec, specBean)
		if err != nil {
			return nil, err
		}
	}
	return &wrap, nil
}

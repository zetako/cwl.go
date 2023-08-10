package client

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// BaseDir 定义执行器的basedir
type BaseDir struct {
	Default string            `yaml:"default" json:"default"` // 默认的basedir
	Diff    map[string]string `yaml:"diff" json:"diff"`       // 特殊集群的basedir
}

func (b BaseDir) Get(partition string) string {
	if tmp, ok := b.Diff[partition]; ok {
		return tmp
	} else {
		return b.Default
	}
}

// Config 星光http客户端的配置
type Config struct {
	Token string `yaml:"token" json:"token"` // Bihu-Token
	// 👇 如果不设置Token,需要这两个
	Username string `yaml:"username" json:"username"` // 用户名
	Password string `yaml:"password" json:"password"` // 密码

	BaseURL     string `yaml:"base_url" json:"base_url"`         // 基础URL
	Timeout     int    `yaml:"timeout" json:"timeout"`           // 响应超时限制
	DialTimeout int    `yaml:"dial_timeout" json:"dial_timeout"` // 响应超时限制
	Retry       int    `yaml:"retry" json:"retry"`               // 最大重试次数

	ContentType string `yaml:"content_type" json:"content_type"` // 数据类型

	BaseDir BaseDir `yaml:"base_dir" json:"base_dir"` // 基础目录
}

// SetDefault 填充默认配置
func (conf *Config) SetDefault() error {
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
	if conf.Token == "" {
		if conf.Username == "" || conf.Password == "" {
			return ErrorNoToken
		}
		encodedPasswd := base64.StdEncoding.EncodeToString([]byte(conf.Password))
		jsonBody := fmt.Sprintf("{\"username\":\"%s\",\"password\":\"%s\"}", conf.Username, encodedPasswd)
		resp, err := http.Post(conf.BaseURL+LoginAPI, "application/json;charset=UTF-8", strings.NewReader(jsonBody))
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("login request failed")
		}
		_, err = GetSpecFromResponse(resp.Body, &conf.Token)
		if err != nil {
			return fmt.Errorf("login request resolve failed")
		}
	}
	return nil
}

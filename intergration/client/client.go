package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	LoginAPI = "/api/keystone/short_term_token/name"

	DefaultBaseURL     = "starlight.nscc-gz.cn"
	DefaultTimeout     = 10
	DefaultDialTimeout = 5
	DefaultRetry       = 3
	DefaultContentType = "application/json; charset=UTF-8"
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
	retry       int             // 最大重试次数，暂时没有被使用
}

func New(ctx context.Context, conf StarlightClientConfig) (*StarlightClient, error) {
	err := conf.SetDefault()
	if err != nil {
		return nil, err
	}

	return &StarlightClient{
		ctx:         ctx,
		token:       conf.Token,
		baseURL:     conf.BaseURL,
		contentType: conf.ContentType,
		timeout:     conf.Timeout,
		dialTimeout: conf.DialTimeout,
		retry:       conf.Retry,
	}, nil
}

// PostSpec 向星光服务发送Post请求并解码返回的Spec.
// Post的数据可以是字符串或者可以序列化的结构体
func (c *StarlightClient) PostSpec(api string, dataBean interface{}, specBean interface{}) (*ResponseWrap, error) {
	var (
		err error
		raw []byte
	)

	// 序列化数据
	switch data := dataBean.(type) {
	case string:
		raw = []byte(data)
	default:
		raw, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	// 发送请求
	resp, err := c.Request(api, http.MethodPost, raw)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http code: %d", resp.StatusCode)
	}
	return GetSpecFromResponse(resp.Body, specBean)
}

// GetSpec 向星光服务发送Get请求并解码返回的Spec.
func (c *StarlightClient) GetSpec(api string, specBean interface{}) (*ResponseWrap, error) {
	resp, err := c.Request(api, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http code: %d", resp.StatusCode)
	}
	return GetSpecFromResponse(resp.Body, specBean)
}

// Request 发送请求
func (c *StarlightClient) Request(url, method string, data []byte) (*http.Response, error) {
	// 检查上下文
	ddl, hasDDL := c.ctx.Deadline()
	if hasDDL {
		if ddl.Before(time.Now()) {
			return nil, fmt.Errorf("context deadline excced")
		}
	}
	// 请求URL
	if len(url) <= 0 {
		return nil, fmt.Errorf("no request url")
	}
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimPrefix(url, "api/")
	url = c.baseURL + "/api/" + url
	// 创建请求
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	// 附加请求头
	req.Header.Set("Content-Type", c.contentType)
	req.Header.Set("Bihu-Token", c.token)
	uuid := c.ctx.Value("UUID")
	if uuid != nil {
		req.Header.Set("UUID", uuid.(string))
	}
	// 创建客户端
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// 允许不安全的证书
				// FIXME: 其他更安全的方法
				InsecureSkipVerify: true,
			},
			DialContext: (&net.Dialer{
				// 设置连接超时
				Timeout: time.Duration(c.dialTimeout) * time.Second,
			}).DialContext,
		},
		// 设置通讯超时
		Timeout: time.Duration(c.timeout) * time.Second,
	}
	// 如果有ddl，重新设置
	if hasDDL {
		if ddl.Before(time.Now()) {
			return nil, fmt.Errorf("context deadline excced")
		}
		// 传递deadline
		req.Header.Set("DEADLINE", ddl.Format(time.RFC3339))
		// 设置请求整体的超时时间
		client.Timeout = time.Now().Sub(ddl)
	}
	// 正式连接
	return client.Do(req)
}

// GetSpecFromResponse 获取请求中的Spec字段，并反序列化为传入的结构体
func GetSpecFromResponse(reader io.Reader, specBean interface{}) (*ResponseWrap, error) {
	var wrap ResponseWrap
	err := json.NewDecoder(reader).Decode(&wrap)
	if err != nil {
		return nil, err
	}
	if wrap.Code != 200 {
		return nil, NewError(wrap.Code, wrap.Info)
	}
	if wrap.Spec != nil && specBean != nil {
		err = json.Unmarshal(wrap.Spec, specBean)
		if err != nil {
			return nil, err
		}
	}
	return &wrap, nil
}

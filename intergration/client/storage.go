package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type StarlightStorageClient struct {
	*StarlightClient
	// TODO 各请求接受Option
}

func (c *StarlightClient) StorageClient() *StarlightStorageClient {
	return &StarlightStorageClient{c}
}

func (c StarlightStorageClient) Download(location string) ([]byte, error) {
	// 发请求
	resp, err := c.Request("/storage/download?file="+location, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查类型
	if v := resp.Header.Get("Content-Type"); v != "application/octet-stream" {
		// 类型不对，应该返回了一个Spec包装的错误
		_, err = GetSpecFromResponse(resp.Body, nil)
		return nil, err
	}
	// 读取文件内容
	return ioutil.ReadAll(resp.Body)
}
func (c StarlightStorageClient) Upload(location string, data []byte, overwrite bool) (int64, error) {
	// 发请求
	params := url.Values{}
	params.Set("file", location)
	if overwrite {
		params.Set("overwrite", "true")
	}
	resp, err := c.Request("/storage/upload?"+params.Encode(), http.MethodPut, data)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	// 获取Spec
	spec := struct {
		File    string
		Written int64
	}{}
	_, err = GetSpecFromResponse(resp.Body, &spec)
	return spec.Written, err
}
func (c StarlightStorageClient) State(location string, checksum bool) (*FileInfo, error) {
	params := url.Values{}
	params.Set("file", location)
	if checksum {
		params.Set("check_sum", "true")
	}

	info := FileInfo{}
	_, err := c.GetSpec("/storage/state?"+params.Encode(), &info)
	return &info, err
}
func (c StarlightStorageClient) StateDir(location string, checksum bool) ([]FileInfo, error) {
	params := url.Values{}
	params.Set("file", location)
	params.Set("list", "true")
	if checksum {
		params.Set("check_sum", "true")
	}

	var infos []FileInfo
	_, err := c.GetSpec("/storage/state?"+params.Encode(), &infos)
	return infos, err
}
func (c StarlightStorageClient) Operation(operation, source, destination string, force, recursive bool, options ...[2]string) (err error) {
	// 参数
	params := url.Values{}
	params.Set("opt", operation)
	if source != "" {
		params.Set("from", source)
	}
	if destination != "" {
		params.Set("target", destination)
	}
	if force {
		params.Set("force", "true")
	}
	if recursive {
		params.Set("recursive", "true")
	}
	for _, option := range options {
		params.Add(option[0], option[1])
	}
	// 请求
	_, err = c.PostSpec("/storage/operation?"+params.Encode(), nil, nil)
	return err
}
func (c StarlightStorageClient) Copy(source, destination string, force, recursive bool) error {
	return c.Operation("cp", source, destination, force, recursive)
}
func (c StarlightStorageClient) Link(source, destination string, symbolic bool) error {
	if symbolic {
		return c.Operation("ln", source, destination, false, false)
	} else {
		return c.Operation("ln-hard", source, destination, false, false)
	}
}
func (c StarlightStorageClient) Mkdir(location string, mode os.FileMode) error {
	if mode != 0 {
		return c.Operation("mkdir", "", location, false, false, [2]string{"mod", fmt.Sprintf("0%o", mode)})
	} else {
		return c.Operation("mkdir", "", location, false, false)
	}
}
func (c StarlightStorageClient) Glob(base string, pattern string, checksum bool) ([]FileInfo, error) {
	params := url.Values{}
	params.Set("dir", base)
	params.Set("pattern", pattern)
	if checksum {
		params.Set("checksum", "true")
	}

	var infos []FileInfo
	_, err := c.GetSpec("/storage/glob?"+params.Encode(), &infos)
	return infos, err
}

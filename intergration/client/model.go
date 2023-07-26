package client

import "encoding/json"

type ResponseWrap struct {
	UUID  string          `json:"uuid"`  // 唯一标识
	Code  int             `json:"code"`  // 响应码
	Info  string          `json:"info"`  // 提示信息，必须用户友好
	Kind  string          `json:"kind"`  // 响应类型
	Total int             `json:"total"` // 符合条件的数据总数
	Spec  json.RawMessage `json:"spec"`  // 响应数据
}

type FileInfo struct {
	Name     string   `json:"name"`               // 文件名或者目录名
	Path     string   `json:"path"`               // 绝对路径
	Size     int64    `json:"size"`               // 文件大小
	Checksum string   `json:"checksum,omitempty"` // 校验和
	ACL      []string `json:"acl,omitempty"`      // ACL列表
	Type     FileType `json:"type"`               // 文件类型
	Perm     string   `json:"perm"`               // 权限
	Time     string   `json:"time"`               // 修改时间
	Uid      int      `json:"uid"`                // LDAP UID
	Gid      int      `json:"gid"`                // LDAP GID
	Target   string   `json:"target,omitempty"`   // 链接指向，只有符号链接文件才会有
}

type FileType int

const (
	FileTypeNotExist       FileType = -1
	FileTypeNormalFile     FileType = 0
	FileTypeDir            FileType = 1
	FileTypeSymlinkFile    FileType = 2
	FileTypeSymLinkDir     FileType = 3
	FileTypeSymLinkInvalid FileType = 4
)

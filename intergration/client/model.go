package client

import "encoding/json"

// ResponseWrap 星光返回值结构
type ResponseWrap struct {
	UUID  string          `json:"uuid"`  // 唯一标识
	Code  int             `json:"code"`  // 响应码
	Info  string          `json:"info"`  // 提示信息，必须用户友好
	Kind  string          `json:"kind"`  // 响应类型
	Total int             `json:"total"` // 符合条件的数据总数
	Spec  json.RawMessage `json:"spec"`  // 响应数据
}

// FileInfo 文件信息
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

// IsDir 返回该文件信息是否是一个文件夹；指向文件夹的链接不视为文件夹
func (info FileInfo) IsDir() bool {
	return info.Type == FileTypeDir
}

// FileType 文件类型
type FileType int

const (
	FileTypeNotExist       FileType = -1
	FileTypeNormalFile     FileType = 0
	FileTypeDir            FileType = 1
	FileTypeSymlinkFile    FileType = 2
	FileTypeSymLinkDir     FileType = 3
	FileTypeSymLinkInvalid FileType = 4
)

// Volume 卷挂载信息
type Volume struct {
	Name      string `yaml:"name" json:"name"`
	HostPath  string `json:"host_path" yaml:"host_path"`
	ReadOnly  bool   `yaml:"read_only" json:"read_only"`
	MountPath string `yaml:"mount_path" json:"mount_path"`
	Local     bool   `yaml:"local" json:"local"`
}

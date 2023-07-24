package sfs

import (
	"context"
	"github.com/lijiang2014/cwl.go"
	"os"
	"path"
	"starlight/common/code"
	"starlight/common/errors"
	"starlight/common/httpclient"
	"starlight/common/model"
	"strings"
)

func New(ctx context.Context, token string, workDir string) (*StarlightFileSystem, error) {
	b, err := httpclient.NewBihuClient(ctx, token)
	if err != nil {
		return nil, err
	}
	return &StarlightFileSystem{
		ctx:           ctx,
		token:         token,
		workDir:       workDir,
		baseClient:    b,
		storageClient: b.Storage(),
	}, nil
}

func FileInfo2CwlFile(info *model.FileInfo) cwl.File {
	file := cwl.File{
		Location: "file://" + info.Path,
		Path:     info.Path,
		Basename: path.Base(info.Path),
		Dirname:  path.Dir(info.Path),
		Nameroot: "", //TODO
		Nameext:  path.Ext(info.Path),
		Checksum: info.Checksum,
		Size:     info.Size,
	}

	// root = base - ext
	file.Nameroot = file.Basename[:len(file.Basename)-len(file.Nameext)]

	return file
}

func FileInfo2CwlDir(info *model.FileInfo) cwl.Directory {
	return cwl.Directory{
		Location: "file://" + info.Path,
		Path:     info.Path,
		Basename: path.Base(info.Path),
		Listing:  nil,
	}
}

func ErrorIsNotExist(err error) bool {
	if os.IsNotExist(err) {
		return true
	}
	if errors.MustCode(err) == code.FILE_NOT_EXIST {
		return true
	}
	return false
}

func GetRootFS(file string) string {
	file = path.Clean(file)
	parts := strings.Split(file, string(os.PathSeparator))
	for _, part := range parts {
		if part != "" {
			return part
		}
	}
	return ""
}

func (s StarlightFileSystem) canLink(src, dst string) bool {
	srcRoot := GetRootFS(src)
	if srcRoot == "" {
		return false
	}
	dstRoot := GetRootFS(dst)
	if dstRoot == "" {
		return false
	}
	if srcRoot == dstRoot {
		return true
	}
	return false
}

func (s StarlightFileSystem) getAbsPath(p string) string {
	p = strings.TrimPrefix(p, "file://")
	if path.IsAbs(p) {
		return p
	}
	return path.Join(s.workDir, p)
}

package sfs

import (
	"context"
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"net/url"
	"os"
	"starlight/common/httpclient"
	"starlight/common/model"
)

// StarlightFileSystem is implementation of runner.FileSystem, using starlight's storage.ms
type StarlightFileSystem struct {
	ctx           context.Context
	token         string
	workDir       string
	baseClient    *httpclient.BihuClient // we still need this for Glob
	storageClient *httpclient.StorageClient
}

func (s StarlightFileSystem) Create(path, contents string) (cwl.File, error) {
	path = s.getAbsPath(path)
	// UploadSimple + State
	_, err := s.storageClient.UploadSimple(path, []byte(contents), false)
	if err != nil {
		return cwl.File{}, err
	}
	fileInfo, err := s.storageClient.State(path)
	return FileInfo2CwlFile(fileInfo), nil
}

func (s StarlightFileSystem) Info(loc string) (cwl.File, error) {
	loc = s.getAbsPath(loc)
	info, err := s.storageClient.State(loc)
	return FileInfo2CwlFile(info), err
}

func (s StarlightFileSystem) DirInfo(loc string, deepLen int) (cwl.Directory, error) {
	loc = s.getAbsPath(loc)
	// 重复StateDirFiles

	// 本体
	dirInfo, err := s.storageClient.State(loc)
	if err != nil {
		return cwl.Directory{}, err
	}
	dir := FileInfo2CwlDir(dirInfo)

	// 列表
	if deepLen > 0 {

		dir.Listing = []cwl.FileDir{}
		fileInfos, err := s.storageClient.StateDirFiles(loc)
		if err != nil {
			return cwl.Directory{}, err
		}
		for _, info := range fileInfos {
			if info.IsDir() {
				subDir, err := s.DirInfo(info.Path, deepLen-1)
				if err != nil {
					return cwl.Directory{}, err
				}
				dir.Listing = append(dir.Listing, cwl.NewFileDir(subDir))
			} else {
				dir.Listing = append(dir.Listing, cwl.NewFileDir(FileInfo2CwlFile(&info)))
			}
		}
	}
	return dir, nil
}

func (s StarlightFileSystem) Copy(source, dest string) error {
	source = s.getAbsPath(source)
	dest = s.getAbsPath(dest)
	// Copy
	return s.storageClient.Copy(source, dest, false, true)
}

func (s StarlightFileSystem) Contents(loc string) (string, error) {
	loc = s.getAbsPath(loc)
	// DownloadSimple
	data, err := s.storageClient.DownloadSimple(loc)
	return string(data), err
}

func (s StarlightFileSystem) Glob(pattern string) ([]cwl.File, error) {
	// Not Match, "/storage/glob?%sdir=%s&pattern=%s"
	var (
		infos         []model.FileInfo
		files         []cwl.File
		checksumParam string
		query         string
	)
	checksumParam = "checksum=true&"
	pattern = url.QueryEscape(pattern)
	query = fmt.Sprintf("/storage/glob?%sdir=%s&pattern=%s", checksumParam, s.workDir, pattern)
	_, err := s.baseClient.GetSpec(query, &infos)
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		files = append(files, FileInfo2CwlFile(&info))
	}
	return files, nil
}

func (s StarlightFileSystem) EnsureDir(dir string, mode os.FileMode) error {
	dir = s.getAbsPath(dir)
	// State + MkdirOpt
	_, err := s.storageClient.State(dir)
	if err != nil {
		if !ErrorIsNotExist(err) {
			return err
		} else {
			return s.storageClient.Mkdir(dir, mode)
		}
	}
	return nil
}

func (s StarlightFileSystem) Migrate(source, dest string) (bool, error) {
	source = s.getAbsPath(source)
	dest = s.getAbsPath(dest)
	// Copy or Ln
	if s.canLink(source, dest) {
		return true, s.storageClient.Ln(source, dest, true)
	} else {
		return false, s.storageClient.Copy(source, dest, true, true)
	}
}

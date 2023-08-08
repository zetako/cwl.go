package sfs

import (
	"context"
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/intergration/client"
	"os"
)

// StarlightFileSystem is implementation of runner.FileSystem, using starlight's storage.ms
type StarlightFileSystem struct {
	ctx      context.Context
	checksum bool
	token    string
	workDir  string
	client   *client.StarlightStorageClient
}

func (s StarlightFileSystem) Create(path, contents string) (cwl.File, error) {
	path = s.getAbsPath(path)
	// UploadSimple + State
	_, err := s.client.Upload(path, []byte(contents), false)
	if err != nil {
		return cwl.File{}, err
	}
	fileInfo, err := s.client.State(path, false)
	return FileInfo2CwlFile(fileInfo), nil
}

func (s StarlightFileSystem) Info(loc string) (cwl.File, error) {
	loc = s.getAbsPath(loc)
	info, err := s.client.State(loc, s.checksum)
	return FileInfo2CwlFile(info), err
}

func (s StarlightFileSystem) DirInfo(loc string, deepLen int) (cwl.Directory, error) {
	loc = s.getAbsPath(loc)
	// 重复StateDirFiles

	// 本体
	dirInfo, err := s.client.State(loc, false)
	if err != nil {
		return cwl.Directory{}, err
	}
	dir := FileInfo2CwlDir(dirInfo)

	// 列表
	if deepLen > 0 {

		dir.Listing = []cwl.FileDir{}
		fileInfos, err := s.client.StateDir(loc, s.checksum)
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
	return s.client.Copy(source, dest, false, true)
}

func (s StarlightFileSystem) Contents(loc string) (string, error) {
	loc = s.getAbsPath(loc)
	// DownloadSimple
	data, err := s.client.Download(loc)
	return string(data), err
}

func (s StarlightFileSystem) Glob(pattern string) ([]cwl.File, error) {
	infos, err := s.client.Glob(s.workDir, pattern, s.checksum)
	if err != nil {
		return nil, err
	}
	var files []cwl.File
	for _, info := range infos {
		files = append(files, FileInfo2CwlFile(&info))
	}
	return files, nil
}

func (s StarlightFileSystem) EnsureDir(dir string, mode os.FileMode) error {
	dir = s.getAbsPath(dir)
	// State + MkdirOpt
	_, err := s.client.State(dir, false)
	if err != nil {
		if !ErrorIsNotExist(err) {
			return err
		} else {
			return s.client.Mkdir(dir, mode)
		}
	}
	return nil
}

func (s StarlightFileSystem) Migrate(source, dest string) (bool, error) {
	source = s.getAbsPath(source)
	dest = s.getAbsPath(dest)
	// Copy or Ln
	if s.canLink(source, dest) {
		return true, s.client.Link(source, dest, true)
	} else {
		return false, s.client.Copy(source, dest, true, true)
	}
}

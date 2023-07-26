package client

import (
	"os"
)

type StarlightStorageClient struct {
	base *StarlightClient
}

func (c StarlightStorageClient) Download(location string) ([]byte, error) {

}
func (c StarlightStorageClient) Upload(location string, data []byte, overwrite bool) (int64, error) {

}
func (c StarlightStorageClient) State(location string) (*FileInfo, error) {

}
func (c StarlightStorageClient) StateDir(location string) ([]FileInfo, error) {

}
func (c StarlightStorageClient) Copy(source, destination string, force, recursive bool) error {

}
func (c StarlightStorageClient) Link(source, destination string, symbolic bool) error {

}
func (c StarlightStorageClient) Mkdir(location string, mode os.FileMode, parent bool) error {

}

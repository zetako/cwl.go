package brunner

import (
	"encoding/json"
	"github.com/lijiang2014/cwl.go"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Importer interface {
	Load(string) ([]byte, error)
}

type DefaultImporter struct {
}

func (i *DefaultImporter) redirectURI(uri string) string {
	return filepath.Join("../../cwl/v1.0/v1.0", uri)
}

func (i *DefaultImporter) Load(uri string) ([]byte, error) {
	uri = i.redirectURI(uri)
	fs, err := os.Open(uri)
	if err != nil {
		return nil, err
	}
	defer fs.Close()
	return ioutil.ReadAll(fs)
}

func (e *K8sEngine) loadRoot(uri string) (*cwl.Root, error) {
	data, err := e.importer.Load(uri)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(uri, ".cwl") {
		data, _ = cwl.Y2J(data)
	}
	root := &cwl.Root{}
	err = json.Unmarshal(data, root)
	return root, err
}

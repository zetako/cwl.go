package server

import (
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/intergration/client"
	"io/ioutil"
	"os"
	"strings"
)

func splitPackedFile(raw string) (fileName, fragID string) {
	tmp := strings.IndexByte(raw, '#')
	if tmp < 0 {
		return raw, ""
	}
	return raw[:tmp], raw[tmp+1:]
}

func openFileAsJSON(pathlike string) ([]byte, error) {
	f, err := os.Open(pathlike)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	data, err = cwl.Y2J(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func getStringPointer(str string) *string {
	tmp := str
	return &tmp
}

func generateStarlightClient() (*client.StarlightClient, error) {
	// TODO
	return nil, fmt.Errorf("TODO")
}

package sfs

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	httpScheme          string = "http://"
	httpsScheme         string = "https://"
	starlightURLPattern string = `.*starlight.*\.nscc-gz\.cn`
)

func (s StarlightFileSystem) Load(doc string) ([]byte, error) {
	if strings.HasPrefix(doc, httpScheme) || strings.HasPrefix(doc, httpsScheme) {
		return s.loadFromNet(doc)
	}
	content, err := s.Contents(doc)
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}

func (s StarlightFileSystem) loadFromNet(doc string) ([]byte, error) {
	client := http.Client{}
	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, doc, nil)
	if err != nil {
		return nil, err
	}
	if isStarlightUrl(doc) {
		req.Header.Add("Bihu-Token", s.token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

func isStarlightUrl(doc string) bool {
	docURL, err := url.Parse(doc)
	if err != nil {
		return false
	}
	docHost := docURL.Hostname()

	matched, err := regexp.MatchString(starlightURLPattern, docHost)
	if err != nil {
		return false
	}
	return matched
}

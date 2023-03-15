package runner

import (
	"bytes"
	"crypto/sha1"
	"fmt"

	"github.com/lijiang2014/cwl.go"

	//"github.com/alecthomas/units"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Local struct {
	workdir      string
	CalcChecksum bool
}

func NewLocal(workdir string) *Local {
	return &Local{workdir, false}
}

func (l *Local) Glob(pattern string) ([]cwl.File, error) {
	var out []cwl.File

	pattern = filepath.Join(l.workdir, pattern)

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err, pattern)
	}

	for _, match := range matches {
		match, _ := filepath.Rel(l.workdir, match)
		f, err := l.Info(match)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", err, match)
		}
		out = append(out, f)
	}
	return out, nil
}

const MaxContentsBytes = 2 * 1024 * 1024

func (l *Local) Create(path, contents string) (cwl.File, error) {
	var x cwl.File
	if path == "" {
		return x, fmt.Errorf("can't create file with empty path")
	}

	b := []byte(contents)
	size := int64(len(b))
	//if units.MetricBytes(size) > process.MaxContentsBytes {
	if size > MaxContentsBytes {
		return x, fmt.Errorf("contents is max allowed size (%d) ", MaxContentsBytes)
	}
	loc := path
	if path[0] != '/' {
		loc = filepath.Join(l.workdir, path)
	}
	abs, err := filepath.Abs(loc)
	if err != nil {
		return x, fmt.Errorf("getting absolute path for %s: %s", loc, err)
	}
	fd, err := os.Create(loc)
	if err != nil {
		return x, fmt.Errorf("create file for %s: %s ( %s , %s)", loc, err, l.workdir, path)
	}
	_, err = fd.Write(b)
	if err != nil {
		return x, fmt.Errorf("write file contents for %s: %s", loc, err)
	}
	return cwl.File{
		ClassBase: cwl.ClassBase{"File"},
		Location:  abs,
		Path:      path,
		Checksum:  "sha1$" + fmt.Sprintf("%x", sha1.Sum(b)),
		Size:      size,
	}, nil
}

func (l *Local) Info(loc string) (cwl.File, error) {
	var x cwl.File
	if !filepath.IsAbs(loc) {
		loc = filepath.Join(l.workdir, loc)
	}

	st, err := os.Stat(loc)
	if os.IsNotExist(err) {
		return x, os.ErrNotExist
	}
	if err != nil {
		return x, err
	}

	abs, err := filepath.Abs(loc)
	if err != nil {
		return x, fmt.Errorf("getting absolute path for %s: %s", loc, err)
	}
	// TODO make this work with directories
	if st.IsDir() {
		return cwl.File{
			ClassBase: cwl.ClassBase{"Directory"},
			Location:  abs,
			Path:      abs,
		}, nil
		// return x, fmt.Errorf("can't call Info() on a directory: %s", loc)
	}

	checksum := ""
	if l.CalcChecksum {
		b, err := ioutil.ReadFile(loc)
		if err != nil {
			return x, fmt.Errorf("calculating checksum for %s: %s", loc, err)
		}
		checksum = "sha1$" + fmt.Sprintf("%x", sha1.Sum(b))
	}
	return cwl.File{
		ClassBase: cwl.ClassBase{"File"},
		Location:  abs,
		Path:      abs,
		Checksum:  checksum,
		Size:      st.Size(),
	}, nil
}

func (l *Local) Contents(loc string) (string, error) {
	if !filepath.IsAbs(loc) {
		loc = filepath.Join(l.workdir, loc)
	}

	fh, err := os.Open(loc)
	if os.IsNotExist(err) {
		return "", os.ErrNotExist
	}
	if err != nil {
		return "", err
	}
	defer fh.Close()

	buf := &bytes.Buffer{}
	r := &io.LimitedReader{R: fh, N: int64(MaxContentsBytes)}
	_, err = io.Copy(buf, r)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (l *Local) EnsureDir(dir string, mode os.FileMode) error {
	err := os.MkdirAll(dir, mode)
	if err == nil {
		return nil
	}
	if err == os.ErrExist {
		info, err := os.Stat(dir)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
	}
	return err
}

func (l *Local) Migrate(source, dest string) error {
	return os.Symlink(source, dest)
}

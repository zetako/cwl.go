package runner

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os/exec"
	"path"

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
	if st.IsDir() {
		return cwl.File{
			ClassBase: cwl.ClassBase{"Directory"},
			Location:  abs,
			Path:      abs,
		}, nil
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

func (l *Local) DirInfo(loc string, deepLen int) (cwl.Directory, error) {
	var x cwl.Directory
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
	if !st.IsDir() {
		return x, fmt.Errorf("can't call DirInfo() on not directory file: %s", loc)
	}
	return l.readAllFile(abs, deepLen)
}

func (l *Local) readAllFile(loc string, deepin int) (cwl.Directory, error) {
	var x cwl.Directory
	rd, err := ioutil.ReadDir(loc)
	if err != nil {
		return x, fmt.Errorf("read dir %s err: %s", loc, err)
	}
	listing := make([]cwl.FileDir, 0)
	if deepin != 0 {
		for _, fi := range rd {
			fullDir := path.Join(loc, fi.Name())
			if fi.IsDir() {
				// 判断是否为符号链接， 符号链接不当作 Dir 处理
				if fi.Mode()&os.ModeSymlink != 0 {
					s, err := l.Info(fullDir)
					if err != nil {
						return x, err
					}
					listing = append(listing, cwl.NewFileDir(s))
					continue
				}
				s, err := l.readAllFile(fullDir, deepin-1)
				if err != nil {
					return x, err
				}
				listing = append(listing, cwl.NewFileDir(s))
			} else {
				s, err := l.Info(fullDir)
				if err != nil {
					return x, err
				}
				listing = append(listing, cwl.NewFileDir(s))
			}
		}
	}
	return cwl.Directory{
		ClassBase: cwl.ClassBase{"Directory"},
		Location:  loc,
		Path:      loc,
		Listing:   listing,
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

func (l *Local) Copy(source, dest string) error {
	ss, err := os.Stat(source)
	if err != nil {
		return err
	}
	if ss.IsDir() {
		// just use cp -rf
		cmd := exec.Command("cp", "-rf", source, dest)
		outs, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("cp err: %s Err: %s", string(outs), err)
		}
		return nil
	}
	sfile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sfile.Close()
	dfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer dfile.Close()
	_, err = io.Copy(dfile, sfile)
	return err
}

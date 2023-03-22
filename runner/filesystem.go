package runner

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/lijiang2014/cwl.go"
)

type Filesystem interface {
	Create(path, contents string) (cwl.File, error)
	Info(loc string) (cwl.File, error)
	DirInfo(loc string, deepLen int) (cwl.Directory, error)
	Copy(source, dest string) error
	Contents(loc string) (string, error)
	Glob(pattern string) ([]cwl.File, error)
	EnsureDir(dir string, mode os.FileMode) error
	Migrate(source, dest string) error
}

func (process *Process) resolveFile(f cwl.File, loadContents bool) (cwl.File, error) {
	// TODO revisit pointer to File
	var x cwl.File
	var err error

	x.Class = f.Class
	// http://www.commonwl.org/v1.0/CommandLineTool.html#File
	// "As a special case, if the path field is provided but the location field is not,
	// an implementation may assign the value of the path field to location,
	// and remove the path field."
	if f.Location == "" && f.Path != "" && f.Contents == "" {
		f.Location = f.Path
		f.Path = ""
	}

	if f.Location == "" && f.Contents == "" {
		return x, process.error("location and contents are empty")
	}

	// If both location and contents are set, one will get overwritten.
	// Can't know which one the caller intended, so fail instead.
	if f.Location != "" && f.Contents != "" {
		return x, process.error("location and contents are both non-empty")
	}

	if f.Contents != "" {
		// Determine the file path of the literal.
		// Use the path, or the basename, or generate a random name.
		path := f.Path
		if path == "" {
			path = f.Basename
		}
		if path == "" {
			id, err := uuid.NewRandom()
			if err != nil {
				return x, process.errorf("generating a random name for a file literal: %s", err)
			}
			path = id.String()
			f.Path = path
		}
		// Create Later
		//x, err = process.fs.Create(path, f.Contents)
		//if err != nil {
		//  return x, process.errorf("creating file from inline content: %s", err)
		//}
		process.filesToCreate = append(process.filesToCreate, cwl.NewFileDir(f))
		// literal File logic
		// f.Location = x.Location
		// f.Checksum = x.Checksum
		f.Size = int64(len(f.Contents))
		if f.Basename == "" {
			f.Basename = filepath.Base(f.Path)
		}
		f.Nameroot, f.Nameext = splitname(f.Basename)
		f.Dirname = filepath.Dir(f.Path)
		return f, nil
	} else {
		x, err = process.fs.Info(f.Location)

		if err != nil {
			return x, process.errorf("getting file info for %q: %s", f.Location, err)
		}

		if loadContents {
			f.Contents, err = process.fs.Contents(f.Location)
			if err != nil {
				return x, process.errorf("loading file contents: %s", err)
			}
		}
	}

	// TODO clean this up. "x" was needed before a package reorg.
	//      possibly can be removed now.
	f.Location = x.Location
	// TODO figure out how to stage files.
	//      namespace inputs so they don't conflict.
	//      remember, the args building depends on this path, so it must happen
	//      in the ProcessBase code.
	//f.Path = filepath.Join("/inputs", filepath.Base(x.Path))
	f.Path = filepath.Base(x.Path)
	// f.Path = filepath.Join("inputs", filepath.Base(x.Path))
	f.Checksum = x.Checksum
	f.Size = x.Size

	// cwl spec:
	// "If basename is provided, it is not required to match the value from location"
	if f.Basename == "" {
		f.Basename = filepath.Base(f.Path)
	}
	f.Nameroot, f.Nameext = splitname(f.Basename)
	f.Dirname = filepath.Dir(f.Path)
	//f.Nameroot = process.runtime.RootHost
	return f, nil
}

func (process *Process) resolveDir(d cwl.Directory, loadListingType cwl.LoadListingEnum) (cwl.Directory, error) {
	// TODO revisit pointer to File
	var x cwl.Directory
	var err error

	x.Class = d.Class
	// https://www.commonwl.org/v1.0/CommandLineTool.html#Directory
	if d.Location == "" && d.Path != "" && d.Listing == nil {
		d.Location = d.Path
		d.Path = ""
	}

	if d.Location == "" && len(d.Listing) == 0 {
		return x, process.error("location and listing are empty")
	}

	// If both location and listing are set, one will get overwritten.
	// Can't know which one the caller intended, so fail instead.
	if d.Location != "" && len(d.Listing) != 0 {
		return x, process.error("location and listing are both non-empty")
	}

	if len(d.Listing) != 0 {
		// Determine the file path of the literal.
		// Use the path, or the basename, or generate a random name.
		path := d.Path
		if path == "" {
			path = d.Basename
		}
		if path == "" {
			id, err := uuid.NewRandom()
			if err != nil {
				return x, process.errorf("generating a random name for a file literal: %s", err)
			}
			path = id.String()
			d.Path = path
		}
		process.filesToCreate = append(process.filesToCreate, cwl.NewFileDir(d))
		if d.Basename == "" {
			d.Basename = filepath.Base(d.Path)
		}
		// resovle Files in Dir

		return d, nil
	}
	// location Dir
	var x2 cwl.File
	x2, err = process.fs.Info(d.Location)

	if err != nil {
		return x, process.errorf("getting file info for %q: %s", d.Location, err)
	}
	d.Location = x2.Location
	d.Path = filepath.Base(x2.Path)
	// cwl spec:
	// "If basename is provided, it is not required to match the value from location"
	if d.Basename == "" {
		d.Basename = filepath.Base(d.Path)
	}
	//f.Nameroot = process.runtime.RootHost
	if len(d.Listing) == 0 && loadListingType != cwl.NO_LISTING {
		// load listing
		var dl = 1
		if loadListingType == cwl.DEEP_LISTING {
			dl = -1
		}
		di, err := process.fs.DirInfo(d.Location, dl)
		if err != nil {
			return d, err
		}
		d.Listing = di.Listing
		clearDirListingPath(d.Listing, path.Clean(d.Location))
	}
	return d, nil
}

func clearDirListingPath(listing []cwl.FileDir, basedir string) {
	basedir = path.Clean(basedir)
	dirbase := path.Base(basedir)
	for i, fdi := range listing {
		filei, diri, err := fdi.Value()
		if err != nil {
			continue
		}
		if filei != nil {
			if strings.HasPrefix(filei.Path, basedir) {
				filei.Path = path.Join(dirbase, filei.Path[len(basedir)+1:])
				if filei.Basename == "" {
					filei.Basename = path.Base(filei.Path)
				}
				listing[i] = cwl.NewFileDir(filei)
				continue
			}
		} else if diri != nil {
			if strings.HasPrefix(diri.Path, basedir) {
				diri.Path = path.Join(dirbase, diri.Path[len(basedir)+1:])
				if diri.Basename == "" {
					diri.Basename = path.Base(diri.Path)
				}
				clearDirListingPath(diri.Listing, basedir)
				listing[i] = cwl.NewFileDir(diri)
				continue
			}
		}

	}
}

func (process *Process) resolveSecondaryFiles(file *cwl.File, x cwl.SecondaryFileSchema) error {

	// cwl spec:
	// "If the value is an expression, the value of self in the expression
	// must be the primary input or output File object to which this binding applies.
	// The basename, nameroot and nameext fields must be present in self.
	// For CommandLineTool outputs the path field must also be present.
	// The expression must return a filename string relative to the path
	// to the primary File, a File or Directory object with either path
	// or location and basename fields set, or an array consisting of strings
	// or File or Directory objects. It is legal to reference an unchanged File
	// or Directory object taken from input as a secondaryFile.
	//// TODO
	//if expr.IsExpression(x.FileDir) {
	//  process.eval(x, file)
	//}

	// TODO 需要获取 eval

	// cwl spec:
	// "If a value in secondaryFiles is a string that is not an expression,
	// it specifies that the following pattern should be applied to the location
	// of the primary file to yield a filename relative to the primary File:"

	// "If string begins with one or more caret ^ characters, for each caret,
	// remove the last file extension from the location (the last period . and all
	// following characters).
	pattern := string(x.Pattern)
	// TODO location or path? cwl spec says "path" but I'm suspicious.
	location := file.Location

	for strings.HasPrefix(pattern, "^") {
		pattern = strings.TrimPrefix(pattern, "^")
		location = strings.TrimSuffix(location, filepath.Ext(location))
	}

	// "Append the remainder of the string to the end of the file location."
	sec := cwl.File{
		Location: location + pattern,
	}
	sec.Class = "File"

	// TODO does LoadContents apply to secondary files? not in the spec
	f, err := process.resolveFile(sec, false)
	if err != nil {
		return err
	}

	file.SecondaryFiles = append(file.SecondaryFiles, cwl.NewFileDir(f))
	return nil
}

// splitname splits a file name into root and extension,
// with some special CWL rules.
func splitname(n string) (root, ext string) {
	// "For the purposess of path splitting leading periods on the basename are ignored;
	// a basename of .cshrc will have a nameroot of .cshrc."
	x := strings.TrimPrefix(n, ".")
	ext = filepath.Ext(x)
	root = strings.TrimSuffix(n, ext)
	return root, ext
}

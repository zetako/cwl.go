package cwl

import (
	"os"
	"path/filepath"
)

// Entry represents fs entry, it means [File|Directory|Dirent]
type Entry struct {
	Class    string  `json:"class,omitempty"`
	Location string  `json:"location,omitempty"`
	Path     string  `json:"path,omitempty"`
	Basename string  `json:"basename,omitempty"`
	File
	Directory
	Dirent
}

// File represents file entry.
// @see http://www.commonwl.org/v1.0/CommandLineTool.html#File
type File struct {
	Dirname string `json:"dirname,omitempty"`
	Size    int64 `json:"size"`
	Format  string `json:"format,omitempty"`
	//
	// extends
	//
	Nameroot       string    `json:"nameroot,omitempty"`
	Nameext        string    `json:"nameext,omitempty"`
	Checksum       string    `json:"checksum,omitempty"`
	Contents       string    `json:"contents,omitempty"`
	SecondaryFiles []Entry `json:"secondaryFiles,omitempty"`
}

// Directory represents direcotry entry.
// @see http://www.commonwl.org/v1.0/CommandLineTool.html#Directory
type Directory struct {
	Listing []Entry `json:"listing,omitempty"`
}

// Dirent represents ?
// @see http://www.commonwl.org/v1.0/CommandLineTool.html#Dirent
type Dirent struct {
	Entry     string `json:"entry,omitempty"`
	EntryName string `json:"entryname,omitempty"`
	Writable  bool `json:"writable,omitempty"`
}

// NewList constructs a list of Entry from interface
func (_ Entry) NewList(i interface{}) []Entry {
	dest := []Entry{}
	switch x := i.(type) {
	case string:
		dest = append(dest, Entry{}.New(x))
	case []interface{}:
		for _, v := range x {
			dest = append(dest, Entry{}.New(v))
		}
	}
	return dest
}

// New constructs an Entry from interface
func (_ Entry) New(i interface{}) Entry {
	dest := Entry{}
	switch x := i.(type) {
	case string:
		dest.Location = x
	case map[string]interface{}:
		for key, v := range x {
			switch key {
			case "entryname":
				dest.EntryName = v.(string)
			case "entry":
				dest.Entry = v.(string)
			case "writable":
				dest.Writable = v.(bool)
			}
		}
	}
	return dest
}

// LinkTo creates hardlink of this entry under destdir.
func (entry *Entry) LinkTo(destdir, srcdir string) error {
	destpath := filepath.Join(destdir, filepath.Base(entry.Location))
	if filepath.IsAbs(entry.Location) {
		return os.Link(entry.Location, destpath)
	}
	return os.Link(filepath.Join(srcdir, entry.Location), destpath)
}

package cwl

import (
	"fmt"
	"io/ioutil"
)

// Provided represents the provided input value
// by parameter files.
type Provided struct {
	ID    string
	Raw   interface{}
	Entry *FileDir // In most cases, it's "File" if "FileDir" exists
	Error error

	// TODO: Refactor
	Int int
}

// New constructs new "Provided" struct.
//func (provided Provided) New(id string, i interface{}) *Provided {
//	dest := &Provided{ID: id, Raw: i}
//	switch v := i.(type) {
//	case nil:
//		return nil
//	case int:
//		dest.Int = v
//	case map[interface{}]interface{}: // It's "File" in most cases
//		dest.Entry, dest.Error = dest.EntryFromDictionary(v)
//	}
//	return dest
//}

// EntryFromDictionary ...
func (provided Provided) EntryFromDictionary(dict map[interface{}]interface{}) (FileDir, error) {
	if dict == nil {
		return nil, nil
	}
	class := dict["class"].(string)
	location := dict["location"]
	contents := dict["contents"]
	if class == "" && location == nil && contents == nil {
		return nil, nil
	}
	switch class {
	case "File":
		// Use location if specified
		if location != nil {
			return &File{
				ClassBase: ClassBase{class},
				Location:  fmt.Sprintf("%v", location),
			}, nil
		}
		// Use contents if specified
		if contentsstring, ok := contents.(string); ok {
			tmpfile, err := ioutil.TempFile("/tmp", provided.ID)
			if err != nil {
				return nil, err
			}
			defer tmpfile.Close()
			if _, err := tmpfile.WriteString(contentsstring); err != nil {
				return nil, err
			}
			return &Directory{
				ClassBase: ClassBase{class},
				Location:  tmpfile.Name(),
			}, nil
		}
	}
	return nil, nil
}

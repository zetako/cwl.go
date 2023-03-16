package runner

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/lijiang2014/cwl.go"
	"github.com/spf13/cast"
)

func (e *Engine) Outputs() (cwl.Values, error) {
	return e.process.Outputs(e.outputFS)
}

// Outputs binds cwl.Tool output descriptors to concrete values.
func (process *Process) Outputs(fs Filesystem) (cwl.Values, error) {
	process.fs = fs
	outdoc, err := fs.Contents("cwl.output.json")
	if err == nil {
		outputs := cwl.Values{}
		err = json.Unmarshal([]byte(outdoc), &outputs)
		return outputs, err
	}

	outputs := cwl.Values{}
	for _, outi := range process.tool.Outputs {
		out := outi.(*cwl.CommandOutputParameter)
		v, err := process.bindOutput(fs, out.Type.SaladType, out.OutputBinding, out.SecondaryFiles, nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to bind value for "%s": %s`, out.ID, err)
		}
		outputs[out.ID] = v
	}
	return outputs, nil
}

// bindOutput binds the output value for a single CommandOutput.
// 解析 output 并返回
func (process *Process) bindOutput(
	fs Filesystem,
	typei cwl.SaladType,
	binding *cwl.CommandOutputBinding,
	secondaryFiles []cwl.SecondaryFileSchema,
	val interface{},
) (interface{}, error) {
	var err error

	if binding == nil && typei.IsRecord() {
		// 需要递归处理 的 record
		outRecord := typei.MustRecord().(*cwl.CommandOutputRecordSchema)
		valRecord := cwl.Values{}
		for _, fieldi := range outRecord.Fields {
			var outi interface{}
			fieldiCmdOut := fieldi.(*cwl.CommandOutputRecordField)
			outi, err = process.bindOutput(fs, fieldiCmdOut.Type, fieldiCmdOut.OutputBinding, fieldiCmdOut.SecondaryFiles, nil)
			if err != nil {
				return nil, err
			}
			valRecord[fieldi.FieldName()] = outi
		}
		return valRecord, nil
	}

	if binding != nil && len(binding.Glob) > 0 {
		// glob patterns may be expressions. evaluate them.
		globs, err := process.evalGlobPatterns(binding.Glob)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate glob expressions: %s", err)
		}

		files, err := process.matchFiles(fs, globs, binding.LoadContents.LoadContents)
		if err != nil {
			return nil, fmt.Errorf("failed to match files: %s", err)
		}
		val = files
	}

	if binding != nil && binding.OutputEval != "" {
		val, err = toJSONMap(val)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate outputEval: %s", err)
		}
		val, err = process.eval(binding.OutputEval, val)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate outputEval: %s", err)
		}
	}

	if val == nil {
		if typei.IsNullable() {
			return nil, nil
		}
	}

	types := make([]cwl.SaladType, 0)
	if typei.IsMulti() {
		types = typei.MustMulti()
	} else {
		types = append(types, typei)
	}

	for _, t := range types {
		switch t.TypeName() {
		// TODO validate stdout/err can only be at root
		//      validate that stdout/err doesn't occur more than once
		case "stdout":
			files, err := process.matchFiles(fs, []string{process.stdout}, false)
			if err != nil {
				return nil, fmt.Errorf("failed to match files: %s", err)
			}
			if len(files) == 0 {
				return nil, fmt.Errorf(`failed to match stdout file "%s"`, process.stdout)
			}
			if len(files) > 1 {
				return nil, fmt.Errorf(`matched multiple stdout files "%s"`, process.stdout)
			}
			return process.clearOutputFile(*(files[0].Entery().(*cwl.File)), process.runtime.RootHost), nil
			// return *(files[0].Entery().(*cwl.File)), nil

		case "stderr":
			files, err := process.matchFiles(fs, []string{process.stderr}, false)
			if err != nil {
				return nil, fmt.Errorf("failed to match files: %s", err)
			}
			if len(files) == 0 {
				return nil, fmt.Errorf(`failed to match stderr file "%s"`, process.stderr)
			}
			if len(files) > 1 {
				return nil, fmt.Errorf(`matched multiple stderr files "%s"`, process.stderr)
			}
			return *(files[0].Entery().(*cwl.File)), nil
		}
	}

	if val == nil {
		return nil, fmt.Errorf("missing value")
	}

	// Bind the output value to one of the allowed types.
Loop:
	for _, t := range types {
		switch t.TypeName() {
		case "bool":
			v, err := cast.ToBoolE(val)
			if err == nil {
				return v, nil
			}
		case "int":
			v, err := cast.ToInt32E(val)
			if err == nil {
				return v, nil
			}
		case "long":
			v, err := cast.ToInt64E(val)
			if err == nil {
				return v, nil
			}
		case "float":
			v, err := cast.ToFloat32E(val)
			if err == nil {
				return v, nil
			}
		case "double":
			v, err := cast.ToFloat64E(val)
			if err == nil {
				return v, nil
			}
		case "string":
			v, err := cast.ToStringE(val)
			if err == nil {
				return v, nil
			}
		case "File":
			switch y := val.(type) {
			case []cwl.File:
				if len(y) != 1 {
					continue Loop
				}
				f := y[0]
				for _, expr := range secondaryFiles {
					err := process.resolveSecondaryFiles(f, expr)
					if err != nil {
						return nil, fmt.Errorf("resolving secondary files: %s", err)
					}
				}
				return f, nil
			case []cwl.FileDir:
				if len(y) != 1 {
					continue Loop
				}
				f := y[0].Entery().(*cwl.File)
				for _, expr := range secondaryFiles {
					err := process.resolveSecondaryFiles(*f, expr)
					if err != nil {
						return nil, fmt.Errorf("resolving secondary files: %s", err)
					}
				}
				// return *f, nil
				return process.clearOutputFile(*f, process.runtime.RootHost), nil
			case cwl.File:
				return y, nil
			case cwl.FileDir:
				yi := y.Entery()
				fi, ok := yi.(*cwl.File)
				if !ok {
					return nil, fmt.Errorf("not File, not ok yet!: %s", yi.ClassName())
				}
				for _, expr := range secondaryFiles {
					err := process.resolveSecondaryFiles(*fi, expr)
					if err != nil {
						return nil, fmt.Errorf("resolving secondary files: %s", err)
					}
				}
				return process.clearOutputFile(*fi, process.runtime.RootHost), nil
			default:
				continue Loop
			}
		case "Directory":
			// TODO
			var dir cwl.Directory
			if fd, ok := val.([]cwl.FileDir); ok {
				if len(fd) != 1 {
					continue Loop
				}
				val = fd[0]
			}
			if fd, ok := val.(cwl.FileDir); ok {
				dirPt, got := fd.Entery().(*cwl.Directory)
				if !got {
					continue Loop
				}
				dir = *dirPt
			} else if dd, ok := val.(cwl.Directory); ok {
				dir = dd
			}
			return process.clearOutputDir(dir, process.runtime.RootHost), nil
		case "array":
			typ := reflect.TypeOf(val)
			if typ.Kind() != reflect.Slice {
				continue Loop
			}
			array := t.MustArraySchema().(*cwl.CommandOutputArraySchema)
			var res []cwl.Value
			arr := reflect.ValueOf(val)
			for i := 0; i < arr.Len(); i++ {
				item := arr.Index(i)
				if !item.CanInterface() {
					return nil, fmt.Errorf("can't get interface of array item")
				}
				r, err := process.bindOutput(fs, *array.GetItems(), nil, nil, item.Interface())
				if err != nil {
					return nil, err
				}
				res = append(res, r)
			}
			return res, nil

		case "record":
			// TODO

		}
	}

	return nil, fmt.Errorf("no type could be matched")
}

// matchFiles executes the list of glob patterns, returning a list of matched files.
// matchFiles must return a non-nil list on success, even if no files are matched.
func (process *Process) matchFiles(fs Filesystem, globs []string, loadContents bool) ([]cwl.FileDir, error) {
	// it's important this slice isn't nil, because the outputEval field
	// expects it to be non-null during expression evaluation.
	files := []cwl.FileDir{}

	// resolve all the globs into file objects.
	for _, pattern := range globs {
		matches, err := fs.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to execute glob: %s", err)
		}

		for _, m := range matches {
			// TODO handle directories
			if m.ClassName() == "Directory" {
				v := cwl.Directory{
					ClassBase: cwl.ClassBase{"Directory"},
					Location:  m.Location,
					Path:      m.Path,
				}
				d, err := process.resolveDir(v)
				if err != nil {
					return nil, err
				}
				files = append(files, cwl.NewFileDir(&d))
				continue
			}
			v := cwl.File{
				ClassBase: cwl.ClassBase{"File"},
				//Class:    "File",
				Location: m.Location,
				Path:     m.Path,
				//File: cwl.File{
				Checksum: m.Checksum,
				Size:     m.Size,
				//},
			}
			f, err := process.resolveFile(v, loadContents)
			if err != nil {
				return nil, err
			}
			files = append(files, cwl.NewFileDir(&f))
		}
	}
	return files, nil
}

// evalGlobPatterns evaluates a list of potential expressions as defined by the CWL
// OutputBinding.glob field. It returns a list of strings, which are glob expression,
// or an error.
//
// cwl spec:
// "If an expression is provided, the expression must return a string or an array
//  of strings, which will then be evaluated as one or more glob patterns."
func (process *Process) evalGlobPatterns(patterns []cwl.Expression) ([]string, error) {
	var out []string

	for _, pattern := range patterns {
		// TODO what is "self" here?
		val, err := process.eval(pattern, nil)
		if err != nil {
			return nil, err
		}

		switch z := val.(type) {
		case string:
			out = append(out, z)
		case []string:
			out = append(out, z...)
		case []interface{}:
			for _, val := range z {
				z, ok := val.(string)
				if !ok {
					return nil, fmt.Errorf(
						"glob expression returned an invalid type. Only string or []string "+
							"are allowed. Got: %#v", z)
				}
				out = append(out, z)
			}
		case []cwl.Value:
			for _, val := range z {
				z, ok := val.(string)
				if !ok {
					return nil, fmt.Errorf(
						"glob expression returned an invalid type. Only string or []string "+
							"are allowed. Got: %#v", z)
				}
				out = append(out, z)
			}
		default:
			return nil, fmt.Errorf(
				"glob expression returned an invalid type. Only string or []string "+
					"are allowed. Got: %#v", z)
		}
	}
	return out, nil
}

func (process *Process) clearOutputFileDir(in cwl.FileDir, root string) (out cwl.FileDir) {
	entry := in.Entery()
	if file, ok := entry.(cwl.File); ok {
		fileout := process.clearOutputFile(file, root)
		return cwl.NewFileDir(&fileout)
	}
	if dict, ok := entry.(cwl.Directory); ok {
		dirout := process.clearOutputDir(dict, root)
		return cwl.NewFileDir(&dirout)
	}
	if file, ok := entry.(*cwl.File); ok {
		fileout := process.clearOutputFile(*file, root)
		return cwl.NewFileDir(&fileout)
	}
	if dict, ok := entry.(*cwl.Directory); ok {
		dirout := process.clearOutputDir(*dict, root)
		return cwl.NewFileDir(&dirout)
	}
	return out
}

func (process *Process) clearOutputFile(in cwl.File, root string) (file cwl.File) {
	file = in
	if strings.HasPrefix(file.Location, root) {
		location := file.Location[len(root):]
		if location != "" && location[0] == '/' {
			location = location[1:]
		}
		file.Location = location
	}
	file.Path, file.Basename, file.Dirname, file.Nameroot, file.Nameext = "", "", "", "", ""
	for i, sfi := range file.SecondaryFiles {
		file.SecondaryFiles[i] = process.clearOutputFileDir(sfi, root)
	}
	return file
}

func (process *Process) clearOutputDir(in cwl.Directory, root string) (file cwl.Directory) {
	file = in
	if len(file.Listing) == 0 {
		// 尝试获取 Listing
		d, _ := process.fs.DirInfo(file.Location, -1)
		file.Listing = d.Listing
	}
	if strings.HasPrefix(file.Location, root) {
		location := file.Location[len(root):]
		if location != "" && location[0] == '/' {
			location = location[1:]
		}
		file.Location = location
	}
	file.Path, file.Basename = "", ""

	for i, sfi := range file.Listing {
		file.Listing[i] = process.clearOutputFileDir(sfi, root)
	}
	return file
}

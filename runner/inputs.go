package runner

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/lijiang2014/cwl.go"
	"github.com/spf13/cast"
)

// bindInput
// 根据 input 产生 binding 数据
func (process *Process) bindInput(
	name string,
	typein cwl.SaladType,
	clb *cwl.CommandLineBinding,
	secondaryFiles []cwl.SecondaryFileSchema,
	//val interface{},
	val cwl.Value,
	key sortKey,
) ([]*Binding, error) {

	// If no value was found, check if the type is allowed to be null.
	// If so, return a binding, otherwise fail.
	if val == nil {
		if typein.IsNullable() {
			return []*Binding{
				{clb, typein, nil, key, nil, name},
			}, nil
		}
		return nil, process.error("missing value")
	}

	types := make([]cwl.SaladType, 0)
	if typein.IsMulti() {
		types = typein.MustMulti()
	} else {
		types = append(types, typein)
	}

Loop:
	// An input descriptor describes multiple allowed types.
	// Loop over the types, looking for the best match for the given input value.
	for _, ti := range types {
		switch ti.TypeName() {
		case "null":
			continue
		case "array":
			vals, ok := val.([]cwl.Value)
			if !ok {
				// input value is not an array.
				continue Loop
			}

			// The input array is allowed to be empty,
			// so this must be a non-nil slice.
			out := []*Binding{}
			t := ti.MustArraySchema().(*cwl.CommandInputArraySchema)

			for i, itemVal := range vals {
				subkey := append(key, sortKey{getPos(t.InputBinding), i}...)
				var b, err = process.bindInput("", t.Items, t.InputBinding, nil, itemVal, subkey)
				if err != nil {
					return nil, err
				}
				if b == nil {
					// array item values did not bind to the array descriptor.
					continue Loop
				}
				out = append(out, b...)
			}

			nested := make([]*Binding, len(out))
			copy(nested, out)
			b := &Binding{clb, ti, val, key, nested, name}
			// TODO revisit whether creating a nested tree (instead of flat) is always better/ok
			return []*Binding{b}, nil

		case "enum":
			v, ok := val.(string)
			if !ok {
				// input value is not a record.
				continue Loop
			}
			t := typein.MustEnum().(*cwl.CommandInputEnumSchema)
			if clb == nil && t.InputBinding != nil {
				clb = t.InputBinding
			}
			for _, symbol := range t.Symbols {
				if v == symbol {
					return []*Binding{
						{clb, ti, v, key, nil, name},
					}, nil
				}
			}
			continue Loop
		case "record":
			vals, ok := val.(map[string]cwl.Value)
			if !ok {
				// input value is not a record.
				continue Loop
			}
			t := typein.MustRecord().(*cwl.CommandInputRecordSchema)
			var out []*Binding

			for i, fieldi := range t.Fields {
				field := fieldi.(*cwl.CommandInputRecordField)
				val, ok := vals[field.Name]
				// TODO lower case?
				if !ok {
					if field.Type.IsNullable() {
						continue
					}
					continue Loop
				}
				// 如没有指定则采用 key 进行排序
				_ = i
				subkey := append(key, sortKey{getPos(field.InputBinding), field.Name}...)
				if len(key) == 1 && key[0] == 0 {
					// ✅  T73 ❌ T3
					subkey = sortKey{getPos(field.InputBinding), field.Name}
				}
				b, err := process.bindInput(field.Name, field.Type, field.InputBinding, nil, val, subkey)
				if err != nil {
					//return nil, err
					continue Loop
				}
				if b == nil {
					continue Loop
				}
				out = append(out, b...)
			}

			if out != nil {
				nested := make([]*Binding, len(out))
				copy(nested, out)
				b := &Binding{clb, ti, vals, key, nested, name}
				return []*Binding{b}, nil
				// return out, nil
			}

		case "any", "Any":
			return []*Binding{
				{clb, ti, val, key, nil, name},
			}, nil

		case "boolean":
			v, err := cast.ToBoolE(val)
			if err != nil {
				continue Loop
			}
			return []*Binding{
				{clb, ti, v, key, nil, name},
			}, nil

		case "int":
			v, err := cast.ToInt32E(val)
			if err != nil {
				continue Loop
			}
			return []*Binding{
				{clb, ti, v, key, nil, name},
			}, nil

		case "long":
			v, err := cast.ToInt64E(val)
			if err != nil {
				continue Loop
			}
			return []*Binding{
				{clb, ti, v, key, nil, name},
			}, nil

		case "float":
			v, err := cast.ToFloat32E(val)
			if err != nil {
				continue Loop
			}
			return []*Binding{
				{clb, ti, v, key, nil, name},
			}, nil

		case "double":
			v, err := cast.ToFloat64E(val)
			if err != nil {
				continue Loop
			}
			return []*Binding{
				{clb, ti, v, key, nil, name},
			}, nil

		case "string":
			v, err := cast.ToStringE(val)
			if err != nil {
				continue Loop
			}

			return []*Binding{
				{clb, ti, v, key, nil, name},
			}, nil

		case "File":
			var v cwl.File
			def, ok := val.(*cwl.File)
			if ok {
				v = *def
			} else {
				v, ok = val.(cwl.File)
			}
			if !ok {
				continue Loop
			}

			f, err := process.resolveFile(v, getLoadContents(clb))
			if err != nil {
				return nil, err
			}
			// use process.runtime.RootHost as /
			// TODO 这种命名方式可能导致同名文件冲突问题
			f.Path = filepath.Join(process.runtime.RootHost, "/inputs/", f.Path)

			//f.Path = "/inputs/" + f.Path
			for _, expr := range secondaryFiles {
				process.resolveSecondaryFiles(f, expr)
			}

			return []*Binding{
				{clb, ti, f, key, nil, name},
			}, nil

		case "Directory":
			var v cwl.Directory
			def, ok := val.(*cwl.Directory)
			if ok {
				v = *def
			} else {
				v, ok = val.(cwl.Directory)
			}
			if !ok {
				continue Loop
			}
			f, err := process.resolveDir(v)
			if err != nil {
				return nil, err
			}
			// use process.runtime.RootHost as /
			// TODO 这种命名方式可能导致同名文件冲突问题
			f.Path = filepath.Join(process.runtime.RootHost, "/inputs/", f.Path)

			// TODO resolve directory
			return []*Binding{
				{clb, ti, f, key, nil, name},
			}, nil
		default:
			tiTypeName := ti.TypeName()
			//if len(tiTypeName) > 0 && tiTypeName[0] == '#' {
			// is Ref Types
			if rsd := process.root.Process.Base().RequiresSchemaDef(); rsd != nil {
				for _, rsdT := range rsd.Types {
					var binding *cwl.CommandLineBinding
					rsdType := rsdT.(*cwl.CommandInputType)
					SName := rsdType.SchemaTypename()
					if SName == "" {
						if record := rsdType.MustRecord(); record != nil {
							cmdRecord := record.(*cwl.CommandInputRecordSchema)
							SName = cmdRecord.Name
							binding = cmdRecord.InputBinding
						} else if enum := rsdType.MustEnum(); enum != nil {
							cmdEnum := enum.(*cwl.CommandInputEnumSchema)
							SName = cmdEnum.Name
							binding = cmdEnum.InputBinding

						}
					}
					log.Println(SName, rsdType.TypeName())

					if SName == tiTypeName || (SName == tiTypeName[1:] && tiTypeName[0] == '#') {
						//	//do something
						//	// how to parse Value -> rsdT.Type ?
						//	st := rsdT.Type
						//	return []*Binding{
						//		{clb, ti, st, key, nil, name},
						//	}, nil
						b, err := process.bindInput(name, rsdType.SaladType, binding, nil, val, key)
						if err != nil {
							continue Loop
						}
						if b == nil {
							continue Loop
						}
						return b, nil
					}
					continue
				}
			}
			// TODO with TypeRef
			continue Loop
		}
	}

	return nil, process.error("missing value")
}

func getLoadContents(clb *cwl.CommandLineBinding) bool {
	if clb == nil {
		return false
	}
	if clb.LoadContents == nil {
		return false
	}
	return *clb.LoadContents
}

func (process *Process) MigrateInputs() (err error) {
	files := []cwl.File{}
	dirs := []cwl.Directory{}
	flatted := flatBinding(process.bindings, false)
	for _, in := range flatted {
		if f, ok := in.Value.(cwl.File); ok {
			files = append(files, FlattenFiles(f)...)
		} else if d, ok := in.Value.(cwl.Directory); ok {
			dirs = append(dirs, d)
		}
	}
	// Do migrate
	fs := process.fs
	//if err = fs.EnsureDir(process.runtime.RootHost, 0750); err != nil {
	//  return err
	//}
	if err = fs.EnsureDir(process.runtime.RootHost+"/inputs", 0750); err != nil {
		return err
	}
	for _, filei := range files {
		if filei.Contents != "" && filei.Location == "" && filei.Path != "" {
			if _, err = fs.Create(filei.Path, filei.Contents); err != nil {
				return err
			}
		} else if err = fs.Migrate(filei.Location, filei.Path); err != nil {
			return err
		}
	}
	for _, diri := range dirs {
		// 没有 listing 的文件夹 直接迁移；有 listing 的文件夹 按 listing 创建
		// https://common-workflow-lab.github.io/CWLDotNet/reference/CWLDotNet.Directory.html
		if len(diri.Listing) != 0 {
			// TODO 递归创建
			return process.error("dir listing stage not ok yet !")
		}
		// if diri.Path == "" {
		// 	diri.Path = path.Join(process.runtime.RootHost, "inputs", )
		// }
		// log.Println(diri.Location, diri.Path)
		if err = fs.Migrate(diri.Location, diri.Path); err != nil {
			return err
		}
	}

	if riwd := process.tool.RequiresInitialWorkDir(); riwd != nil {
		if err = process.initWorkDir(riwd.Listing); err != nil {
			return err
		}
	}
	return nil
}

func FlattenFiles(f cwl.File) []cwl.File {
	return []cwl.File{f}
}

func (process *Process) initWorkDir(listing []cwl.FileDirExpDirent) error {
	for _, v := range listing {
		if v.Expression != "" {
			out, err := process.eval(v.Expression, nil)
			if err != nil {
				return process.error(fmt.Sprintf("initWorkDir eval expression err %s", err))
			}
			filedir := cwl.FileDir{}
			raw, err := json.Marshal(out)
			if err != nil {
				return process.error(fmt.Sprintf("initWorkDir eval Marshal err %s", err))
			}
			err = json.Unmarshal(raw, &filedir)
			if err != nil {
				return process.error(fmt.Sprintf("initWorkDir eval Unmarshal err %s", err))
			}
			if filedir.ClassName() == "File" {
				v.File = filedir.Entery().(*cwl.File)
			} else if filedir.ClassName() == "Directory" {
				v.Directory = filedir.Entery().(*cwl.Directory)
			} else {
				return process.error(fmt.Sprintf("initWorkDir eval need File/Directory but not! :%s", string(raw)))
			}
		}
		if dirent := v.Dirent; dirent != nil {
			if err := process.initWorkDirDirent(*dirent); err != nil {
				return err
			}
			continue
		}
		if file := v.File; file != nil {
			if err := process.initWorkDirFile(*file); err != nil {
				return err
			}
			continue
		}
		if dir := v.Directory; dir != nil {
			return process.error(fmt.Sprintf("initWorkDir Directory not ok yet!"))
		}
	}
	return nil
}

func (process *Process) initWorkDirDirent(dirent cwl.Dirent) error {
	filename, err := process.Eval(dirent.EntryName, nil)
	if err != nil {
		return err
	}
	entry, err := process.Eval(dirent.Entry, nil)
	if err != nil {
		return err
	}
	filenamestr, ok := filename.(string)
	if !ok {
		return process.error("entryName need be string")
	}
	switch entryV := entry.(type) {
	case string:
		if _, err = process.fs.Create(process.runtime.RootHost+"/"+filenamestr, entryV); err != nil {
			return err
		}
	case map[string]interface{}:
		if entryV["class"] == "File" {
			var file cwl.File
			raw, _ := json.Marshal(entryV)
			err = json.Unmarshal(raw, &file)
			if err != nil {
				return err
			}
			file.Path = filenamestr
			return process.initWorkDirFile(file)
		} else if entryV["class"] == "Directory" {
			var dir cwl.Directory
			raw, _ := json.Marshal(entryV)
			err = json.Unmarshal(raw, &dir)
			if err != nil {
				return err
			}
			dir.Path = filenamestr
			return process.initWorkDirDirectory(dir)
		}
		if entryV["class"] != "File" {
			return process.error("entry is not File, not ok yet!")
		}

	default:
		return process.error("bad entry")
	}
	return nil
}

func (process *Process) initWorkDirFile(file cwl.File) error {
	// filenamestr := file.Basename
	if filepath := file.Path; filepath != "" {
		inputdir := path.Join(process.runtime.RootHost, "inputs")
		if strings.HasPrefix(filepath, inputdir) {
			file.Path = file.Path[len(inputdir)+1:]
		}
		if !path.IsAbs(file.Path) {
			file.Path = path.Join(process.runtime.RootHost, file.Path)
		}
		// TODO 对于 initWorkDir， 大部分场景应该使用 copy 而非 link
		return process.fs.Copy(file.Location, file.Path)
		// return process.fs.Migrate(file.Location, file.Path)
	}
	return process.error("bad file")
}

func (process *Process) initWorkDirDirectory(dir cwl.Directory) error {
	if filepath := dir.Path; filepath != "" {
		inputdir := path.Join(process.runtime.RootHost, "inputs")
		if strings.HasPrefix(filepath, inputdir) {
			dir.Path = dir.Path[len(inputdir)+1:]
		}
		if !path.IsAbs(dir.Path) {
			dir.Path = path.Join(process.runtime.RootHost, dir.Path)
		}
		if dir.Location != "" {
			return process.fs.Copy(dir.Location, dir.Path)
		}
		if err := process.fs.EnsureDir(dir.Path, 0750); err != nil {
			return process.errorf("init workdir >> ensureDir err : %s", err)
		}
		for _, filediri := range dir.Listing {
			if filediri.ClassName() == "File" {
				file := filediri.Entery().(*cwl.File)
				if err := process.initWorkDirFile(*file); err != nil {
					return process.errorf("init workdir >> initFile %s %s err : %s", file.Path, file.Location, err)
				}
			} else if filediri.ClassName() == "Directory" {
				directory := filediri.Entery().(*cwl.Directory)
				if err := process.initWorkDirDirectory(*directory); err != nil {
					return process.errorf("init workdir >> initDirectory %s %s err : %s", directory.Path, directory.Location, err)
				}
			}
		}
		return nil
		// return process.fs.Migrate(file.Location, file.Path)
	}
	return process.error("bad file")
}

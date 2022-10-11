package cwl

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"

	"reflect"
	"strings"
)

type RecordFieldGraph struct {
	Example interface{}
	Fields map[string]*RecordFieldGraph
	ID string
}

type testClass struct {
	Class string `json:"class"`
}

var classMap = map[string]interface{}{}

var noClassError = fmt.Errorf("no class for struct")
var unknownClassError = fmt.Errorf("unknown class for struct")

// 根据接口映射产生
func GenerateTypesFormClass(raw []byte, db map[string]interface{}) (reflect.Type, error) {
	class := &testClass{}
	json.Unmarshal(raw, class)
	if name := class.Class; name != "" {
		if db != nil {
			v, got := db[name]
			if got {
				return reflect.TypeOf(v), nil
			}
		}
		return reflect.TypeOf(nil), fmt.Errorf("unknown class for struct %s", name)
	}
	return reflect.TypeOf(nil), noClassError
}

// 根据接口映射产生
func GenerateTypesFormInterface(iType reflect.Type, db map[string]*RecordFieldGraph) (reflect.Type, error) {
	record, got := db[iType.Name()]
	if got {
		return reflect.TypeOf(record.Example), nil
	}
	return reflect.TypeOf(nil), fmt.Errorf("Cannot Generate Type %s", iType.Name())
}

// Just For test
func (p *ProcessBase) UnmarshalJSON(data []byte) error {
	typeOfRecv := reflect.TypeOf(*p)
	valueOfRecv := reflect.ValueOf(p).Elem()
	db := make(map[string]*RecordFieldGraph)
	db["InputParameter"] = &RecordFieldGraph{Example: InputParameterBase{}}
	db["OutputParameter"] = &RecordFieldGraph{Example: OutputParameterBase{}}
	if err := parseObject(typeOfRecv, valueOfRecv, data, db); err != nil {
		return err
	}
	return nil
}

func JsonUnmarshal(data []byte,bean interface{}, graphs... RecordFieldGraph) error {
	typeOfRecv := reflect.TypeOf(bean).Elem()
	valueOfRecv := reflect.ValueOf(bean).Elem()
	db := make(map[string]*RecordFieldGraph)
	for i, gi := range graphs {
		db[gi.ID] = &graphs[i]
	}
	return parseObject(typeOfRecv, valueOfRecv, data, db)
}

// parseObject
// 根据 salad 扩展 json 结构体的解析
// ✅ 支持 salad.mapSubject 特性
// - [ ] TODO 支持 salad.default 特性
// ✅ 支持 通过 ClassBase 来推断 Interface 的 具体 Struct
// ✅ 支持 json:inline 特性
//	- [ ] 优化：避免 map[key]Raw 的重复解析
func parseObject(typeOfRecv reflect.Type, valueOfRecv reflect.Value,
	data []byte, db map[string]*RecordFieldGraph) (err error) {
	//log.Println("old value", typeOfRecv.Name(), valueOfRecv.Interface(), valueOfRecv.Type().Name())
	if (valueOfRecv.Kind() == reflect.Interface || valueOfRecv.Kind() == reflect.Ptr) && valueOfRecv.IsNil() {
		return nil
	}
	if typeOfRecv.Kind() == reflect.Interface {
		typeOfRecv = typeOfRecv.Elem()
	}
	if valueOfRecv.Kind() == reflect.Interface {
		valueOfRecv = valueOfRecv.Elem()
		//log.Println("new value", typeOfRecv.Name(), valueOfRecv.Interface(), valueOfRecv.Type().Name())
		if valueOfRecv.IsNil() {
			return nil
		}
	}
	recvName := typeOfRecv.Name()
	_ = recvName
	if valueOfRecv.Kind() == reflect.Ptr {
		valueOfRecv = valueOfRecv.Elem()
		//log.Println("pt value", typeOfRecv.Name(), valueOfRecv.Interface(), valueOfRecv.Type().Name())
	}
	bean := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &bean); err != nil {
		return err
	}
	keyMap := make(map[string]string) // json-Name : go-Name
	inlineFields := make(map[string]string)
	saladFields := make(map[string]saladTags)

	for i := 0; i < typeOfRecv.NumField(); i++ {
		field := typeOfRecv.Field(i)
		keyGo := field.Name
		key := keyGo
		if tag := field.Tag.Get("json"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				key = parts[0]
			}
			if len(parts) > 1 && parts[0] == "" && parts[1] == "inline" {
				inlineFields[keyGo] = key
			}
		}
		if tag := field.Tag.Get("salad"); tag != "" {
			v := getSaladTags(tag)
			saladFields[key] = v
			if v.Default != "" {
				setFieldDefaultValue(typeOfRecv.Field(i).Type, valueOfRecv.Field(i), v.Default)
			}
		}
		keyMap[key] = keyGo
	}
	// handle inline fields
	for keyGo, _ := range inlineFields {
		field, got := typeOfRecv.FieldByName(keyGo)
		if !got {
			continue
		}
		if valueOfRecv.Kind() == reflect.Interface {
			continue
		}
		valueField := valueOfRecv.FieldByName(keyGo)
		if err = parseObject(field.Type, valueField, data, db); err != nil {
			return err
		}
	}
	// handle other fields
	for key, value := range bean {
		var salad saladTags
		var keyGo = key

		//keyGo = strings.ToUpper(key[0:1]) + key[1:]
		keyGo = keyMap[key]
		field, got := typeOfRecv.FieldByName(keyGo)
		if !got {
			continue
		}
		fieldValue := valueOfRecv.FieldByName(keyGo)
		fieldType := field.Type
		if v, got := saladFields[key]; got {
			salad = v
		}
		//log.Println("set fieldName", key, recvName)
		fdb := db
		if nextdb , got := db[fieldType.Name()]; got && nextdb.Fields != nil {
			fdb = nextdb.Fields
		}
		if salad.IsType {
			err = setType(fieldType, fieldValue, value, salad,  fdb)
		} else {
			err = setField(fieldType, fieldValue, value, salad, fdb)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func debugType(fieldType reflect.Type) {
	for i := 0; i < fieldType.NumMethod(); i++ {
		log.Printf("%s %d %s", fieldType.Name(), i, fieldType.Method(i).Name)
	}
}

func setField(fieldType reflect.Type, fieldValue reflect.Value, bean []byte,
	salad saladTags, db map[string]*RecordFieldGraph) (err error) {
	fkind := fieldType.Kind()
	//log.Println("setField", fieldType.Name(), fkind.String(), fieldValue.Type().Name(), fieldValue.Interface())
	// 如果本身有解析函数则直接调用 ✅
	//debugType(fieldValue.Type())
	// 可能需要分配空间的情况
	switch fkind {
	// 列表的解析
	case reflect.Ptr:
		fieldType = fieldType.Elem()
		if fieldValue.IsNil() {
			if len(bean) == 0 || (len(bean) == 4 && string(bean) == "null") {
				if salad.Default == "" {
					return nil
				}
				return setFieldDefaultValue(fieldType, fieldValue, salad.Default)
			}
			// 需要初始化Value
			fieldValue.Set(reflect.New(fieldType))
		}
		return setField(fieldType, fieldValue, bean, salad, db)
	case reflect.Struct:
		if ok, err := checkUnmarshal(fieldValue, bean); ok || err != nil {
			return err
		}
		return parseObject(fieldType, fieldValue, bean, db)
	case reflect.Slice:
		fieldType = fieldType.Elem()
		isInterface := fieldType.Kind() == reflect.Interface
		var values []json.RawMessage
		// 处理 mapSubject
		// 初始化处理
		isNull := false
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.MakeSlice(reflect.SliceOf(fieldType), len(values), len(values)))
			isNull = true
		}
		if got, err := checkUnmarshal(fieldValue, bean); got || err != nil {
			return err
		}

		if salad.MapSubject != "" {
			// 需要进行 mapSubject 处理
			values, err = JsonldPredicateMapSubject(bean, salad.MapSubject, salad.MapPredicate)
			//return nil
		} else {
			values = make([]json.RawMessage, 0)
			err = json.Unmarshal(bean, &values)
		}
		if err != nil {
			return err
		}
		// 空值处理
		if len(values) == 0 {
			if isNull {
				fieldValue.Set(reflect.ValueOf(nil))
			}
			return nil
		}
		fieldValue.Set(reflect.MakeSlice(reflect.SliceOf(fieldType), len(values), len(values)))
		if isInterface {
			for i, valuei := range values {
				var nextType reflect.Type
				if _, classable := fieldType.MethodByName("ClassName"); classable {
					nextType, err = GenerateTypesFormClass(valuei, classMap)
					//log.Println("set field by class name", nextType.Name())
				} else {
					nextType, err = GenerateTypesFormInterface(fieldType, db)
				}
				if err != nil {
					return err
				}
				fieldValue.Index(i).Set(reflect.New(nextType))
				err = parseObject(nextType, fieldValue.Index(i), valuei, db)
				if err != nil {
					return err
				}
			}
			return nil
		}
		// TODO values to array
		//_ = values
		//bean,_  = json.Marshal(values)
		// 如果已经实现 unmarshal 则直接使用

		for i, valuei := range values {
			nextType := fieldType
			fieldValue.Index(i).Set(reflect.New(nextType).Elem())
			if got, err := checkUnmarshal(fieldValue.Index(i), valuei); got || err != nil {
				if err != nil {
					return err
				}
				continue
			}

			err = parseObject(nextType, fieldValue.Index(i), valuei, db)
			if err != nil {
				return err
			}
		}
		return nil
	// 直接值的解析 ✅
	case reflect.String, reflect.Int, reflect.Bool, reflect.Int64, reflect.Float64, reflect.Float32:
		err := json.Unmarshal(bean, fieldValue.Addr().Interface())
		if err != nil {
			return err
		}
	case reflect.Interface:
		if fieldType.Name() == "Value" {
			var val interface{}
			err = json.Unmarshal(bean, &val)
			if err != nil {
				return err
			}
			newVal, err := ConvertToValue(val)
			if err != nil {
				return err
			}
			fieldValue.Set(reflect.ValueOf(newVal))
			return nil
		}
		return json.Unmarshal(bean, fieldValue.Addr().Interface())
	default:
		return json.Unmarshal(bean, fieldValue.Addr().Interface())
		//return fmt.Errorf("not set values")
	}

	return nil
}

func setFieldDefaultValue(fieldType reflect.Type, fieldValue reflect.Value, defStr string) (err error) {
	var any interface{}
	switch fieldType.Kind() {
	case reflect.String:
		any = defStr
	case reflect.Bool:
		if defStr == "true" {
			any = true
		} else if defStr == "false" {
			any = false
		} else {
			return fmt.Errorf("bool default must be true/false")
		}
	case reflect.Int:
		any, err = strconv.Atoi(defStr)
	case reflect.Int64:
		any, err = strconv.ParseInt(defStr, 0, 64)
	case reflect.Float32:
		var float float64
		float, err = strconv.ParseFloat(defStr, 32)
		any = float32(float)
	case reflect.Float64:
		any, err = strconv.ParseFloat(defStr, 32)
	default:
		return fmt.Errorf("type does not support simple default")
	}
	if err != nil {
		return err
	}
	if reflect.ValueOf(any).Kind() == reflect.String {
		fieldValue.SetString(any.(string))
	} else {
		fieldValue.Set(reflect.ValueOf(any))
	}
	return nil
}

func checkUnmarshal(fieldValue reflect.Value, bean []byte) (bool, error) {
	if _, got := fieldValue.Addr().Type().MethodByName("UnmarshalJSON"); got {
		outvals := fieldValue.Addr().MethodByName("UnmarshalJSON").Call([]reflect.Value{reflect.ValueOf(bean)})
		if len(outvals) != 1 {
			return true, fmt.Errorf("Bad UnmarshalJSON %v", outvals)
		}
		if !outvals[0].IsNil() {
			return true, fmt.Errorf("UnmarshalJSON err %v", outvals)
		}
		return true, nil
	}
	if _, got := fieldValue.Type().MethodByName("UnmarshalJSON"); got {
		outvals := fieldValue.MethodByName("UnmarshalJSON").Call([]reflect.Value{reflect.ValueOf(bean)})
		if len(outvals) != 1 {
			return true, fmt.Errorf("Bad UnmarshalJSON %v", outvals)
		}
		if !outvals[0].IsNil() {
			return true, fmt.Errorf("UnmarshalJSON err %v", outvals)
		}
		return true, nil
	}
	return false, nil
}

type saladTags struct {
	MapSubject   string
	MapPredicate string
	Default      string
	IsType			 bool
}

func getSaladTags(txt string) saladTags {
	s := saladTags{}
	parts := strings.Split(txt, ",")
	for _, p := range parts {
		pp := strings.SplitN(p, ":", 2)
		h := pp[0]
		v := ""
		if len(pp) == 2 {
			v = pp[1]
		}
		switch h {
		case "mapSubject":
			s.MapSubject = v
		case "mapPredicate":
			s.MapPredicate = v
		case "default":
			s.Default = v
		case "type":
			s.IsType = true
		}
	}
	return s
}

func ParseCWLProcess(data []byte) (Process, error) {
	var err error
	trimed := strings.TrimSpace(string(data))
	if len(trimed) == 0 {
		return nil, io.EOF
	}
	if trimed[0] == '#' {
		// 去除脚本解释行
		parts := strings.SplitN(trimed, "\n", 2)
		if len(parts) == 1 {
			return nil, io.EOF
		}
		trimed = parts[1]
	}
	raw := []byte(trimed)
	if trimed[0] != '{' {
		raw, err = Y2J(raw)
		if err != nil {
			return nil, err
		}
	}
	class := &testClass{}
	json.Unmarshal(raw, class)
	if name := class.Class; name != "" {
		var p Process
		switch name {
		case "CommandLineTool":
			p = &CommandLineTool{}
		case "ExpressionTool":
			p = &ExpressionTool{}
		case "Workflow":
			p = &Workflow{}
		case "Operation":
			p = &Operation{}
		default:
			return nil, fmt.Errorf("unknown class for Process %s", name)
		}
		err = json.Unmarshal(raw, p)
		return p, err
	}
	return nil, fmt.Errorf("no Process name")
}

// Only For test
func (p *ProcessBase) UnmarshalJSON_man(data []byte) error {
	type typealias ProcessBase
	var (
		inputs  []InputParameter
		outputs []OutputParameter
	)
	palias := (*typealias)(p)

	bean := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &bean); err != nil {
		return err
	}
	if raw, got := bean["inputs"]; got {
		values, err := JsonldPredicateMapSubject(raw, "id", "type")
		if err != nil {
			return err
		}
		inputs = make([]InputParameter, len(values))
		for i, vali := range values {
			val := InputParameterBase{}
			err = json.Unmarshal(vali, &val)
			if err != nil {
				return err
			}
			inputs[i] = val
		}
		delete(bean, "inputs")
	}
	if raw, got := bean["outputs"]; got {
		values, err := JsonldPredicateMapSubject(raw, "id", "type")
		if err != nil {
			return err
		}
		outputs = make([]OutputParameter, len(values))
		for i, vali := range values {
			val := OutputParameterBase{}
			err = json.Unmarshal(vali, &val)
			if err != nil {
				return err
			}
			outputs[i] = val
		}
		delete(bean, "outputs")
	}
	delete(bean, "outputs")
	delete(bean, "requirements")
	//p.Inputs = make([]InputParameterBase,0)
	data2, _ := json.Marshal(bean)
	palias.Inputs = inputs
	palias.Outputs = outputs
	return json.Unmarshal(data2, palias)
}

func NewBean(db map[string]*RecordFieldGraph, name string) ( interface{}, error) {
	record, got := db[name]
	if got {
		return reflect.New(reflect.TypeOf(record.Example)).Interface() , nil
	}
	return nil, fmt.Errorf("Cannot Generate Type %s", name)
}


func setType(fieldType reflect.Type, fieldValue reflect.Value, data []byte, salad saladTags,
	 db map[string]*RecordFieldGraph) (err error) {
	//fkind := fieldType.Kind()
	//log.Println("setType", fieldType.Name(), fkind.String(), fieldValue.Type().Name(), fieldValue.Interface())
	saladVal :=  fieldValue
	if fieldType.Name() != "SaladType" {
		saladVal = saladVal.FieldByName("SaladType")
	}
	if saladVal.CanAddr() {
		saladVal = saladVal.Addr()
	}
	t := saladVal.Interface().(*SaladType)
	saladType := reflect.TypeOf(SaladType{})
	arrayType := reflect.TypeOf(db["ArrayType"].Example)
	enumType := reflect.TypeOf(db["EnumType"].Example)
	recordType := reflect.TypeOf(db["RecordType"].Example)
	// .. ..
	var bean interface{}
	if err = json.Unmarshal(data,&bean); err != nil {
		return  err
	}
	switch v:= bean.(type) {
	case string:
		isOptional , isArray , restType := typeDSLResolution(v)
		if !isOptional && !isArray {
			t.SetTypename(restType)
			return
		}
		innerType := &SaladType{}
		innerType.SetTypename(restType)
		if isOptional {
			nullType := &SaladType{}
			nullType.SetNull()
			if isArray {
				arrayValue := reflect.New(arrayType)
				array := arrayValue.Interface().(ArrayType)
				array.SetItems(*innerType)
				//t.multi = []SaladType{  {primitive: "null"}, {array: &ArraySchema{Items: innerType}} }
				tmpType := SaladType{}
				tmpType.SetArray(array)
				t.SetMulti([]SaladType{*nullType, tmpType})
				return nil
			}
			t.SetMulti([]SaladType{*nullType, *innerType})
			return nil
		}
		if isArray {
			arrayValue := reflect.New(arrayType)
			array := arrayValue.Interface().(ArrayType)
			array.SetItems(*innerType)
			t.SetArray(array)
			return nil
		}
		return nil
	case map[string]interface{}:
		typenameRaw , got := v["type"]
		if !got {
			return fmt.Errorf("type filed is need for type object")
		}
		typenameStr , got := typenameRaw.(string)
		if !got {
			return fmt.Errorf("type filed need be STRING type for type object")
		}
		switch typenameStr {
		case "record":
			recordValue := reflect.New(recordType)
			record := recordValue.Interface().(RecordType)
			err = parseObject(recordType, recordValue, data, db)
			if err != nil {
				return err
			}
			t.SetRecord(record)
			//return nil
		case "enum":
			enumValue := reflect.New(enumType)
			enum := enumValue.Interface().(EnumType)
			err = parseObject(enumType, enumValue, data, db)
			if err != nil {
				return err
			}
			t.SetEnum(enum)
			//return err
		case "array":
			arrayValue := reflect.New(arrayType)
			array := arrayValue.Interface().(ArrayType)
			//log.Printf("%#v", array)
			err = parseObject(arrayType, arrayValue, data, db)
			if err != nil {
				return err
			}
			t.SetArray(array)
			//return nil
		}
		// 可能有其他字段
		beans := make(map[string]json.RawMessage,0)
		if err = json.Unmarshal(data, &beans); err != nil {
			return err
		}
		delete(beans, "type")
		delete(beans, "items")
		delete(beans, "fields")
		delete(beans, "symbols")
		data ,_ = json.Marshal(beans)
		return setField(fieldType,fieldValue , data, salad, db)
		//return nil
	case []interface{}:
		beans := make([]json.RawMessage,0)
		if err = json.Unmarshal(data, &beans); err != nil {
			return err
		}
		types := make([]SaladType, len(beans))
		for i, beani := range beans {
			err = setType(saladType, reflect.ValueOf(&types[i]), beani, salad, db)
			if err!= nil {
				return err
			}
		}
		t.SetMulti(types)
		return nil
	}
	return fmt.Errorf("unknown type %s", string(data))
}

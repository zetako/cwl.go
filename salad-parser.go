package cwl

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type StringUnmarshalable interface {
	UnmarshalFromString(string) error
}

type Parser struct {
	Name     string
	classMap map[string]interface{}
	//graph map[string]*RecordFieldGraph
	root  *RecordFieldGraph
	salad saladTags
}

type RecordFieldGraph struct {
	Example interface{}
	Fields  map[string]*RecordFieldGraph
	ID      string
}

type saladTags struct {
	MapSubject   string
	MapPredicate string
	Default      string
	IsType       bool
	IsValue      bool
	IsList       bool // 将 单个对象 转换成列表
	IsAbstract   bool
}

func NewParser(root *RecordFieldGraph, classMap map[string]interface{}) *Parser {
	if root == nil {
		root = &RecordFieldGraph{Fields: map[string]*RecordFieldGraph{}}
	}
	if classMap == nil {
		classMap = map[string]interface{}{}
	}
	return &Parser{"", classMap, root, saladTags{}}
}

func (p *Parser) Unmarshal(data []byte, bean interface{}) error {
	typeOfRecv := reflect.TypeOf(bean)
	valueOfRecv := reflect.ValueOf(bean)
	return p.setField(typeOfRecv, valueOfRecv, data, p.salad)
}

func (p *Parser) Fork(fieldname string) *Parser {
	leaf, got := p.root.Fields[fieldname]
	if got && leaf.Fields != nil {
		return &Parser{fieldname, p.classMap, leaf, p.salad}
	}
	return p
}

// setField 解析
func (p *Parser) setField(fieldType reflect.Type, fieldValue reflect.Value, bean []byte, salad saladTags) (err error) {
	//db := p.root.Fields
	fkind := fieldType.Kind()
	//log.Println("setField", fieldType.Name(), fkind.String(), fieldValue.Type().Name(), fieldValue.Interface())
	// 如果本身有解析函数则直接调用 ✅
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
		fieldValue = fieldValue.Elem()

		return p.setField(fieldType, fieldValue, bean, salad)
	case reflect.Struct:
		//debugType(fieldType)
		// 避免死循环, 避免覆盖
		if reflect.TypeOf(p.root.Example).Name() != fieldType.Name() && !salad.IsAbstract {
			if ok, err := checkUnmarshal(fieldValue, bean); ok || err != nil {
				return err
			}
		}
		if salad.IsType {
			return p.parseType(fieldType, fieldValue, bean, salad)
		}
		return p.parseObject(fieldType, fieldValue, bean)
	case reflect.Slice:
		return p.applySlice(fieldType, fieldValue, bean, salad)
	// 直接值的解析 ✅
	case reflect.String, reflect.Int, reflect.Bool, reflect.Int64, reflect.Float64, reflect.Float32:
		//debugType(fieldValue.Type())
		err := json.Unmarshal(bean, fieldValue.Addr().Interface())
		if err != nil {
			return err
		}
	case reflect.Interface:
		return p.applyInterface(fieldType, fieldValue, bean, salad)
	case reflect.Map:
		return p.applyMap(fieldType, fieldValue, bean, salad)
	default:
		return json.Unmarshal(bean, fieldValue.Addr().Interface())
		//return fmt.Errorf("not set values")
	}

	return nil
}

// ✅parseObject
// 根据 salad 扩展 json 结构体的解析
// ✅ 支持 salad.mapSubject 特性
// ✅ 支持 salad.default 特性
//
//	支持 通过 ClassBase 来推断 Interface 的 具体 Struct
//
// ✅ 支持 json:inline 特性
//   - [ ] 优化：避免 map[key]Raw 的重复解析
func (p *Parser) parseObject(typeOfRecv reflect.Type, valueOfRecv reflect.Value, data []byte) (err error) {
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
	if reflect.TypeOf(SaladType{}) == typeOfRecv {
		return nil
	}
	if valueOfRecv.Kind() == reflect.Ptr {
		valueOfRecv = valueOfRecv.Elem()
		//log.Println("pt value", typeOfRecv.Name(), valueOfRecv.Interface(), valueOfRecv.Type().Name())
	}
	//debugType(typeOfRecv)
	//debugValue(valueOfRecv)
	if su, ok := valueOfRecv.Addr().Interface().(StringUnmarshalable); ok {
		if len(data) > 0 && data[0] == '"' {
			var str string
			if err := json.Unmarshal(data, &str); err != nil {
				return err
			}
			return su.UnmarshalFromString(str)
		}
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
		if field.Anonymous {
			inlineFields[keyGo] = key
		}
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
		var salad = p.salad
		field, got := typeOfRecv.FieldByName(keyGo)
		if !got {
			continue
		}
		if valueOfRecv.Kind() == reflect.Interface {
			continue
		}
		if v, got := saladFields[keyGo]; got {
			salad = v
		}

		valueField := valueOfRecv.FieldByName(keyGo)
		//if err = p.parseObject(field.Type, valueField, data); err != nil {
		//  return err
		//}
		if _, got := bean["type"]; salad.IsType && !got {
			continue
		}
		if err = p.setField(field.Type, valueField, data, salad); err != nil {
			return err
		}
	}
	// handle other fields
	for key, value := range bean {
		var salad = p.salad
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
		fieldName := fieldType.Name()
		nextp := p.Fork(fieldName)
		err = nextp.setField(fieldType, fieldValue, value, salad)

		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) applySlice(fieldType reflect.Type, fieldValue reflect.Value, bean []byte, salad saladTags) (err error) {
	fieldType = fieldType.Elem()
	//isInterface := fieldType.Kind() == reflect.Interface
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
	if salad.IsList && len(bean) > 0 && bean[0] != '[' {
		if (bean[0] == '{' && salad.MapSubject == "") || bean[0] == '"' {
			bean = []byte("[" + string(bean) + "]")
		}
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
			//fieldValue.Set(reflect.ValueOf(nil))
		}
		return nil
	}
	//debugType(fieldType)
	if fieldValue.CanSet() {
		fieldValue.Set(reflect.MakeSlice(reflect.SliceOf(fieldType), len(values), len(values)))
	}
	//fieldValue.SetLen(len(values))
	for i, valuei := range values {
		nextType := fieldType
		fieldValue.Index(i).Set(reflect.New(nextType).Elem())
		if got, err := checkUnmarshal(fieldValue.Index(i), valuei); got || err != nil {
			if err != nil {
				return err
			}
			continue
		}

		err = p.setField(nextType, fieldValue.Index(i), valuei, salad)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) applyMap(fieldType reflect.Type, fieldValue reflect.Value, bean []byte, salad saladTags) (err error) {
	//debugValue(fieldValue)
	fieldType = fieldType.Elem()
	//isInterface := fieldType.Kind() == reflect.Interface
	values := map[string]json.RawMessage{}
	// 处理 mapSubject
	// 初始化处理
	isNull := false
	if fieldValue.IsNil() {
		fieldValue.Set(reflect.MakeMap(reflect.MapOf(reflect.TypeOf("string"), fieldType)))
		isNull = true
	}
	//if got, err := checkUnmarshal(fieldValue, bean); got || err != nil {
	//  return err
	//}
	//if salad.IsList && len(bean) >0 && bean[0] != '[' {
	//  if (bean[0] == '{' && salad.MapSubject == "" ) || bean[0] == '"' {
	//    bean = []byte("[" + string(bean) +"]")
	//  }
	//}
	err = json.Unmarshal(bean, &values)
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

	for keyi, valuei := range values {
		nextType := fieldType
		nextValue := reflect.New(nextType).Elem()
		err = p.setField(nextType, nextValue, valuei, salad)
		if err != nil {
			return err
		}
		fieldValue.SetMapIndex(reflect.ValueOf(keyi), nextValue)
	}
	return nil
}

type testClass struct {
	Class string `json:"class"`
}

var noClassError = fmt.Errorf("no class for struct")
var unknownClassError = fmt.Errorf("unknown class for struct")

func (p *Parser) testValueClass(bean []byte) (reflect.Type, error) {
	class := &testClass{}
	if err := json.Unmarshal(bean, class); err != nil {
		return reflect.TypeOf(nil), err
	}
	if name := class.Class; name != "" {
		v, got := p.classMap[name]
		if !got {
			return reflect.TypeOf(nil), unknownClassError
		}
		return reflect.TypeOf(v), nil
	} else {
		return reflect.TypeOf(nil), noClassError
	}
}

func (p *Parser) applyInterface(fieldType reflect.Type, fieldValue reflect.Value, bean []byte, salad saladTags) (err error) {
	var nextType reflect.Type
	//debugType(fieldType)
	fieldTypeName := fieldType.Name()
	nextp := p.Fork(fieldTypeName)
	if p.root.Example != nil && reflect.TypeOf(p.root.Example).Implements(fieldType) && p.Name == fieldType.Name() {
		nextType = reflect.TypeOf(p.root.Example)
	} else if salad.IsValue {
		return p.parseValues(fieldType, fieldValue, bean, salad)
		//if bean[0] == '[' {
		//}
		//nextType, err = p.testValueClass(bean)
		//if err != nil && err != noClassError && err != unknownClassError {
		//  return err
		//}
		//if err != nil {
		//  return json.Unmarshal(bean, fieldValue.Addr().Interface())
		//}
	} else if _, classable := fieldType.MethodByName("ClassName"); classable {
		nextType, err = p.testValueClass(bean)
		if err != nil {
			return err
		}
	} else if record, got := p.root.Fields[fieldTypeName]; got {
		nextType = reflect.TypeOf(record.Example)
	} else {
		// 根据值做推测 (JSON 默认推断)
		return json.Unmarshal(bean, fieldValue.Addr().Interface())
	}
	fieldValue.Set(reflect.New(nextType))
	return nextp.setField(nextType, fieldValue, bean, salad)
}

func (p *Parser) parseType(fieldType reflect.Type, fieldValue reflect.Value, data []byte, salad saladTags) (err error) {
	db := p.root.Fields
	saladVal := fieldValue
	if TypeName := fieldType.Name(); TypeName != "SaladType" {
		if fieldValue.Kind() != reflect.Struct {
			if saladVal.Kind() == reflect.Interface {
				saladVal = saladVal.Elem()
			}
			if saladVal.Kind() == reflect.Ptr {
				saladVal = saladVal.Elem()
			}
		}
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
	salad.IsType = false // 避免重复调用
	// .. ..
	var bean interface{}
	if err = json.Unmarshal(data, &bean); err != nil {
		return err
	}
	switch v := bean.(type) {
	case string:
		isOptional, isArray, restType := typeDSLResolution(v)
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
		typenameRaw, got := v["type"]
		if !got {
			return fmt.Errorf("type filed is need for type object")
		}
		typenameStr, got := typenameRaw.(string)
		if !got {
			return fmt.Errorf("type filed need be STRING type for type object")
		}
		switch typenameStr {
		case "record":
			recordValue := reflect.New(recordType)
			record := recordValue.Interface().(RecordType)
			err = p.setField(recordType, recordValue, data, salad)
			if err != nil {
				return err
			}
			t.SetRecord(record)
			//return nil
		case "enum":
			enumValue := reflect.New(enumType)
			enum := enumValue.Interface().(EnumType)
			err = p.setField(enumType, enumValue, data, salad)
			if err != nil {
				return err
			}
			t.SetEnum(enum)
			//return err
		case "array":
			arrayValue := reflect.New(arrayType)
			array := arrayValue.Interface().(ArrayType)
			//log.Printf("%#v", array)
			err = p.setField(arrayType, arrayValue, data, salad)
			if err != nil {
				return err
			}
			t.SetArray(array)
			//return nil
		}
		// 可能有其他字段
		beans := make(map[string]json.RawMessage, 0)
		if err = json.Unmarshal(data, &beans); err != nil {
			return err
		}
		delete(beans, "type")
		delete(beans, "items")
		delete(beans, "fields")
		delete(beans, "symbols")
		data, _ = json.Marshal(beans)
		return p.parseObject(fieldType, fieldValue, data)
		//return nil
	case []interface{}:
		beans := make([]json.RawMessage, 0)
		if err = json.Unmarshal(data, &beans); err != nil {
			return err
		}
		types := make([]SaladType, len(beans))
		for i, beani := range beans {
			err = p.parseType(saladType, reflect.ValueOf(&types[i]), beani, salad)
			if err != nil {
				return err
			}
		}
		t.SetMulti(types)
		return nil
	}
	return fmt.Errorf("unknown type %s", string(data))
}

func (p *Parser) parseValues(fieldType reflect.Type, fieldValue reflect.Value, data []byte, salad saladTags) (err error) {
	//return nil
	var bean interface{}
	err = json.Unmarshal(data, &bean)
	if err != nil {
		return err
	}
	val, err := ConvertToValue(bean)
	if err != nil {
		return err
	}
	if val == nil {
		return nil
	}
	fieldValue.Set(reflect.ValueOf(val))
	return nil
	if data[0] == '[' {
		nextVal := reflect.ValueOf(make([]Value, 0))
		fieldValue.Set(nextVal)
		return p.applySlice(nextVal.Type(), nextVal, data, salad)
	} else if data[0] == '{' {
		nextType, err := p.testValueClass(data)
		if err == nil {
			fieldValue.Set(reflect.New(nextType))
			return p.setField(fieldValue.Type(), fieldValue, data, saladTags{})
		}
		fieldValue.Set(reflect.ValueOf(map[string]Value{}))
		return p.setField(fieldType, fieldValue, data, salad)
	}
	return json.Unmarshal(data, fieldValue.Addr().Interface())
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
		case "value":
			s.IsValue = true
		case "list":
			s.IsList = true
		case "abstract":
			s.IsAbstract = true
		}
	}
	return s
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
	if unmarshaler, ok := fieldValue.Interface().(json.Unmarshaler); ok {
		err := unmarshaler.UnmarshalJSON(bean)
		return ok, err
	}
	if fieldValue.CanAddr() {
		if unmarshaler, ok := fieldValue.Addr().Interface().(json.Unmarshaler); ok {
			err := unmarshaler.UnmarshalJSON(bean)
			return ok, err
		}
	}
	return false, nil
}

func JsonldPredicateMapSubject(raw []byte, subject, predicate string) ([]json.RawMessage, error) {
	rawArray := make([]json.RawMessage, 0)
	rawMap := make(map[string]json.RawMessage)
	if raw[0] == '[' {
		err := json.Unmarshal(raw, &rawArray)
		return rawArray, err
	}
	err := json.Unmarshal(raw, &rawMap)
	if err != nil {
		return nil, err
	}
	for key, value := range rawMap {
		newObj := make(map[string]interface{})
		if len(value) > 0 && value[0] == '{' {
			err = json.Unmarshal(value, &newObj)
			if err != nil {
				return nil, err
			}
		} else {
			newObj[predicate] = value
		}
		newObj[subject] = key
		newObjRaw, err := json.Marshal(newObj)
		if err != nil {
			return nil, err
		}
		rawArray = append(rawArray, newObjRaw)
	}
	return rawArray, nil
}

func debugType(fieldType reflect.Type) {
	fName := fieldType.Name()
	log.Printf("debug Type Name:%s Kind:%s ", fName, fieldType.Kind())
	if fieldType.Kind() == reflect.Struct {
		for i := 0; i < fieldType.NumField(); i++ {
			fi := fieldType.Field(i)
			log.Printf("F: %d %s %v %v %s", i, fi.Name, fi.Anonymous, fi.Index, fi.PkgPath)
		}
	}
	for i := 0; i < fieldType.NumMethod(); i++ {
		log.Printf("M: %d %s", i, fieldType.Method(i).Name)
	}
	//fieldType.S
	if fieldType.Kind() == reflect.Ptr {
		//debugType(fieldType.Elem())
	}
	if fieldType.Kind() == reflect.Interface {
		//debugType(fieldType.Elem())
	}
	if fieldType.Kind() == reflect.Slice {
		//debugType(fieldType.Elem())
	}
}

func ConvertToValue(bean interface{}) (out Value, err error) {
	beanRef := reflect.ValueOf(bean)
	switch beanRef.Kind() {
	case reflect.Slice, reflect.Array:
		arr := make([]Value, beanRef.Len())
		for i := 0; i < beanRef.Len(); i++ {
			item := beanRef.Index(i).Interface()
			val, err := ConvertToValue(item)
			if err != nil {
				return nil, err
			}
			arr[i] = val
		}
		return arr, nil
	case reflect.Map:
		// 这一步保证Key是string
		keys := beanRef.MapKeys()
		if len(keys) == 0 || keys[0].Kind() != reflect.String {
			return bean, nil
		}
		// 这一步转换map["class"]为string
		var className string = "class" // 这里先设为需要查找的Key来取得Index
		classRef := beanRef.MapIndex(reflect.ValueOf(className))
		if !classRef.IsValid() {
			className = ""
		} else {
			var ok bool
			className, ok = classRef.Interface().(string)
			if !ok {
				className = ""
			}
		}

		// 根据map["class"]来分支
		switch className {
		case "File":
			var entry File
			raw, err := json.Marshal(bean)
			if err != nil {
				return nil, err
			}
			if err = json.Unmarshal(raw, &entry); err != nil {
				return nil, err
			}
			return entry, nil
		case "Directory":
			var entry Directory
			entry.Listing = make([]FileDir, 0)
			raw, err := json.Marshal(bean)
			if err != nil {
				return nil, err
			}
			err = JsonUnmarshal(raw, &entry)
			return entry, nil
		default:
			values := make(map[string]Value)
			iter := beanRef.MapRange()
			for iter.Next() {
				key := iter.Key().String()        // 可以直接转，前面判断过了
				value := iter.Value().Interface() // 转成空接口来提供给函数
				newValue, err := ConvertToValue(value)
				if err != nil {
					return nil, err
				}
				values[key] = newValue
			}
			return values, nil
		}
	default:
		return bean, nil
	}
}

func debugValue(val reflect.Value) {
	log.Printf("debug Value %s %s ", val.Type().Name(), val.Kind())
	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			fi := val.Field(i)
			log.Printf("F: %d %s %v", i, fi.Type().Name(), fi.Interface())
		}
		for i := 0; i < val.Addr().NumMethod(); i++ {
			mi := val.Addr().Method(i)
			log.Printf("*M: %d %s %s", i, mi.String(), mi.Type().Name())
		}
		_, ok := val.Addr().Interface().(StringUnmarshalable)
		log.Printf("StringUnmarshalable %v", ok)

	}
	for i := 0; i < val.NumMethod(); i++ {
		log.Printf("M: %d %s", i, val.Method(i).Type().Name())
	}
}

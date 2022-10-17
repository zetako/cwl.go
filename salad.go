package cwl

import (
	"encoding/json"
	"fmt"
	"strings"
)

type RecordType interface {
	recordType()
}

type EnumType interface {
	enumType()
	GetEnumSchema() *EnumSchema
}

type ArrayType interface {
	arrayType()
	GetItems() *SaladType
	SetItems(SaladType)
}

// TypeBase 的一个基础实例
type SaladType struct {
  name      string // null / boolean / int / long / float / double / string / `definedType`
  primitive string
  //record    *RecordSchema
  //enum      *EnumSchema
  //array      *ArraySchema
  //multi     []SaladType
	record    RecordType
	enum      EnumType
	array     ArrayType
	multi     []SaladType
}

func (SaladType) typebase() {
}

func (t *SaladType) SetTypename(s string) {
	if isPrimitive(s) {
		t.primitive = s
	} else {
		t.name = s
	}
}

func (t *SaladType) SetNull() {
		t.primitive = "null"
}

func (t *SaladType) SetMulti(m []SaladType) {
	t.multi = m
}

func (t *SaladType) SetArray(arrayType ArrayType) {
	t.array = arrayType
}

func (t *SaladType) SetRecord(i RecordType) {
	t.record = i
}
func (t *SaladType) SetEnum(i EnumType) {
	t.enum = i
}


// 4.1.3
type RecordSchema struct {
  Type string `json:"fields"`
  Fields []FieldType `json:"fields"`
}

func (_ RecordSchema) 	recordType() {
}

func (s *RecordSchema) 	Len() int {
	return len(s.Fields)
}

func (s *RecordSchema) 	Index(i int) FieldType {
	return s.Fields[i]
}

type FieldType interface {
  fieldType()
	FieldName() string
	FieldType() *SaladType
}

type RecordField struct {
 Name string `json:"name"`
 Type SaladType `json:"type" salad:"type"`
}



func (_ RecordField) fieldType()  {
}


func (f *RecordField) FieldName() string {
	return f.Name
}

func (f *RecordField) FieldType() *SaladType {
	return &f.Type
}


// 4.1.4.1
type EnumSchema struct {
  Type string // must be enum
  Symbols []string `json:"symbols"`
}


func (_ EnumSchema) enumType() {
}

func (s *EnumSchema) GetEnumSchema() *EnumSchema{
	return s
}


// 4.1.4.2
type ArraySchema struct {
  Type  string    `json:"type"` // must be array
  Items SaladType `json:"items" salad:"type"`
}

func (_ ArraySchema) arrayType() {
}

func (s *ArraySchema) GetItems() *SaladType {
	return &s.Items
}

func (s *ArraySchema) SetItems(b SaladType)  {
	s.Items = b
}

func isPrimitive(v string) bool {
  return v == "null" || v== "boolean" || v == "int" || v == "long" ||
    v == "float" || v == "double" || v == "string"
}

func IsPrimitiveSaladType(v string) bool {
  return v == "null" || v== "boolean" || v == "int" || v == "long" ||
    v == "float" || v == "double" || v == "string"
}
//
//func ParseToType(data []byte, db map[string]*RecordFieldGraph) (t TypeBase, err error){
//	newTypeBase := func() (TypeBase ,error){
//		beanType, err := NewBean(db,"TypeBase")
//		if err != nil {
//			return nil, err
//		}
//		t := beanType.(TypeBase)
//		return t, nil
//	}
//	newTypeArray := func() (ArrayType ,error){
//		beanType, err := NewBean(db,"ArrayType")
//		if err != nil {
//			return nil, err
//		}
//		t := beanType.(ArrayType)
//		return t, nil
//	}
//	newTypeEnum := func() (EnumType ,error){
//		beanType, err := NewBean(db,"EnumType")
//		if err != nil {
//			return nil, err
//		}
//		t := beanType.(EnumType)
//		return t, nil
//	}
//	newTypeRecord := func() (RecordType ,error){
//		beanType, err := NewBean(db,"RecordType")
//		if err != nil {
//			return nil, err
//		}
//		t := beanType.(RecordType)
//		return t, nil
//	}
//	if t , err = newTypeBase(); err != nil {
//		return nil, err
//	}
//	var bean interface{}
//	if err := json.Unmarshal(data,&bean); err != nil {
//		return nil, err
//	}
//	switch v:= bean.(type) {
//	case string:
//		isOptional , isArray , restType := typeDSLResolution(v)
//		t.SetTypename(restType)
//		if !isOptional && !isArray {
//			return
//		}
//		innerType := t
//		{
//			beanType, err := NewBean(db,"TypeBase")
//			if err != nil {
//				return nil, err
//			}
//			t = beanType.(TypeBase)
//		}
//		if isOptional {
//			nullType , _ := newTypeBase()
//			nullType.SetNull()
//			if isArray {
//				array, err := newTypeArray()
//				if err != nil {
//					return nil, err
//				}
//				array.SetItems(innerType)
//				//t.multi = []SaladType{  {primitive: "null"}, {array: &ArraySchema{Items: innerType}} }
//				tmpType ,_ := newTypeBase()
//				tmpType.SetArray(array)
//				t.SetMulti([]TypeBase{nullType, tmpType})
//				return t, nil
//			}
//			t.SetMulti([]TypeBase{nullType, innerType})
//			return t, nil
//		}
//		if isArray {
//			array, err := newTypeArray()
//			if err != nil {
//				return nil, err
//			}
//			array.SetItems(innerType)
//			t.SetArray(array)
//			return t, nil
//		}
//		return innerType, nil
//	case map[string]interface{}:
//		typenameRaw , got := v["type"]
//		if !got {
//			return nil, fmt.Errorf("type filed is need for type object")
//		}
//		typenameStr , got := typenameRaw.(string)
//		if !got {
//			return nil, fmt.Errorf("type filed need be STRING type for type object")
//		}
//		switch typenameStr {
//		case "record":
//			record, err := newTypeRecord()
//			if err != nil {
//				return nil, err
//			}
//			err = json.Unmarshal(data, record)
//			t.SetRecord(record)
//			return t, err
//		case "enum":
//			enum, err := newTypeEnum()
//			if err != nil {
//				return nil, err
//			}
//			err = json.Unmarshal(data, enum)
//			t.SetEnum(enum)
//			return t, err
//		case "array":
//			var array ArrayType
//			array, err = newTypeArray()
//			if err != nil {
//				return nil, err
//			}
//			err = json.Unmarshal(data, array)
//			t.SetArray(array)
//			return t, err
//		}
//	case []interface{}:
//		record, got := db["TypeBase"]
//		if !got {
//			return nil, fmt.Errorf("Cannot Generate Type %s", "TypeBase")
//		}
//		fieldType := reflect.TypeOf(record.Example)
//		multiValue := reflect.MakeSlice(reflect.SliceOf(fieldType), 0, 0)
//		err = json.Unmarshal(data, multiValue.Interface())
//		multi :=make([]TypeBase,multiValue.Len())
//		for i:=0; i< len(multi); i++ {
//			multi[i] = multiValue.Index(i).Interface().(TypeBase)
//		}
//		t.SetMulti(multi)
//		return
//	}
//	return nil, fmt.Errorf("unknown type %s", string(data))
//}
//
//func (t *SaladType) UnmarshalJSON(data []byte) error{
//	db := map[string]*RecordFieldGraph{
//		"TypeBase": &RecordFieldGraph{ Example:  SaladType{} },
//		"ArrayType": &RecordFieldGraph{ Example:  ArraySchema{} },
//		"EnumType": &RecordFieldGraph{ Example:  EnumSchema{} },
//		"RecordType": &RecordFieldGraph{ Example:  RecordSchema{} },
//	}
//	newt , err := ParseToType(data, db)
//	if err != nil {
//		return err
//	}
//	newSal := newt.(*SaladType)
//	t.name, t.primitive, t.enum, t.record, t.array, t.multi =
//		newSal.name, newSal.primitive, newSal.enum, newSal.record, newSal.array, newSal.multi
//	return nil
//  //var bean interface{}
//  //if err := json.Unmarshal(data,&bean); err != nil {
//  //  return err
//  //}
//  //switch v:= bean.(type) {
//  //case string:
//  //  isOptional , isArray , restType := typeDSLResolution(v)
//  //  if isPrimitive(v) {
//  //    t.primitive = restType
//  //  } else {
//  //    t.name = restType
//  //  }
//  //  if !isOptional && !isArray {
//  //    return nil
//  //  }
//  //  innerType := *t
//  //  t.name, t.primitive = "", ""
//  //  if isOptional {
//  //    if isArray {
//  //      t.multi = []SaladType{  {primitive: "null"}, {array: &ArraySchema{Items: innerType}} }
//  //      return nil
//  //    }
//  //    t.multi = []SaladType{  {primitive: "null"}, innerType }
//  //    return nil
//  //  }
//  //  if isArray {
//  //    t.array = &ArraySchema{Items: innerType}
//  //    return nil
//  //  }
//  //  t.name, t.primitive = innerType.name, innerType.primitive
//  //  return nil
//  //case map[string]interface{}:
//  //  typenameRaw , got := v["type"]
//  //  if !got {
//  //    return fmt.Errorf("type filed is need for type object")
//  //  }
//  //  typenameStr , got := typenameRaw.(string)
//  //  if !got {
//  //    return fmt.Errorf("type filed need be STRING type for type object")
//  //  }
//  //  switch typenameStr {
//  //  case "record":
//  //    t.record = &RecordSchema{}
//  //    return json.Unmarshal(data, t.record)
//  //  case "enum":
//  //    t.enum = &EnumSchema{}
//  //    return json.Unmarshal(data, t.enum)
//  //  case "array":
//  //    t.array = &ArraySchema{}
//  //    return json.Unmarshal(data, t.array)
//  //  }
//  //case []interface{}:
//  //  t.multi = make([]SaladType,0)
//  //  return json.Unmarshal(data, &t.multi)
//  //}
//  //return fmt.Errorf("unknown type %s", string(data))
//}

func (t ArraySchema) MarshalJSON()([]byte, error){
  t.Type = "array"
  type rawArray ArraySchema
  return json.Marshal((rawArray)(t))
}

func (t SaladType) MarshalJSON()([]byte, error){
  if t.primitive != "" {
    return json.Marshal(t.primitive)
  } else if t.name != "" {
    return json.Marshal(t.name)
  } else if t.array != nil {
    return json.Marshal(t.array)
  } else if t.enum != nil {
    return json.Marshal(t.enum)
  } else if t.record != nil {
    return json.Marshal(t.record)
  } else if t.multi != nil {
    return json.Marshal(t.multi)
  }
  return nil, fmt.Errorf("invaild type")
}

func (t *SaladType) String() string {
  raw , err := t.MarshalJSON()
  if err != nil {
    return "ErrType:" + err.Error()
  }
  return string(raw)
}

func (t *SaladType) IsPrimitive() bool {
  return t.primitive != ""
}


func (t *SaladType) MustString() string  {
  if t.primitive != "" {
    return t.primitive
  } else if t.name !="" {
    return t.name
  }
  return ""
}

func (t *SaladType) IsNullable() bool {
  if t.primitive == "null" {
    return true
  }
  for i := 0; i < len(t.multi) ; i++{
    if t.multi[i].TypeName() == "null" {
      return true
    }
  }
  return false
}

func (t *SaladType) IsArray() bool {
  if t.array != nil  {
    return true
  }
  return false
}

func (t *SaladType) IsMulti() bool {
	return len(t.multi) != 0
}

func (t *SaladType) Len() int {
	return len(t.multi)
}

func (t *SaladType) Index(i int) SaladType {
	return t.multi[i]
}


func (t *SaladType) MustArraySchema() ArrayType  {
  return t.array
}

func (t *SaladType) MustEnum() EnumType  {
	return t.enum
}


func (t *SaladType) MustMulti() []SaladType  {
  return t.multi
}

func (t *SaladType) MustRecord() RecordType  {
	return t.record
}


func typeDSLResolution(dslType string) (isOptional bool, isArray    bool, restType string) {
  if strings.HasSuffix(dslType, "?") {
    isOptional = true
    dslType = dslType[:len(dslType)-1]
  }
  if strings.HasSuffix(dslType, "[]") {
    isArray = true
    dslType = dslType[:len(dslType)-2]
  }
  return isOptional, isArray, dslType
}

func (t *SaladType) TypeName() string {
  if t.primitive != "" {
    return t.primitive
  } else if t.name !="" {
    return t.name
  } else if t.array != nil {
    return "array"
  } else if t.enum != nil {
    return "enum"
  } else if t.record != nil {
    return "record"
  } else if t.multi != nil {
    types := make([]string, len(t.multi))
    for i, ti := range t.multi {
      types[i] = ti.TypeName()
    }
    return "[" + strings.Join(types, ",") + "]"
  }
  return "_unknownType_"
}

type secondaryFilesDSLPattern struct {
  Pattern string `json:"pattern"`
  Required interface{} `json:"required"`
}

func NewType(name string) SaladType {
  return SaladType{ name: name }
}

func secondaryFilesDSLResolution(in string) (out secondaryFilesDSLPattern) {
  p := secondaryFilesDSLPattern{ Pattern: in, Required: nil }
  if strings.HasSuffix(in, "?") {
    p.Pattern = in[:len(in)-1]
    p.Required = false
  }
  return p
}

func checkPrimitive(typePrim string, v interface{}) bool {
  switch typePrim {
  case "string":
    _, got := v.(string)
    return got
  case "boolean":
    _, got := v.(bool)
    return got
  case "int","long":
    _, got := v.(int64)
    return got
  case "float","double":
    _, got := v.(float64)
    return got
  }
  return false
}

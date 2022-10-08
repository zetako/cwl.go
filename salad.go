package cwl

import (
  "encoding/json"
  "fmt"
  "strings"
)

// 4.1
// primitive or record or enum
// or array of them
type SaladType struct {
  name      string // null / boolean / int / long / float / double / string / `definedType`
  primitive string
  record    *RecordSchema
  enum      *EnumSchema
  array      *ArraySchema
  multi     []SaladType
}

// 4.1.3
type RecordSchema struct {
  Type string
  fields []RecordField
}

// 4.1.4
type RecordField struct {
  Name string
  //Doc ArrayString
  Type SaladType
}

// 4.1.4.1
type EnumSchema struct {
  Type string // must be enum
  Symbols []string
}

// 4.1.4.2
type ArraySchema struct {
  Type  string    `json:"type"` // must be array
  Items SaladType `json:"items"`
}

func isPrimitive(v string) bool {
  return v == "null" || v== "boolean" || v == "int" || v == "long" ||
    v == "float" || v == "double" || v == "string"
}

func IsPrimitiveSaladType(v string) bool {
  return v == "null" || v== "boolean" || v == "int" || v == "long" ||
    v == "float" || v == "double" || v == "string"
}

func (t *SaladType) UnmarshalJSON(data []byte) error{
  var bean interface{}
  if err := json.Unmarshal(data,&bean); err != nil {
    return err
  }
  switch v:= bean.(type) {
  case string:
    isOptional , isArray , restType := typeDSLResolution(v)
    if isPrimitive(v) {
      t.primitive = restType
    } else {
      t.name = restType
    }
    if !isOptional && !isArray {
      return nil
    }
    innerType := *t
    t.name, t.primitive = "", ""
    if isOptional {
      if isArray {
        t.multi = []SaladType{  {primitive: "null"}, {array: &ArraySchema{Items: innerType}} }
        return nil
      }
      t.multi = []SaladType{  {primitive: "null"}, innerType }
      return nil
    }
    if isArray {
      t.array = &ArraySchema{Items: innerType}
      return nil
    }
    t.name, t.primitive = innerType.name, innerType.primitive
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
      t.record = &RecordSchema{}
      return json.Unmarshal(data, t.record)
    case "enum":
      t.enum = &EnumSchema{}
      return json.Unmarshal(data, t.enum)
    case "array":
      t.array = &ArraySchema{}
      return json.Unmarshal(data, t.array)
    }
  case []interface{}:
    t.multi = make([]SaladType,0)
    return json.Unmarshal(data, &t.multi)
  }
  return fmt.Errorf("unknown type %s", string(data))
}

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
  for _, i := range t.multi {
    if i.primitive == "null" {
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


func (t *SaladType) MustArraySchema() *ArraySchema  {
  return t.array
}

func (t *SaladType) MustMulti() []SaladType  {
  return t.multi
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
    return t.array.Type
  } else if t.enum != nil {
    return t.enum.Type
  } else if t.record != nil {
    return t.record.Type
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

// jsonldPredicateMapSubject
// convert { key: obj1 , key2: notObjVal } => [{ sub: key, obj1... },{ sub: key2, predicate: notObjVal }]
func JsonldPredicateMapSubject(raw []byte, subject , predicate string) ([]json.RawMessage, error) {
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
    newObjRaw , err := json.Marshal(newObj)
    if err != nil {
      return nil, err
    }
    rawArray = append(rawArray, newObjRaw)
  }
  return rawArray, nil
}

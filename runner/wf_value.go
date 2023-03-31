package runner

//import (
//	"github.com/lijiang2014/cwl.go"
//	"reflect"
//)
//
//func specifyCwlValuesArray(raw cwl.Values) cwl.Values {
//	for key, value := range raw {
//		if list, ok := value.([]cwl.Value); ok {
//			if len(list) > 0 {
//				arrayType := reflect.TypeOf(list[0])
//				isSame := true
//				for _, entry := range list {
//					if arrayType != reflect.TypeOf(entry) {
//						isSame = false
//						break
//					}
//				}
//			}
//		}
//	}
//}

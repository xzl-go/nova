package binding

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"reflect"
	"strconv"

	"github.com/xzl/nova/pkg/validator"
)

// Binding 参数绑定接口
type Binding interface {
	Name() string
	Bind(*http.Request, interface{}) error
}

// JSONBinding JSON 绑定
type JSONBinding struct{}

func (JSONBinding) Name() string {
	return "json"
}

func (JSONBinding) Bind(req *http.Request, obj interface{}) error {
	if req.Body == nil {
		return errors.New("invalid request")
	}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return validator.ValidateStruct(obj)
}

// XMLBinding XML 绑定
type XMLBinding struct{}

func (XMLBinding) Name() string {
	return "xml"
}

func (XMLBinding) Bind(req *http.Request, obj interface{}) error {
	if req.Body == nil {
		return errors.New("invalid request")
	}
	decoder := xml.NewDecoder(req.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return validator.ValidateStruct(obj)
}

// FormBinding Form 绑定
type FormBinding struct{}

func (FormBinding) Name() string {
	return "form"
}

func (FormBinding) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	return mapForm(obj, req.Form)
}

// QueryBinding Query 绑定
type QueryBinding struct{}

func (QueryBinding) Name() string {
	return "query"
}

func (QueryBinding) Bind(req *http.Request, obj interface{}) error {
	return mapForm(obj, req.URL.Query())
}

// FormPostBinding Form Post 绑定
type FormPostBinding struct{}

func (FormPostBinding) Name() string {
	return "form-urlencoded"
}

func (FormPostBinding) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	return mapForm(obj, req.PostForm)
}

// FormMultipartBinding Form Multipart 绑定
type FormMultipartBinding struct{}

func (FormMultipartBinding) Name() string {
	return "multipart/form-data"
}

func (FormMultipartBinding) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseMultipartForm(defaultMemory); err != nil {
		return err
	}
	return mapForm(obj, req.MultipartForm.Value)
}

// 默认内存大小
const defaultMemory = 32 << 20

// 绑定器映射
var (
	JSON          = JSONBinding{}
	XML           = XMLBinding{}
	Form          = FormBinding{}
	Query         = QueryBinding{}
	FormPost      = FormPostBinding{}
	FormMultipart = FormMultipartBinding{}
)

// mapForm 将表单数据映射到结构体
func mapForm(ptr interface{}, form map[string][]string) error {
	typ := reflect.TypeOf(ptr).Elem()
	val := reflect.ValueOf(ptr).Elem()

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if !structField.CanSet() {
			continue
		}

		structFieldKind := structField.Kind()
		inputFieldName := typeField.Tag.Get("form")
		if inputFieldName == "" {
			inputFieldName = typeField.Name
		}

		inputValue, exists := form[inputFieldName]
		if !exists {
			continue
		}

		numElems := len(inputValue)
		if structFieldKind == reflect.Slice && numElems > 0 {
			sliceOf := structField.Type().Elem().Kind()
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for j := 0; j < numElems; j++ {
				if err := setWithProperType(sliceOf, inputValue[j], slice.Index(j)); err != nil {
					return err
				}
			}
			val.Field(i).Set(slice)
		} else {
			if err := setWithProperType(typeField.Type.Kind(), inputValue[0], structField); err != nil {
				return err
			}
		}
	}
	return nil
}

// setWithProperType 设置适当类型的值
func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	switch valueKind {
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.New("unknown type")
	}
	return nil
}

// setIntField 设置整型字段
func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

// setUintField 设置无符号整型字段
func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

// setBoolField 设置布尔字段
func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

// setFloatField 设置浮点数字段
func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

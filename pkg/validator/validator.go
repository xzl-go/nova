package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// 全局验证器实例
var validate *validator.Validate

func init() {
	validate = validator.New()

	// 注册自定义验证器
	registerCustomValidators()
}

// 注册自定义验证器
func registerCustomValidators() {
	// 手机号验证
	validate.RegisterValidation("mobile", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return regexp.MustCompile(`^1[3-9]\d{9}$`).MatchString(value)
	})

	// 身份证号验证
	validate.RegisterValidation("idcard", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return regexp.MustCompile(`^[1-9]\d{5}(19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`).MatchString(value)
	})

	// 密码强度验证
	validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if len(value) < 8 {
			return false
		}
		var hasUpper, hasLower, hasNumber, hasSpecial bool
		for _, c := range value {
			switch {
			case c >= 'A' && c <= 'Z':
				hasUpper = true
			case c >= 'a' && c <= 'z':
				hasLower = true
			case c >= '0' && c <= '9':
				hasNumber = true
			case c >= '!' && c <= '/' || c >= ':' && c <= '@' || c >= '[' && c <= '`' || c >= '{' && c <= '~':
				hasSpecial = true
			}
		}
		return hasUpper && hasLower && hasNumber && hasSpecial
	})

	// 中文验证
	validate.RegisterValidation("chinese", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return regexp.MustCompile(`^[\u4e00-\u9fa5]+$`).MatchString(value)
	})

	// 英文验证
	validate.RegisterValidation("english", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(value)
	})

	// 数字验证
	validate.RegisterValidation("numeric", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return regexp.MustCompile(`^\d+$`).MatchString(value)
	})

	// 字母数字验证
	validate.RegisterValidation("alphanumeric", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(value)
	})

	// 日期验证
	validate.RegisterValidation("date", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		_, err := time.Parse("2006-01-02", value)
		return err == nil
	})

	// 时间验证
	validate.RegisterValidation("datetime", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		_, err := time.Parse("2006-01-02 15:04:05", value)
		return err == nil
	})

	// IP地址验证
	validate.RegisterValidation("ip", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`).MatchString(value)
	})

	// URL验证
	validate.RegisterValidation("url", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return regexp.MustCompile(`^(http|https)://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(?:/[a-zA-Z0-9\-\._~:/?#[\]@!$&'()*+,;=]*)?$`).MatchString(value)
	})

	// 邮箱验证
	validate.RegisterValidation("email", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(value)
	})

	// 邮政编码验证
	validate.RegisterValidation("postcode", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return regexp.MustCompile(`^\d{6}$`).MatchString(value)
	})

	// 中文姓名验证
	validate.RegisterValidation("chinese_name", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return regexp.MustCompile(`^[\u4e00-\u9fa5]{2,}$`).MatchString(value)
	})

	// 英文姓名验证
	validate.RegisterValidation("english_name", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return regexp.MustCompile(`^[a-zA-Z\s]{2,}$`).MatchString(value)
	})

	// 银行卡号验证
	validate.RegisterValidation("bankcard", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return regexp.MustCompile(`^\d{16,19}$`).MatchString(value)
	})

	// 社会信用代码验证
	validate.RegisterValidation("credit_code", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return regexp.MustCompile(`^[0-9A-HJ-NPQRTUWXY]{2}\d{6}[0-9A-HJ-NPQRTUWXY]{10}$`).MatchString(value)
	})
}

// RegisterValidation 注册自定义验证器
func RegisterValidation(tag string, fn validator.Func) error {
	return validate.RegisterValidation(tag, fn)
}

// ValidateStruct 验证结构体
func ValidateStruct(obj interface{}) error {
	return validate.Struct(obj)
}

// ValidateVar 验证变量
func ValidateVar(field interface{}, tag string) error {
	return validate.Var(field, tag)
}

// GetValidationErrors 获取验证错误信息
func GetValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	if err == nil {
		return errors
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		errors["error"] = err.Error()
		return errors
	}

	for _, e := range validationErrors {
		field := e.Field()
		tag := e.Tag()
		param := e.Param()

		// 获取字段的 json 标签
		t := reflect.TypeOf(e.Value())
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if f, ok := t.FieldByName(field); ok {
			if jsonTag := f.Tag.Get("json"); jsonTag != "" {
				field = strings.Split(jsonTag, ",")[0]
			}
		}

		// 生成错误信息
		message := getErrorMessage(field, tag, param)
		errors[field] = message
	}

	return errors
}

// getErrorMessage 获取错误信息
func getErrorMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s 不能为空", field)
	case "email":
		return fmt.Sprintf("%s 必须是有效的邮箱地址", field)
	case "url":
		return fmt.Sprintf("%s 必须是有效的URL地址", field)
	case "min":
		return fmt.Sprintf("%s 不能小于 %s", field, param)
	case "max":
		return fmt.Sprintf("%s 不能大于 %s", field, param)
	case "len":
		return fmt.Sprintf("%s 长度必须为 %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s 必须是以下值之一: %s", field, param)
	case "unique":
		return fmt.Sprintf("%s 不能重复", field)
	case "alpha":
		return fmt.Sprintf("%s 只能包含字母", field)
	case "numeric":
		return fmt.Sprintf("%s 只能包含数字", field)
	case "alphanumeric":
		return fmt.Sprintf("%s 只能包含字母和数字", field)
	case "datetime":
		return fmt.Sprintf("%s 必须是有效的日期时间格式", field)
	case "mobile":
		return fmt.Sprintf("%s 必须是有效的手机号", field)
	case "idcard":
		return fmt.Sprintf("%s 必须是有效的身份证号", field)
	case "password":
		return fmt.Sprintf("%s 必须包含大小写字母、数字和特殊字符，且长度不少于8位", field)
	case "chinese":
		return fmt.Sprintf("%s 只能包含中文字符", field)
	case "english":
		return fmt.Sprintf("%s 只能包含英文字符", field)
	case "date":
		return fmt.Sprintf("%s 必须是有效的日期格式", field)
	case "ip":
		return fmt.Sprintf("%s 必须是有效的IP地址", field)
	case "postcode":
		return fmt.Sprintf("%s 必须是有效的邮政编码", field)
	case "chinese_name":
		return fmt.Sprintf("%s 必须是有效的中文姓名", field)
	case "english_name":
		return fmt.Sprintf("%s 必须是有效的英文姓名", field)
	case "bankcard":
		return fmt.Sprintf("%s 必须是有效的银行卡号", field)
	case "credit_code":
		return fmt.Sprintf("%s 必须是有效的社会信用代码", field)
	default:
		return fmt.Sprintf("%s 验证失败", field)
	}
}

// 常用验证标签
const (
	Required    = "required"     // 必填
	Email       = "email"        // 邮箱
	URL         = "url"          // URL
	Min         = "min"          // 最小值
	Max         = "max"          // 最大值
	Len         = "len"          // 长度
	OneOf       = "oneof"        // 枚举值
	Unique      = "unique"       // 唯一值
	Alpha       = "alpha"        // 字母
	Numeric     = "numeric"      // 数字
	AlphaNum    = "alphanum"     // 字母和数字
	DateTime    = "datetime"     // 日期时间
	Mobile      = "mobile"       // 手机号
	IDCard      = "idcard"       // 身份证号
	Password    = "password"     // 密码
	Chinese     = "chinese"      // 中文
	English     = "english"      // 英文
	Date        = "date"         // 日期
	IP          = "ip"           // IP地址
	PostCode    = "postcode"     // 邮政编码
	ChineseName = "chinese_name" // 中文姓名
	EnglishName = "english_name" // 英文姓名
	BankCard    = "bankcard"     // 银行卡号
	CreditCode  = "credit_code"  // 社会信用代码
)

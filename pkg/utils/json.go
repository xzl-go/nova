package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// ToJSON 将对象转换为 JSON 字符串
func ToJSON(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ToJSONIndent 将对象转换为格式化的 JSON 字符串
func ToJSONIndent(v interface{}, indent string) (string, error) {
	bytes, err := json.MarshalIndent(v, "", indent)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON 将 JSON 字符串转换为对象
func FromJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

// FromJSONBytes 将 JSON 字节数组转换为对象
func FromJSONBytes(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// FromJSONFile 从文件读取 JSON 并转换为对象
func FromJSONFile(path string, v interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ToJSONFile 将对象转换为 JSON 并写入文件
func ToJSONFile(path string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

// ToJSONFileIndent 将对象转换为格式化的 JSON 并写入文件
func ToJSONFileIndent(path string, v interface{}, indent string) error {
	data, err := json.MarshalIndent(v, "", indent)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

// IsJSON 判断字符串是否为有效的 JSON
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// GetJSONValue 获取 JSON 字符串中指定路径的值
func GetJSONValue(data string, path string) (interface{}, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		return nil, err
	}

	// 解析路径
	keys := strings.Split(path, ".")
	current := v

	for _, key := range keys {
		switch val := current.(type) {
		case map[string]interface{}:
			if value, ok := val[key]; ok {
				current = value
			} else {
				return nil, fmt.Errorf("key not found: %s", key)
			}
		case []interface{}:
			index, err := strconv.Atoi(key)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", key)
			}
			if index < 0 || index >= len(val) {
				return nil, fmt.Errorf("array index out of range: %d", index)
			}
			current = val[index]
		default:
			return nil, fmt.Errorf("invalid path: %s", path)
		}
	}

	return current, nil
}

// SetJSONValue 设置 JSON 字符串中指定路径的值
func SetJSONValue(data string, path string, value interface{}) (string, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		return "", err
	}

	// 解析路径
	keys := strings.Split(path, ".")
	if err := setJSONValueRecursive(&v, keys, value); err != nil {
		return "", err
	}

	result, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// setJSONValueRecursive 递归设置 JSON 路径的值，避免 map index 取地址
func setJSONValueRecursive(current *interface{}, keys []string, value interface{}) error {
	if len(keys) == 0 {
		return nil
	}
	key := keys[0]
	if len(keys) == 1 {
		switch val := (*current).(type) {
		case map[string]interface{}:
			val[key] = value
			return nil
		case []interface{}:
			index, err := strconv.Atoi(key)
			if err != nil {
				return fmt.Errorf("invalid array index: %s", key)
			}
			if index < 0 || index >= len(val) {
				return fmt.Errorf("array index out of range: %d", index)
			}
			val[index] = value
			return nil
		default:
			return fmt.Errorf("invalid path: %s", strings.Join(keys, "."))
		}
	}

	switch val := (*current).(type) {
	case map[string]interface{}:
		next, ok := val[key]
		if !ok {
			m := make(map[string]interface{})
			val[key] = m
			next = m
		}
		return setJSONValueRecursive(&next, keys[1:], value)
	case []interface{}:
		index, err := strconv.Atoi(key)
		if err != nil {
			return fmt.Errorf("invalid array index: %s", key)
		}
		if index < 0 || index >= len(val) {
			return fmt.Errorf("array index out of range: %d", index)
		}
		next := val[index]
		return setJSONValueRecursive(&next, keys[1:], value)
	default:
		return fmt.Errorf("invalid path: %s", strings.Join(keys, "."))
	}
}

// MergeJSON 合并两个 JSON 字符串
func MergeJSON(json1, json2 string) (string, error) {
	var v1, v2 interface{}
	if err := json.Unmarshal([]byte(json1), &v1); err != nil {
		return "", err
	}
	if err := json.Unmarshal([]byte(json2), &v2); err != nil {
		return "", err
	}

	merged := mergeValues(v1, v2)
	result, err := json.Marshal(merged)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// mergeValues 递归合并两个值
func mergeValues(v1, v2 interface{}) interface{} {
	switch val1 := v1.(type) {
	case map[string]interface{}:
		if val2, ok := v2.(map[string]interface{}); ok {
			result := make(map[string]interface{})
			for k, v := range val1 {
				result[k] = v
			}
			for k, v := range val2 {
				if existing, ok := result[k]; ok {
					result[k] = mergeValues(existing, v)
				} else {
					result[k] = v
				}
			}
			return result
		}
	case []interface{}:
		if val2, ok := v2.([]interface{}); ok {
			return append(val1, val2...)
		}
	}
	return v2
}

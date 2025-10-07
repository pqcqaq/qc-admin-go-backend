package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/go-viper/mapstructure/v2"
)

// ToJSON 将对象转换为JSON字符串
func ToJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToJSONBytes 将对象转换为JSON字节切片
func ToJSONBytes(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// ToJSONIndent 将对象转换为格式化的JSON字符串
func ToJSONIndent(v interface{}, indent string) (string, error) {
	data, err := json.MarshalIndent(v, "", indent)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON 从JSON字符串解析对象
func FromJSON(jsonStr string, v interface{}) error {
	return json.Unmarshal([]byte(jsonStr), v)
}

// FromJSONBytes 从JSON字节切片解析对象
func FromJSONBytes(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ReadJSONFile 从文件读取JSON并解析
func ReadJSONFile(filename string, v interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// WriteJSONFile 将对象写入JSON文件
func WriteJSONFile(filename string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

// IsValidJSON 检查字符串是否为有效的JSON
func IsValidJSON(jsonStr string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}

// PrettyJSON 格式化JSON字符串
func PrettyJSON(jsonStr string) (string, error) {
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CompactJSON 压缩JSON字符串（移除空格和换行）
func CompactJSON(jsonStr string) (string, error) {
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return "", err
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MapToStruct 将 map[string]interface{} 转换为指定结构体
func MapToStruct(m map[string]interface{}, out interface{}) error {
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           out,
		TagName:          "json", // 使用 json tag
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return fmt.Errorf("create decoder error: %w", err)
	}

	if err := decoder.Decode(m); err != nil {
		return fmt.Errorf("decode map to struct error: %w", err)
	}

	return nil
}

func MapToString(m map[string]interface{}) string {
	data, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(data)
}

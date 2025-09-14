package database

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	database "go-backend/database/ent"
)

// ExportConfig 导出配置
type ExportConfig struct {
	// OutputDir 输出目录，默认为 "./exports"
	OutputDir string
	// PrettyFormat 是否格式化JSON输出，默认为true
	PrettyFormat bool
	// Context 上下文，用于控制超时等
	Context context.Context
	// ExcludeEntities 排除的实体名称列表
	ExcludeEntities []string
	// IncludeEntities 仅包含的实体名称列表（如果设置，则只导出这些实体）
	IncludeEntities []string
	// ExcludeFields 排除的字段名称列表（支持公共字段如id, create_time等）
	ExcludeFields []string
}

// EntityExportResult 单个实体导出结果
type EntityExportResult struct {
	EntityName  string `json:"entity_name"`
	FilePath    string `json:"file_path"`
	RecordCount int    `json:"record_count"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

// ExportResult 导出结果
type ExportResult struct {
	TotalEntities   int                  `json:"total_entities"`
	SuccessCount    int                  `json:"success_count"`
	FailedCount     int                  `json:"failed_count"`
	OutputDirectory string               `json:"output_directory"`
	Results         []EntityExportResult `json:"results"`
}

// Querier 定义查询接口，所有的 *Client 都应该实现这个接口
type Querier interface {
	All(ctx context.Context) (interface{}, error)
}

// getDefaultExportConfig 获取默认导出配置
func getDefaultExportConfig() *ExportConfig {
	return &ExportConfig{
		OutputDir:    "./exports",
		PrettyFormat: true,
		Context:      context.Background(),
		// 默认排除的公共字段
		ExcludeFields: []string{"id", "create_time", "update_time", "delete_time"},
	}
}

// ExportAllTables 导出所有表的数据为JSON格式
func ExportAllTables(client *database.Client, config *ExportConfig) (*ExportResult, error) {
	if logger != nil {
		logger.Info("开始导出所有表数据...")
	}

	if client == nil {
		if logger != nil {
			logger.Error("数据库客户端为空，无法导出")
		}
		return nil, fmt.Errorf("database client is nil")
	}

	if config == nil {
		config = getDefaultExportConfig()
		if logger != nil {
			logger.Info("使用默认导出配置: 输出目录=%s, 格式化=%v", config.OutputDir, config.PrettyFormat)
		}
	} else {
		if logger != nil {
			logger.Info("使用自定义导出配置: 输出目录=%s, 格式化=%v", config.OutputDir, config.PrettyFormat)
			if len(config.IncludeEntities) > 0 {
				logger.Info("仅导出指定实体: %v", config.IncludeEntities)
			}
			if len(config.ExcludeEntities) > 0 {
				logger.Info("排除实体: %v", config.ExcludeEntities)
			}
			// 排除字段
			if len(config.ExcludeFields) > 0 {
				logger.Info("排除字段: %v", config.ExcludeFields)
			}
		}
	}

	// 确保输出目录存在
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		if logger != nil {
			logger.Error("创建输出目录失败: %v", err)
		}
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	result := &ExportResult{
		OutputDirectory: config.OutputDir,
		Results:         make([]EntityExportResult, 0),
	}

	// 使用反射获取client的所有字段
	clientValue := reflect.ValueOf(client).Elem()
	clientType := clientValue.Type()

	for i := 0; i < clientValue.NumField(); i++ {
		field := clientValue.Field(i)
		fieldType := clientType.Field(i)

		// 跳过非导出字段和非指针字段
		if !field.CanInterface() || field.Kind() != reflect.Ptr {
			continue
		}

		// 获取指针指向的类型名称
		elemType := fieldType.Type.Elem()
		typeName := elemType.Name()

		// 检查类型名是否以"Client"结尾（实体客户端）
		if !strings.HasSuffix(typeName, "Client") {
			continue
		}

		// 获取实体名称（去掉"Client"后缀）
		entityName := strings.TrimSuffix(typeName, "Client")

		// 检查是否在排除列表中
		if isEntityExcluded(entityName, config.ExcludeEntities) {
			continue
		}

		// 检查是否在包含列表中（如果设置了包含列表）
		if len(config.IncludeEntities) > 0 && !isEntityIncluded(entityName, config.IncludeEntities) {
			continue
		}

		result.TotalEntities++

		// 导出单个实体
		entityResult := exportSingleEntity(field.Interface(), entityName, config)
		result.Results = append(result.Results, entityResult)

		if entityResult.Success {
			result.SuccessCount++
			if logger != nil {
				logger.Info("实体 %s 导出成功: %d 条记录", entityName, entityResult.RecordCount)
			}
		} else {
			result.FailedCount++
			if logger != nil {
				logger.Error("实体 %s 导出失败: %s", entityName, entityResult.Error)
			}
		}
	}

	// 记录导出统计信息
	if logger != nil {
		logger.Info("导出完成统计: 总计=%d个实体, 成功=%d个, 失败=%d个, 输出目录=%s",
			result.TotalEntities, result.SuccessCount, result.FailedCount, result.OutputDirectory)
	}

	return result, nil
}

// exportSingleEntity 导出单个实体的数据
func exportSingleEntity(clientInterface interface{}, entityName string, config *ExportConfig) EntityExportResult {
	result := EntityExportResult{
		EntityName: entityName,
		Success:    false,
	}

	// 使用反射调用Query方法
	clientValue := reflect.ValueOf(clientInterface)

	queryMethod := clientValue.MethodByName("Query")

	if !queryMethod.IsValid() {
		result.Error = "Query method not found"
		return result
	}

	// 调用Query方法获取查询对象
	queryResults := queryMethod.Call(nil)
	if len(queryResults) != 1 {
		result.Error = "unexpected Query method signature"
		return result
	}

	queryValue := queryResults[0]

	// 调用All方法获取所有记录
	allMethod := queryValue.MethodByName("All")
	if !allMethod.IsValid() {
		result.Error = "All method not found"
		return result
	}

	// 调用All方法
	ctxValue := reflect.ValueOf(config.Context)
	allResults := allMethod.Call([]reflect.Value{ctxValue})

	if len(allResults) != 2 {
		result.Error = "unexpected All method signature"
		return result
	}

	records := allResults[0].Interface()
	errInterface := allResults[1].Interface()

	if errInterface != nil {
		if err, ok := errInterface.(error); ok {
			result.Error = fmt.Sprintf("failed to query records: %v", err)
			if logger != nil {
				logger.Error("实体 %s: 查询记录失败: %v", entityName, err)
			}
			return result
		}
	}

	// 获取记录数量
	recordsValue := reflect.ValueOf(records)
	if recordsValue.Kind() == reflect.Slice {
		result.RecordCount = recordsValue.Len()
	}

	// 过滤排除的字段
	if len(config.ExcludeFields) > 0 {
		records = filterExcludeFields(records, config.ExcludeFields, entityName)
	}

	// 序列化为JSON
	var jsonData []byte
	var err error

	if config.PrettyFormat {
		jsonData, err = json.MarshalIndent(records, "", "  ")
	} else {
		jsonData, err = json.Marshal(records)
	}

	if err != nil {
		result.Error = fmt.Sprintf("failed to marshal to JSON: %v", err)
		return result
	}

	// 写入文件
	fileName := fmt.Sprintf("%s.json", entityName)
	filePath := filepath.Join(config.OutputDir, fileName)

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		result.Error = fmt.Sprintf("failed to write file: %v", err)
		return result
	}

	result.FilePath = filePath
	result.Success = true

	return result
}

// isEntityExcluded 检查实体是否在排除列表中
func isEntityExcluded(entityName string, excludeList []string) bool {
	for _, excluded := range excludeList {
		if strings.EqualFold(entityName, excluded) {
			return true
		}
	}
	return false
}

// isEntityIncluded 检查实体是否在包含列表中
func isEntityIncluded(entityName string, includeList []string) bool {
	for _, included := range includeList {
		if strings.EqualFold(entityName, included) {
			return true
		}
	}
	return false
}

// ExportAllTablesWithDefaultConfig 使用默认配置导出所有表
func ExportAllTablesWithDefaultConfig(client *database.Client) (*ExportResult, error) {
	return ExportAllTables(client, nil)
}

// ExportSpecificTables 导出指定的表
func ExportSpecificTables(client *database.Client, entityNames []string, outputDir string) (*ExportResult, error) {
	config := &ExportConfig{
		OutputDir:       outputDir,
		PrettyFormat:    true,
		Context:         context.Background(),
		IncludeEntities: entityNames,
	}

	if config.OutputDir == "" {
		config.OutputDir = "./exports"
	}

	return ExportAllTables(client, config)
}

// ExportExcludeTables 导出除了指定表之外的所有表
func ExportExcludeTables(client *database.Client, excludeEntityNames []string, outputDir string) (*ExportResult, error) {
	config := &ExportConfig{
		OutputDir:       outputDir,
		PrettyFormat:    true,
		Context:         context.Background(),
		ExcludeEntities: excludeEntityNames,
	}

	if config.OutputDir == "" {
		config.OutputDir = "./exports"
	}

	return ExportAllTables(client, config)
}

// filterExcludeFields 过滤排除的字段
func filterExcludeFields(records interface{}, excludeFields []string, entityName string) interface{} {
	if len(excludeFields) == 0 {
		return records
	}

	recordsValue := reflect.ValueOf(records)
	if recordsValue.Kind() != reflect.Slice {
		if logger != nil {
			logger.Warn("实体 %s: 无法过滤字段，数据不是切片类型", entityName)
		}
		return records
	}

	// 创建新的切片来存储过滤后的数据
	newSlice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(make(map[string]interface{}))), 0, recordsValue.Len())

	for i := 0; i < recordsValue.Len(); i++ {
		item := recordsValue.Index(i)
		filteredItem := filterSingleRecord(item.Interface(), excludeFields, entityName)
		newSlice = reflect.Append(newSlice, reflect.ValueOf(filteredItem))
	}

	return newSlice.Interface()
}

// filterSingleRecord 过滤单条记录的字段
func filterSingleRecord(record interface{}, excludeFields []string, entityName string) map[string]interface{} {
	result := make(map[string]interface{})

	recordValue := reflect.ValueOf(record)
	recordType := reflect.TypeOf(record)

	// 如果是指针，获取其指向的值
	if recordValue.Kind() == reflect.Ptr {
		if recordValue.IsNil() {
			return result
		}
		recordValue = recordValue.Elem()
		recordType = recordType.Elem()
	}

	// 只处理结构体类型
	if recordValue.Kind() != reflect.Struct {
		if logger != nil {
			logger.Warn("实体 %s: 记录不是结构体类型，无法过滤字段", entityName)
		}
		return result
	}

	// 遍历所有字段
	for i := 0; i < recordValue.NumField(); i++ {
		field := recordValue.Field(i)
		fieldType := recordType.Field(i)

		// 跳过无法导出的字段
		if !field.CanInterface() {
			continue
		}

		// 获取字段的JSON标签名或使用字段名
		jsonTag := fieldType.Tag.Get("json")
		fieldName := fieldType.Name

		if jsonTag != "" && jsonTag != "-" {
			// 解析JSON标签，取第一部分作为字段名
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		// 检查是否应该排除此字段
		shouldExclude := false
		for _, excludeField := range excludeFields {
			if strings.EqualFold(fieldName, excludeField) || strings.EqualFold(fieldType.Name, excludeField) {
				shouldExclude = true
				break
			}
		}

		// 如果不应该排除，则添加到结果中
		if !shouldExclude {
			result[fieldName] = field.Interface()
		}
	}

	return result
}

package database

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	database "go-backend/database/ent"
)

// ImportConfig 导入配置
type ImportConfig struct {
	// InputDir 输入目录，默认为 "./exports"
	InputDir string
	// Context 上下文，用于控制超时等
	Context context.Context
	// ExcludeEntities 排除的实体名称列表
	ExcludeEntities []string
	// IncludeEntities 仅包含的实体名称列表（如果设置，则只导入这些实体）
	IncludeEntities []string
	// BatchSize 批次大小，默认为100
	BatchSize int
	// SkipExisting 是否跳过已存在的记录
	SkipExisting bool
	// ClearBeforeImport 导入前是否清空表
	ClearBeforeImport bool
}

// EntityImportResult 单个实体导入结果
type EntityImportResult struct {
	EntityName   string `json:"entity_name"`
	FilePath     string `json:"file_path"`
	RecordCount  int    `json:"record_count"`
	SuccessCount int    `json:"success_count"`
	FailedCount  int    `json:"failed_count"`
	SkippedCount int    `json:"skipped_count"`
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
}

// ImportResult 导入结果
type ImportResult struct {
	TotalEntities   int                  `json:"total_entities"`
	SuccessCount    int                  `json:"success_count"`
	FailedCount     int                  `json:"failed_count"`
	InputDirectory  string               `json:"input_directory"`
	TotalRecords    int                  `json:"total_records"`
	ImportedRecords int                  `json:"imported_records"`
	SkippedRecords  int                  `json:"skipped_records"`
	Results         []EntityImportResult `json:"results"`
}

// getDefaultImportConfig 获取默认导入配置
func getDefaultImportConfig() *ImportConfig {
	return &ImportConfig{
		InputDir:          "./exports",
		Context:           context.Background(),
		BatchSize:         100,
		SkipExisting:      false,
		ClearBeforeImport: false,
	}
}

// ImportAllTables 从JSON文件导入所有表的数据
func ImportAllTables(client *database.Client, config *ImportConfig) (*ImportResult, error) {
	if logger != nil {
		logger.Info("开始导入所有表数据...")
	}

	if client == nil {
		if logger != nil {
			logger.Error("数据库客户端为空，无法导入")
		}
		return nil, fmt.Errorf("database client is nil")
	}

	if config == nil {
		config = getDefaultImportConfig()
		if logger != nil {
			logger.Info("使用默认导入配置: 输入目录=%s, 批次大小=%d", config.InputDir, config.BatchSize)
		}
	} else {
		if logger != nil {
			logger.Info("使用自定义导入配置: 输入目录=%s, 批次大小=%d", config.InputDir, config.BatchSize)
			if len(config.IncludeEntities) > 0 {
				logger.Info("仅导入指定实体: %v", config.IncludeEntities)
			}
			if len(config.ExcludeEntities) > 0 {
				logger.Info("排除实体: %v", config.ExcludeEntities)
			}
		}
	}

	// 检查输入目录是否存在
	if _, err := os.Stat(config.InputDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("input directory does not exist: %s", config.InputDir)
	}

	result := &ImportResult{
		InputDirectory: config.InputDir,
		Results:        make([]EntityImportResult, 0),
	}

	// 开始数据库事务
	tx, err := client.Tx(config.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// 禁用外键约束检查（PostgreSQL）
	if _, err := tx.ExecContext(config.Context, "SET session_replication_role = replica;"); err != nil {
		if logger != nil {
			logger.Warn("禁用外键约束失败: %v", err)
		}
	} else {
		if logger != nil {
			logger.Info("已禁用外键约束检查")
		}
	}

	defer func() {
		if err != nil {
			// 在回滚前尝试重新启用外键约束
			if _, enableErr := tx.ExecContext(config.Context, "SET session_replication_role = DEFAULT;"); enableErr != nil {
				if logger != nil {
					logger.Warn("回滚时重新启用外键约束失败: %v", enableErr)
				}
			}

			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				if logger != nil {
					logger.Error("事务回滚失败: %v", rollbackErr)
				}
			}
		}
	}()

	// 读取目录中的所有JSON文件
	files, err := filepath.Glob(filepath.Join(config.InputDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to find JSON files: %w", err)
	}

	if len(files) == 0 {
		if logger != nil {
			logger.Warn("在目录 %s 中没有找到JSON文件", config.InputDir)
		}
		return result, nil
	}

	// 使用反射获取client的所有字段
	clientValue := reflect.ValueOf(client).Elem()
	clientType := clientValue.Type()

	// 创建实体客户端映射
	entityClients := make(map[string]reflect.Value)
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
		entityClients[entityName] = field
	}

	// 处理每个JSON文件
	for _, filePath := range files {
		fileName := filepath.Base(filePath)
		entityName := strings.TrimSuffix(fileName, ".json")

		// 检查是否在排除列表中
		if isEntityExcluded(entityName, config.ExcludeEntities) {
			continue
		}

		// 检查是否在包含列表中（如果设置了包含列表）
		if len(config.IncludeEntities) > 0 && !isEntityIncluded(entityName, config.IncludeEntities) {
			continue
		}

		result.TotalEntities++

		// 导入单个实体
		entityResult := importSingleEntity(tx, filePath, entityName, entityClients, config)
		result.Results = append(result.Results, entityResult)

		result.TotalRecords += entityResult.RecordCount
		result.ImportedRecords += entityResult.SuccessCount
		result.SkippedRecords += entityResult.SkippedCount

		if entityResult.Success {
			result.SuccessCount++
			if logger != nil {
				logger.Info("实体 %s 导入成功: 总计=%d条, 成功=%d条, 跳过=%d条, 失败=%d条",
					entityName, entityResult.RecordCount, entityResult.SuccessCount,
					entityResult.SkippedCount, entityResult.FailedCount)
			}
		} else {
			result.FailedCount++
			if logger != nil {
				logger.Error("实体 %s 导入失败: %s", entityName, entityResult.Error)
			}
			return result, fmt.Errorf("entity %s import failed: %s", entityName, entityResult.Error)
		}
	}

	// 重新启用外键约束检查（在提交事务之前）
	if _, err := tx.ExecContext(config.Context, "SET session_replication_role = DEFAULT;"); err != nil {
		if logger != nil {
			logger.Warn("重新启用外键约束失败: %v", err)
		}
	} else {
		if logger != nil {
			logger.Info("已重新启用外键约束检查")
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 记录导入统计信息
	if logger != nil {
		logger.Info("导入完成统计: 总计=%d个实体, 成功=%d个, 失败=%d个, 总记录=%d条, 导入记录=%d条, 跳过记录=%d条",
			result.TotalEntities, result.SuccessCount, result.FailedCount,
			result.TotalRecords, result.ImportedRecords, result.SkippedRecords)
	}

	return result, nil
}

// importSingleEntity 导入单个实体的数据
func importSingleEntity(tx *database.Tx, filePath, entityName string, entityClients map[string]reflect.Value, config *ImportConfig) EntityImportResult {
	result := EntityImportResult{
		EntityName: entityName,
		FilePath:   filePath,
		Success:    false,
	}

	// 读取JSON文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to read file: %v", err)
		return result
	}

	// 解析JSON数据为通用map切片
	var records []map[string]interface{}
	if err := json.Unmarshal(data, &records); err != nil {
		result.Error = fmt.Sprintf("failed to unmarshal JSON: %v", err)
		return result
	}

	result.RecordCount = len(records)

	if result.RecordCount == 0 {
		result.Success = true
		if logger != nil {
			logger.Info("实体 %s: 文件为空，跳过导入", entityName)
		}
		return result
	}

	// 查找对应的实体客户端
	clientField, exists := entityClients[entityName]
	if !exists {
		result.Error = fmt.Sprintf("entity client not found for: %s", entityName)
		return result
	}

	// 如果配置了清空表，先删除所有记录
	if config.ClearBeforeImport {
		if err := clearEntityTable(tx, clientField, entityName); err != nil {
			result.Error = fmt.Sprintf("failed to clear table: %v", err)
			return result
		}
		if logger != nil {
			logger.Info("实体 %s: 表已清空", entityName)
		}
	}

	// 批量导入记录
	batchSize := config.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		successCount, skippedCount, err := importBatch(tx, clientField, entityName, batch, config)

		result.SuccessCount += successCount
		result.SkippedCount += skippedCount

		if err != nil {
			result.FailedCount += len(batch) - successCount - skippedCount
			result.Error = fmt.Sprintf("batch import failed: %v", err)
			return result
		}
	}

	result.Success = true
	return result
}

// importBatch 批量导入记录
func importBatch(tx *database.Tx, clientField reflect.Value, entityName string, batch []map[string]interface{}, config *ImportConfig) (successCount, skippedCount int, err error) {
	// 获取事务中的客户端
	txValue := reflect.ValueOf(tx)
	// 直接使用实体名称，不需要转换
	txClientField := txValue.Elem().FieldByName(entityName)

	if !txClientField.IsValid() {
		return 0, 0, fmt.Errorf("transaction client field not found for entity: %s", entityName)
	}

	// 调用CreateBulk方法
	createBulkMethod := txClientField.MethodByName("CreateBulk")
	if !createBulkMethod.IsValid() {
		return 0, 0, fmt.Errorf("CreateBulk method not found for entity: %s", entityName)
	}

	// 准备创建器切片
	creators := make([]reflect.Value, 0, len(batch))

	for _, record := range batch {
		// 调用Create方法获取创建器
		createMethod := txClientField.MethodByName("Create")
		if !createMethod.IsValid() {
			return 0, 0, fmt.Errorf("create method not found for entity: %s", entityName)
		}

		creatorResults := createMethod.Call(nil)
		if len(creatorResults) != 1 {
			return 0, 0, fmt.Errorf("unexpected Create method signature for entity: %s", entityName)
		}

		creator := creatorResults[0]

		// 设置字段值
		if err := setCreatorFields(creator, record, entityName); err != nil {
			if config.SkipExisting {
				skippedCount++
				continue
			}
			return successCount, skippedCount, fmt.Errorf("failed to set fields for entity %s: %v", entityName, err)
		}

		creators = append(creators, creator)
	}

	if len(creators) == 0 {
		return successCount, skippedCount, nil
	}

	// 调用CreateBulk方法 - ent的CreateBulk接受可变参数，不是切片
	bulkResults := createBulkMethod.Call(creators)
	if len(bulkResults) != 1 {
		return successCount, skippedCount, fmt.Errorf("unexpected CreateBulk method signature for entity: %s", entityName)
	}

	bulk := bulkResults[0]

	// 调用Save方法执行批量插入
	saveMethod := bulk.MethodByName("Save")
	if !saveMethod.IsValid() {
		return successCount, skippedCount, fmt.Errorf("save method not found on bulk creator for entity: %s", entityName)
	}

	ctxValue := reflect.ValueOf(config.Context)
	saveResults := saveMethod.Call([]reflect.Value{ctxValue})

	if len(saveResults) != 2 {
		return successCount, skippedCount, fmt.Errorf("unexpected Save method signature for entity: %s", entityName)
	}

	errInterface := saveResults[1].Interface()
	if errInterface != nil {
		if err, ok := errInterface.(error); ok {
			return successCount, skippedCount, fmt.Errorf("bulk save failed for entity %s: %v", entityName, err)
		}
	}

	successCount += len(creators)
	return successCount, skippedCount, nil
}

// setCreatorFields 设置创建器的字段值
func setCreatorFields(creator reflect.Value, record map[string]interface{}, entityName string) error {
	for fieldName, value := range record {
		// 跳过edges字段
		if fieldName == "edges" {
			continue
		}

		// 查找匹配的Set方法
		setMethod := findSetMethod(creator, fieldName)
		if !setMethod.IsValid() {
			// 如果找不到Set方法，记录警告但不报错
			if logger != nil {
				logger.Warn("实体 %s: 找不到字段 %s 的Set方法", entityName, fieldName)
			}
			continue
		}

		// 获取方法的参数类型
		methodType := setMethod.Type()
		if methodType.NumIn() != 1 {
			if logger != nil {
				logger.Warn("实体 %s: Set方法参数数量不正确", entityName)
			}
			continue
		}

		paramType := methodType.In(0)

		// 转换值类型
		convertedValue, err := convertValue(value, paramType)
		if err != nil {
			if logger != nil {
				logger.Warn("实体 %s: 字段 %s 值类型转换失败: %v", entityName, fieldName, err)
			}
			continue
		}

		// 调用Set方法
		setResults := setMethod.Call([]reflect.Value{convertedValue})
		if len(setResults) != 1 {
			return fmt.Errorf("unexpected Set method signature for field %s in entity %s", fieldName, entityName)
		}

		// 更新creator引用
		creator = setResults[0]
	}

	return nil
}

// findSetMethod 查找匹配的Set方法
func findSetMethod(creator reflect.Value, fieldName string) reflect.Value {
	creatorType := creator.Type()

	// 获取所有方法
	for i := 0; i < creatorType.NumMethod(); i++ {
		method := creatorType.Method(i)
		methodName := method.Name

		// 检查是否是Set方法
		if !strings.HasPrefix(methodName, "Set") {
			continue
		}

		// 提取方法名中的字段名部分（去掉"Set"前缀）
		methodFieldName := strings.TrimPrefix(methodName, "Set")

		// 将字段名和方法字段名都转换为小写进行比较
		if strings.EqualFold(methodFieldName, strings.ReplaceAll(fieldName, "_", "")) {
			return creator.MethodByName(methodName)
		}

		// 也尝试直接匹配（处理下划线转驼峰的情况）
		camelCaseFieldName := toCamelCase(fieldName)
		if strings.EqualFold(methodFieldName, camelCaseFieldName) {
			return creator.MethodByName(methodName)
		}
	}

	return reflect.Value{}
}

// convertValue 转换值类型以匹配目标类型
func convertValue(value interface{}, targetType reflect.Type) (reflect.Value, error) {
	if value == nil {
		return reflect.Zero(targetType), nil
	}

	sourceValue := reflect.ValueOf(value)
	sourceType := sourceValue.Type()

	// 如果类型已经匹配
	if sourceType == targetType {
		return sourceValue, nil
	}

	// 如果目标类型是指针类型
	if targetType.Kind() == reflect.Ptr {
		elemType := targetType.Elem()
		convertedElem, err := convertValue(value, elemType)
		if err != nil {
			return reflect.Zero(targetType), err
		}

		ptr := reflect.New(elemType)
		ptr.Elem().Set(convertedElem)
		return ptr, nil
	}

	// 处理数字类型转换
	if sourceType.Kind() == reflect.Float64 && isIntegerType(targetType) {
		floatVal := value.(float64)
		switch targetType.Kind() {
		case reflect.Int:
			return reflect.ValueOf(int(floatVal)), nil
		case reflect.Int8:
			return reflect.ValueOf(int8(floatVal)), nil
		case reflect.Int16:
			return reflect.ValueOf(int16(floatVal)), nil
		case reflect.Int32:
			return reflect.ValueOf(int32(floatVal)), nil
		case reflect.Int64:
			return reflect.ValueOf(int64(floatVal)), nil
		case reflect.Uint:
			return reflect.ValueOf(uint(floatVal)), nil
		case reflect.Uint8:
			return reflect.ValueOf(uint8(floatVal)), nil
		case reflect.Uint16:
			return reflect.ValueOf(uint16(floatVal)), nil
		case reflect.Uint32:
			return reflect.ValueOf(uint32(floatVal)), nil
		case reflect.Uint64:
			return reflect.ValueOf(uint64(floatVal)), nil
		}
	}

	// 处理时间类型转换
	if sourceType.Kind() == reflect.String && targetType.String() == "time.Time" {
		timeStr := value.(string)
		// 尝试多种时间格式
		timeFormats := []string{
			"2006-01-02T15:04:05.999999Z07:00", // RFC3339Nano
			"2006-01-02T15:04:05Z07:00",        // RFC3339
			"2006-01-02T15:04:05.999999Z",      // UTC Nano
			"2006-01-02T15:04:05Z",             // UTC
			"2006-01-02 15:04:05.999999",       // SQL Nano
			"2006-01-02 15:04:05",              // SQL
			"2006-01-02",                       // Date only
		}

		for _, format := range timeFormats {
			if parsedTime, err := time.Parse(format, timeStr); err == nil {
				return reflect.ValueOf(parsedTime), nil
			}
		}

		return reflect.Zero(targetType), fmt.Errorf("cannot parse time string %s", timeStr)
	}

	// 尝试类型转换
	if sourceValue.Type().ConvertibleTo(targetType) {
		return sourceValue.Convert(targetType), nil
	}

	return reflect.Zero(targetType), fmt.Errorf("cannot convert %v (type %v) to %v", value, sourceType, targetType)
}

// isIntegerType 检查是否为整数类型
func isIntegerType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

// toCamelCase 转换为驼峰命名
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.Title(parts[i])
		}
	}
	return strings.Join(parts, "")
}

// clearEntityTable 清空实体表
func clearEntityTable(tx *database.Tx, _ reflect.Value, entityName string) error {
	// 获取事务中的客户端
	txValue := reflect.ValueOf(tx)
	// 直接使用实体名称，不需要转换
	txClientField := txValue.Elem().FieldByName(entityName)

	if !txClientField.IsValid() {
		return fmt.Errorf("transaction client field not found for entity: %s", entityName)
	} // 调用Delete方法
	deleteMethod := txClientField.MethodByName("Delete")
	if !deleteMethod.IsValid() {
		return fmt.Errorf("delete method not found for entity: %s", entityName)
	}

	deleteResults := deleteMethod.Call(nil)
	if len(deleteResults) != 1 {
		return fmt.Errorf("unexpected Delete method signature for entity: %s", entityName)
	}

	deleter := deleteResults[0]

	// 调用Where方法 (传入空条件表示删除所有)
	whereMethod := deleter.MethodByName("Where")
	if !whereMethod.IsValid() {
		// 如果没有Where方法，直接调用Exec
		execMethod := deleter.MethodByName("Exec")
		if !execMethod.IsValid() {
			return fmt.Errorf("exec method not found on deleter for entity: %s", entityName)
		}

		ctxValue := reflect.ValueOf(context.Background())
		execResults := execMethod.Call([]reflect.Value{ctxValue})

		if len(execResults) != 2 {
			return fmt.Errorf("unexpected Exec method signature for entity: %s", entityName)
		}

		errInterface := execResults[1].Interface()
		if errInterface != nil {
			if err, ok := errInterface.(error); ok {
				return fmt.Errorf("failed to clear table for entity %s: %v", entityName, err)
			}
		}
	}

	return nil
}

// ImportAllTablesWithDefaultConfig 使用默认配置导入所有表
func ImportAllTablesWithDefaultConfig(client *database.Client) (*ImportResult, error) {
	return ImportAllTables(client, nil)
}

// ImportSpecificTables 导入指定的表
func ImportSpecificTables(client *database.Client, entityNames []string, inputDir string) (*ImportResult, error) {
	config := &ImportConfig{
		InputDir:        inputDir,
		Context:         context.Background(),
		BatchSize:       100,
		SkipExisting:    false,
		IncludeEntities: entityNames,
	}

	if config.InputDir == "" {
		config.InputDir = "./exports"
	}

	return ImportAllTables(client, config)
}

// ImportExcludeTables 导入除了指定表之外的所有表
func ImportExcludeTables(client *database.Client, excludeEntityNames []string, inputDir string) (*ImportResult, error) {
	config := &ImportConfig{
		InputDir:        inputDir,
		Context:         context.Background(),
		BatchSize:       100,
		SkipExisting:    false,
		ExcludeEntities: excludeEntityNames,
	}

	if config.InputDir == "" {
		config.InputDir = "./exports"
	}

	return ImportAllTables(client, config)
}

// ImportAllTablesGlobal 使用全局客户端导入所有表数据
func ImportAllTablesGlobal(inputDir string) (*ImportResult, error) {
	if Client == nil {
		return nil, fmt.Errorf("database client is not initialized, call InitInstance first")
	}

	config := &ImportConfig{
		InputDir:          inputDir,
		Context:           context.Background(),
		BatchSize:         100,
		SkipExisting:      false,
		ClearBeforeImport: false,
	}

	if config.InputDir == "" {
		config.InputDir = "./exports"
	}

	return ImportAllTables(Client, config)
}

// ImportSpecificTablesGlobal 使用全局客户端导入指定表
func ImportSpecificTablesGlobal(entityNames []string, inputDir string) (*ImportResult, error) {
	if Client == nil {
		return nil, fmt.Errorf("database client is not initialized, call InitInstance first")
	}

	return ImportSpecificTables(Client, entityNames, inputDir)
}

package excel

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

// ColumnConfig 列配置
type ColumnConfig struct {
	Header    string           // 列标题
	Width     float64          // 列宽度
	Formatter func(any) string // 格式化函数
	FieldName string           // 对应的字段名
}

// ExcelProcessor Excel处理器
type ExcelProcessor struct {
	SheetName string
	Columns   []ColumnConfig
}

// NewExcelProcessor 创建新的Excel处理器
func NewExcelProcessor(sheetName string, columns []ColumnConfig) *ExcelProcessor {
	return &ExcelProcessor{
		SheetName: sheetName,
		Columns:   columns,
	}
}

// GenerateExcelStream 生成Excel流
func (p *ExcelProcessor) GenerateExcelStream(data any) (*excelize.File, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// 设置工作表名称
	index, err := f.NewSheet(p.SheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)

	// 删除默认的Sheet1
	err = f.DeleteSheet("Sheet1")
	if err != nil {
		return nil, fmt.Errorf("failed to delete default sheet: %w", err)
	}

	// 写入表头
	for i, col := range p.Columns {
		cell := fmt.Sprintf("%s1", columnNumberToLetter(i+1))
		f.SetCellValue(p.SheetName, cell, col.Header)

		// 设置列宽
		if col.Width > 0 {
			colLetter := columnNumberToLetter(i + 1)
			f.SetColWidth(p.SheetName, colLetter, colLetter, col.Width)
		}
	}

	// 设置表头样式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E0E0E0"},
			Pattern: 1,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create header style: %w", err)
	}

	err = f.SetCellStyle(p.SheetName, "A1", fmt.Sprintf("%s1", columnNumberToLetter(len(p.Columns))), headerStyle)
	if err != nil {
		return nil, fmt.Errorf("failed to set header style: %w", err)
	}

	// 写入数据
	err = p.writeData(f, data)
	if err != nil {
		return nil, fmt.Errorf("failed to write data: %w", err)
	}

	return f, nil
}

// writeData 写入数据
func (p *ExcelProcessor) writeData(f *excelize.File, data any) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Slice {
		return fmt.Errorf("data must be a slice")
	}

	for i := 0; i < val.Len(); i++ {
		rowData := val.Index(i)
		if rowData.Kind() == reflect.Ptr {
			rowData = rowData.Elem()
		}

		row := i + 2 // 从第二行开始写入数据（第一行是表头）

		for j, col := range p.Columns {
			cell := fmt.Sprintf("%s%d", columnNumberToLetter(j+1), row)

			// 获取字段值
			fieldValue := p.getFieldValue(rowData, col.FieldName)

			// 格式化值
			var cellValue string
			if col.Formatter != nil {
				cellValue = col.Formatter(fieldValue)
			} else {
				cellValue = p.defaultFormatter(fieldValue)
			}

			f.SetCellValue(p.SheetName, cell, cellValue)
		}
	}

	return nil
}

// getFieldValue 获取字段值
func (p *ExcelProcessor) getFieldValue(data reflect.Value, fieldName string) any {
	if data.Kind() != reflect.Struct {
		return nil
	}

	field := data.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}

	return field.Interface()
}

// defaultFormatter 默认格式化器
func (p *ExcelProcessor) defaultFormatter(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%.2f", v)
	case bool:
		if v {
			return "是"
		}
		return "否"
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// columnNumberToLetter 将列号转换为字母
func columnNumberToLetter(col int) string {
	result := ""
	for col > 0 {
		col--
		result = string(rune('A'+col%26)) + result
		col /= 26
	}
	return result
}

// StreamResponse Excel流响应结构
type StreamResponse struct {
	File     *excelize.File
	Filename string
}

// GenerateFilename 生成文件名
func GenerateFilename(prefix string) string {
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s_%s.xlsx", prefix, timestamp)
}

// TimeFormatter 时间格式化器
func TimeFormatter(layout string) func(any) string {
	return func(value any) string {
		if t, ok := value.(time.Time); ok {
			return t.Format(layout)
		}
		return ""
	}
}

// BoolFormatter 布尔值格式化器
func BoolFormatter(trueText, falseText string) func(any) string {
	return func(value any) string {
		if b, ok := value.(bool); ok {
			if b {
				return trueText
			}
			return falseText
		}
		return ""
	}
}

// NumberFormatter 数字格式化器
func NumberFormatter(precision int) func(any) string {
	return func(value any) string {
		switch v := value.(type) {
		case float32:
			return strconv.FormatFloat(float64(v), 'f', precision, 32)
		case float64:
			return strconv.FormatFloat(v, 'f', precision, 64)
		case int, int8, int16, int32, int64:
			return fmt.Sprintf("%d", v)
		case uint, uint8, uint16, uint32, uint64:
			return fmt.Sprintf("%d", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
}

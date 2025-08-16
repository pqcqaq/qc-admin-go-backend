package excel

import (
	"testing"
	"time"
)

// TestData 测试数据结构
type TestData struct {
	ID         string
	Content    string
	Success    bool
	CreateTime time.Time
	ImageId    string
	ImageUrl   string
}

func TestExcelProcessor(t *testing.T) {
	// 创建测试数据
	testData := []*TestData{
		{
			ID:         "1",
			Content:    "测试扫描内容1",
			Success:    true,
			CreateTime: time.Now(),
			ImageId:    "img1",
			ImageUrl:   "http://example.com/img1.jpg",
		},
		{
			ID:         "2",
			Content:    "测试扫描内容2",
			Success:    false,
			CreateTime: time.Now().Add(-time.Hour),
			ImageId:    "img2",
			ImageUrl:   "http://example.com/img2.jpg",
		},
	}

	// 配置列
	columns := []ColumnConfig{
		{
			Header:    "ID",
			Width:     15,
			FieldName: "ID",
		},
		{
			Header:    "扫描内容",
			Width:     40,
			FieldName: "Content",
		},
		{
			Header:    "是否成功",
			Width:     15,
			FieldName: "Success",
			Formatter: BoolFormatter("成功", "失败"),
		},
		{
			Header:    "创建时间",
			Width:     25,
			FieldName: "CreateTime",
			Formatter: TimeFormatter("2006-01-02 15:04:05"),
		},
		{
			Header:    "图片ID",
			Width:     15,
			FieldName: "ImageId",
		},
		{
			Header:    "图片URL",
			Width:     50,
			FieldName: "ImageUrl",
		},
	}

	// 创建处理器
	processor := NewExcelProcessor("测试数据", columns)

	// 生成Excel文件
	file, err := processor.GenerateExcelStream(testData)
	if err != nil {
		t.Fatalf("生成Excel文件失败: %v", err)
	}

	if file == nil {
		t.Fatal("生成的Excel文件为空")
	}

	t.Log("Excel处理器测试通过")
}

func TestGenerateFilename(t *testing.T) {
	filename := GenerateFilename("测试")
	if filename == "" {
		t.Fatal("生成的文件名为空")
	}
	t.Logf("生成的文件名: %s", filename)
}

func TestColumnNumberToLetter(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1, "A"},
		{2, "B"},
		{26, "Z"},
		{27, "AA"},
		{28, "AB"},
	}

	for _, test := range tests {
		result := columnNumberToLetter(test.input)
		if result != test.expected {
			t.Errorf("columnNumberToLetter(%d) = %s, 期望 %s", test.input, result, test.expected)
		}
	}
}

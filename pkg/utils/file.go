package utils

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Exists 检查文件或目录是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsFile 检查路径是否为文件
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// IsDir 检查路径是否为目录
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ReadFile 读取文件内容为字符串
func ReadFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadFileBytes 读取文件内容为字节切片
func ReadFileBytes(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// WriteFile 将字符串写入文件
func WriteFile(path, content string) error {
	return ioutil.WriteFile(path, []byte(content), 0644)
}

// WriteFileBytes 将字节切片写入文件
func WriteFileBytes(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0644)
}

// AppendFile 向文件末尾追加内容
func AppendFile(path, content string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// ReadLines 按行读取文件
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// WriteLines 按行写入文件
func WriteLines(path string, lines []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// MoveFile 移动文件
func MoveFile(src, dst string) error {
	return os.Rename(src, dst)
}

// DeleteFile 删除文件
func DeleteFile(path string) error {
	return os.Remove(path)
}

// MkdirAll 创建目录（包括父目录）
func MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

// RemoveAll 递归删除目录及其内容
func RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// GetFileSize 获取文件大小
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFileExt 获取文件扩展名
func GetFileExt(path string) string {
	return filepath.Ext(path)
}

// GetFileName 获取文件名（不含路径）
func GetFileName(path string) string {
	return filepath.Base(path)
}

// GetFileNameWithoutExt 获取不含扩展名的文件名
func GetFileNameWithoutExt(path string) string {
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

// GetDir 获取文件所在目录
func GetDir(path string) string {
	return filepath.Dir(path)
}

// JoinPath 连接路径
func JoinPath(elements ...string) string {
	return filepath.Join(elements...)
}

// AbsPath 获取绝对路径
func AbsPath(path string) (string, error) {
	return filepath.Abs(path)
}

// ListFiles 列出目录中的所有文件
func ListFiles(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var filenames []string
	for _, file := range files {
		if !file.IsDir() {
			filenames = append(filenames, file.Name())
		}
	}

	return filenames, nil
}

// ListDirs 列出目录中的所有子目录
func ListDirs(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var dirnames []string
	for _, file := range files {
		if file.IsDir() {
			dirnames = append(dirnames, file.Name())
		}
	}

	return dirnames, nil
}

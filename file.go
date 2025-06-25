package nova

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FileExists 判断文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// IsDir 判断是否为目录
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile 判断是否为文件
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// ReadFile 读取文件内容
func ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// WriteFile 写入文件内容
func WriteFile(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0644)
}

// AppendFile 追加文件内容
func AppendFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

// RemoveFile 删除文件
func RemoveFile(path string) error {
	return os.Remove(path)
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// MoveFile 移动文件
func MoveFile(src, dst string) error {
	return os.Rename(src, dst)
}

// GetFileSize 获取文件大小
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFileModTime 获取文件修改时间
func GetFileModTime(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.ModTime().Unix(), nil
}

// CreateDir 创建目录
func CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// RemoveDir 删除目录
func RemoveDir(path string) error {
	return os.RemoveAll(path)
}

// ListFiles 列出目录下所有文件
func ListFiles(dir string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// ListDirs 列出目录下所有子目录
func ListDirs(dir string) ([]string, error) {
	dirs := []string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != dir {
			dirs = append(dirs, path)
		}
		return nil
	})
	return dirs, err
}

// GetFileExt 获取文件扩展名
func GetFileExt(path string) string {
	return filepath.Ext(path)
}

// GetFileName 获取文件名（不含扩展名）
func GetFileName(path string) string {
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}

// GetFilePath 获取文件路径（不含文件名）
func GetFilePath(path string) string {
	return filepath.Dir(path)
}

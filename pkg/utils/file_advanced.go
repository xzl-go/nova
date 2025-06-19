package utils

import (
	"archive/zip"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// GetFileMD5 获取文件的 MD5 值
func GetFileMD5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// CompressGzip 使用 Gzip 压缩文件
func CompressGzip(src, dst string) error {
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

	gzipWriter := gzip.NewWriter(destination)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, source)
	return err
}

// DecompressGzip 解压 Gzip 文件
func DecompressGzip(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	gzipReader, err := gzip.NewReader(source)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, gzipReader)
	return err
}

// CompressZip 压缩文件或目录为 ZIP
func CompressZip(src, dst string) error {
	zipFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 创建相对路径
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// 跳过根目录
		if relPath == "." {
			return nil
		}

		// 创建 ZIP 文件头
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// DecompressZip 解压 ZIP 文件
func DecompressZip(src, dst string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(dst, file.Name)

		if file.FileInfo().IsDir() {
			err = os.MkdirAll(path, file.Mode())
			if err != nil {
				return err
			}
			continue
		}

		// 创建目标文件
		writer, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		// 打开源文件
		rc, err := file.Open()
		if err != nil {
			writer.Close()
			return err
		}

		// 复制文件内容
		_, err = io.Copy(writer, rc)
		writer.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetFileMimeType 获取文件的 MIME 类型
func GetFileMimeType(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 读取文件头部
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", err
	}

	// 根据文件头部判断 MIME 类型
	switch {
	case strings.HasPrefix(string(buffer), "PK\x03\x04"):
		return "application/zip", nil
	case strings.HasPrefix(string(buffer), "\x1f\x8b\x08"):
		return "application/gzip", nil
	case strings.HasPrefix(string(buffer), "\x89PNG\r\n\x1a\n"):
		return "image/png", nil
	case strings.HasPrefix(string(buffer), "\xff\xd8\xff"):
		return "image/jpeg", nil
	case strings.HasPrefix(string(buffer), "GIF87a") || strings.HasPrefix(string(buffer), "GIF89a"):
		return "image/gif", nil
	case strings.HasPrefix(string(buffer), "%PDF"):
		return "application/pdf", nil
	default:
		return "application/octet-stream", nil
	}
}

// GetFileSizeString 获取文件大小的可读字符串
func GetFileSizeString(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/TB)
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// GetFilePermissions 获取文件权限
func GetFilePermissions(path string) (os.FileMode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Mode(), nil
}

// SetFilePermissions 设置文件权限
func SetFilePermissions(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

// GetFileOwner 获取文件所有者
func GetFileOwner(path string) (int, int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, 0, err
	}
	stat := info.Sys().(*syscall.Stat_t)
	return int(stat.Uid), int(stat.Gid), nil
}

// SetFileOwner 设置文件所有者
func SetFileOwner(path string, uid, gid int) error {
	return os.Chown(path, uid, gid)
}

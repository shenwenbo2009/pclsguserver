package upload

import (
	"fmt"
	"path/filepath"
	"strings"
)

type FileValidator struct {
	pluginAllowedTypes []string
	fileAllowedTypes   []string
}

func NewFileValidator(pluginTypes, fileTypes []string) *FileValidator {
	return &FileValidator{
		pluginAllowedTypes: pluginTypes,
		fileAllowedTypes:   fileTypes,
	}
}

func (v *FileValidator) ValidatePackageType(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".gz" && strings.HasSuffix(strings.ToLower(filename), ".tar.gz") {
		ext = ".tar.gz"
	}
	for _, allowed := range v.pluginAllowedTypes {
		if ext == strings.ToLower(allowed) {
			return nil
		}
	}
	return fmt.Errorf("不支持的文件类型: %s，允许: %s", ext, strings.Join(v.pluginAllowedTypes, ", "))
}

func (v *FileValidator) ValidateFileType(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowed := range v.fileAllowedTypes {
		if ext == strings.ToLower(allowed) {
			return nil
		}
	}
	return fmt.Errorf("不支持的文件类型: %s，允许: %s", ext, strings.Join(v.fileAllowedTypes, ", "))
}

func (v *FileValidator) ValidateSize(size, maxSize int64) error {
	if size > maxSize {
		return fmt.Errorf("文件大小 %d 超过限制 %d", size, maxSize)
	}
	return nil
}

func (v *FileValidator) SanitizeFilename(filename string) (string, error) {
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, "%", "")
	if filename == "" || filename == "." {
		return "", fmt.Errorf("无效文件名")
	}
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return "", fmt.Errorf("文件名包含非法字符")
	}
	return filename, nil
}

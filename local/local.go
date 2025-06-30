/**
 * @Author: guxline zjguoxin@163.com
 * @Date: 2025/7/1 00:33:00
 * @LastEditors: guxline zjguoxin@163.com
 * @LastEditTime: 2025/7/1 00:33:00
 * Description:
 * Copyright: Copyright (©) 2025 中易综服. All rights reserved.
 */
package local

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zjguoxin/gosuploader/config"
)

// LocalUploader 本地文件上传处理器
type LocalUploader struct {
	basePath string // 基础存储路径
}

// New 创建本地文件上传处理器
// 如果需要自定义路径，可以传入config.LocalConfig结构体
// 例如：config.LocalConfig{BasePath: "custom/path/to/uploads"}
// 注意：BasePath必须是一个绝对路径或相对于当前工作目录的路径
// 如果BasePath为空，则使用默认路径"storage/uploads"
// 如果BasePath不存在，会自动创建
// 如果BasePath不是目录(是一个文件)，则会返回错误
// 如果BasePath是一个相对路径，则会相对于当前工作目录创建目录
// 如果BasePath是一个绝对路径，则会在该路径下创建目录
// 如果BasePath是一个不存在的路径，则会自动创建该路径
func New(cfg config.LocalConfig) *LocalUploader {
	// 如果未配置基础路径，使用默认值
	if cfg.BasePath == "" {
		cfg.BasePath = "storage/uploads"
	}

	return &LocalUploader{
		basePath: cfg.BasePath,
	}
}

// UploadFile 上传multipart表单文件
func (u *LocalUploader) UploadFile(file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", errors.New("file header cannot be nil")
	}

	// 打开上传文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// 生成存储路径和文件名
	filePath, err := u.generateFilePath(file.Filename)
	if err != nil {
		return "", fmt.Errorf("failed to generate file path: %w", err)
	}

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// 返回相对路径
	relPath, err := filepath.Rel(u.basePath, filePath)
	if err != nil {
		return filePath, nil // 如果获取相对路径失败，返回绝对路径
	}

	return relPath, nil
}

// UploadBinary 上传二进制数据
// filename: 原始文件名，用于生成存储路径和文件名
// content: 二进制内容，不能为空
// 返回值: 相对路径或绝对路径，上传失败时返回错误
func (u *LocalUploader) UploadBinary(filename string, content []byte) (string, error) {
	if len(content) == 0 {
		return "", errors.New("content cannot be empty")
	}

	// 生成存储路径和文件名
	filePath, err := u.generateFilePath(filename)
	if err != nil {
		return "", fmt.Errorf("failed to generate file path: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// 返回相对路径
	relPath, err := filepath.Rel(u.basePath, filePath)
	if err != nil {
		return filePath, nil
	}

	return relPath, nil
}

// UploadBase64 上传Base64编码的文件
// filename: 原始文件名，用于生成存储路径和文件名
// base64Str: Base64编码的字符串，不能为空
// 返回值: 相对路径或绝对路径，上传失败时返回错误
// 注意：Base64字符串必须是有效的Base64编码，否则会返回解码错误
func (u *LocalUploader) UploadBase64(filename string, base64Str string) (string, error) {
	if base64Str == "" {
		return "", errors.New("base64 content cannot be empty")
	}

	// 解码Base64数据
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	return u.UploadBinary(filename, data)
}

// Delete 删除文件
// filePath: 文件的相对路径或绝对路径
// 返回值: nil表示删除成功，非nil表示删除失败
// 注意：如果文件不存在，会返回错误
// 如果filePath是相对路径，则相对于basePath进行查找
// 如果filePath是绝对路径，则直接使用该路径进行删除
// 如果filePath是一个目录，则会返回错误
func (u *LocalUploader) Delete(filePath string) error {
	fullPath := filepath.Join(u.basePath, filePath)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file not exists: %s", fullPath)
	}

	// 删除文件
	err := os.Remove(fullPath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

// generateFilePath 生成完整的文件存储路径
func (u *LocalUploader) generateFilePath(originalName string) (string, error) {
	// 生成日期目录
	dateDir := time.Now().Format("2006/01/02")
	storageDir := filepath.Join(u.basePath, dateDir)

	// 创建目录
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %w", err)
	}

	// 生成唯一文件名
	ext := filepath.Ext(originalName)
	baseName := strings.TrimSuffix(filepath.Base(originalName), ext)
	uniqueName := fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixNano(), ext)

	return filepath.Join(storageDir, uniqueName), nil
}

// ensureBasePathExists 确保基础路径存在
func (u *LocalUploader) ensureBasePathExists() error {
	return os.MkdirAll(u.basePath, 0755)
}

/**
 * @Author: guxline zjguoxin@163.com
 * @Date: 2025/7/1 00:39:00
 * @LastEditors: guxline zjguoxin@163.com
 * @LastEditTime: 2025/7/1 00:39:00
 * Description:
 * Copyright: Copyright (©) 2025 中易综服. All rights reserved.
 */
package aliyun

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/zjguoxin/gosuploader/config"
)

// AliUploader 阿里云OSS上传处理器
type AliUploader struct {
	client   *oss.Client
	bucket   *oss.Bucket
	config   config.AliyunConfig
	endpoint string
}

// New 创建阿里云OSS上传处理器
func New(cfg config.AliyunConfig) (*AliUploader, error) {
	// 验证必要配置
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" || cfg.BucketName == "" {
		return nil, errors.New("aliyun OSS configuration is incomplete")
	}

	// 创建OSS客户端
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %w", err)
	}

	// 获取存储空间
	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}

	// 构建endpoint
	endpoint := "https://" + cfg.BucketName + "." + cfg.Endpoint
	if cfg.Domain != "" {
		endpoint = "https://" + cfg.Domain
	}

	return &AliUploader{
		client:   client,
		bucket:   bucket,
		config:   cfg,
		endpoint: endpoint,
	}, nil
}

// UploadFile 上传multipart表单文件
func (u *AliUploader) UploadFile(file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", errors.New("file header cannot be nil")
	}

	// 打开上传文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// 生成存储对象键
	objectKey := u.generateObjectKey(file.Filename)

	// 上传文件到OSS
	err = u.bucket.PutObject(objectKey, src)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to OSS: %w", err)
	}

	return u.getFileURL(objectKey), nil
}

// UploadBinary 上传二进制数据
func (u *AliUploader) UploadBinary(filename string, content []byte) (string, error) {
	if len(content) == 0 {
		return "", errors.New("content cannot be empty")
	}

	// 生成存储对象键
	objectKey := u.generateObjectKey(filename)

	// 上传文件到OSS
	err := u.bucket.PutObject(objectKey, bytes.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("failed to upload binary to OSS: %w", err)
	}

	return u.getFileURL(objectKey), nil
}

// UploadBase64 上传Base64编码的文件
func (u *AliUploader) UploadBase64(filename string, base64Str string) (string, error) {
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

// Delete 删除OSS文件
func (u *AliUploader) Delete(objectKey string) error {
	if objectKey == "" {
		return errors.New("object key cannot be empty")
	}

	err := u.bucket.DeleteObject(objectKey)
	if err != nil {
		return fmt.Errorf("failed to delete OSS object: %w", err)
	}

	return nil
}

// generateObjectKey 生成存储对象键
func (u *AliUploader) generateObjectKey(originalName string) string {
	// 获取文件扩展名
	ext := filepath.Ext(originalName)
	baseName := strings.TrimSuffix(filepath.Base(originalName), ext)

	// 生成日期路径和唯一文件名
	datePath := time.Now().Format("2006/01/02")
	uniqueName := fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixNano(), ext)

	return filepath.Join(datePath, uniqueName)
}

// getFileURL 获取文件访问URL
func (u *AliUploader) getFileURL(objectKey string) string {
	return fmt.Sprintf("%s/%s", u.endpoint, objectKey)
}

// SetACL 设置文件访问权限
func (u *AliUploader) SetACL(objectKey string, acl oss.ACLType) error {
	return u.bucket.SetObjectACL(objectKey, acl)
}

// GetSignedURL 获取带签名的临时URL
func (u *AliUploader) GetSignedURL(objectKey string, expiredInSec int64) (string, error) {
	signedURL, err := u.bucket.SignURL(objectKey, oss.HTTPGet, expiredInSec)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}
	return signedURL, nil
}

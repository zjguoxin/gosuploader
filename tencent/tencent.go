/**
 * @Author: guxline zjguoxin@163.com
 * @Date: 2025/7/1 00:48:47
 * @LastEditors: guxline zjguoxin@163.com
 * @LastEditTime: 2025/7/1 00:48:47
 * Description:
 * Copyright: Copyright (©) 2025 中易综服. All rights reserved.
 */
package tencent

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/zjguoxin/gosuploader/config"
)

// TencentUploader 腾讯云COS上传处理器
type TencentUploader struct {
	client *cos.Client
	config config.TencentConfig
}

// New 创建腾讯云COS上传处理器
func New(cfg config.TencentConfig) (*TencentUploader, error) {
	// 验证必要配置
	if cfg.SecretID == "" || cfg.SecretKey == "" || cfg.BucketName == "" || cfg.Region == "" {
		return nil, errors.New("tencent COS configuration is incomplete")
	}

	// 构建存储桶URL
	bucketURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.BucketName, cfg.Region)
	// if cfg.Domain != "" {
	// 	bucketURL = "https://" + cfg.Domain
	// }

	// 解析URL
	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse COS URL: %w", err)
	}

	// 创建COS客户端
	baseURL := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})

	// 验证连接
	_, err = client.Bucket.Head(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to COS bucket: %w", err)
	}

	return &TencentUploader{
		client: client,
		config: cfg,
	}, nil
}

// UploadFile 上传multipart表单文件
func (u *TencentUploader) UploadFile(file *multipart.FileHeader) (string, error) {
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

	// 上传文件到COS
	_, err = u.client.Object.Put(context.Background(), objectKey, src, nil)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to COS: %w", err)
	}

	return u.getFileURL(objectKey), nil
}

// UploadBinary 上传二进制数据
func (u *TencentUploader) UploadBinary(filename string, content []byte) (string, error) {
	if len(content) == 0 {
		return "", errors.New("content cannot be empty")
	}

	// 生成存储对象键
	objectKey := u.generateObjectKey(filename)

	// 上传文件到COS
	_, err := u.client.Object.Put(context.Background(), objectKey, bytes.NewReader(content), nil)
	if err != nil {
		return "", fmt.Errorf("failed to upload binary to COS: %w", err)
	}

	return u.getFileURL(objectKey), nil
}

// UploadBase64 上传Base64编码的文件
func (u *TencentUploader) UploadBase64(filename string, base64Str string) (string, error) {
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

// Delete 删除COS文件
func (u *TencentUploader) Delete(objectKey string) error {
	if objectKey == "" {
		return errors.New("object key cannot be empty")
	}

	_, err := u.client.Object.Delete(context.Background(), objectKey)
	if err != nil {
		return fmt.Errorf("failed to delete COS object: %w", err)
	}

	return nil
}

// generateObjectKey 生成存储对象键
func (u *TencentUploader) generateObjectKey(originalName string) string {
	// 获取文件扩展名
	ext := filepath.Ext(originalName)
	baseName := strings.TrimSuffix(filepath.Base(originalName), ext)

	// 生成日期路径和唯一文件名
	datePath := time.Now().Format("2006/01/02")
	uniqueName := fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixNano(), ext)

	return filepath.Join(datePath, uniqueName)
}

// getFileURL 获取文件访问URL
func (u *TencentUploader) getFileURL(objectKey string) string {
	if u.config.Domain != "" {
		return fmt.Sprintf("https://%s/%s", u.config.Domain, objectKey)
	}
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", u.config.BucketName, u.config.Region, objectKey)
}

// GetPresignedURL 获取预签名URL
func (u *TencentUploader) GetPresignedURL(objectKey string, expired time.Duration) (string, error) {
	presignedURL, err := u.client.Object.GetPresignedURL(
		context.Background(),
		http.MethodGet,
		objectKey,
		u.config.SecretID,
		u.config.SecretKey,
		expired,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

// SetACL 设置文件访问权限
func (u *TencentUploader) SetACL(objectKey string, acl string) error {
	_, err := u.client.Object.PutACL(context.Background(), objectKey, &cos.ObjectPutACLOptions{
		Header: &cos.ACLHeaderOptions{
			XCosACL: acl,
		},
	})
	return err
}

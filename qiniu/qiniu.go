/**
 * @Author: guxline zjguoxin@163.com
 * @Date: 2025/7/1 00:13:40
 * @LastEditors: guxline zjguoxin@163.com
 * @LastEditTime: 2025/7/1 00:13:40
 * Description:
 * Copyright: Copyright (©) 2025 中易综服. All rights reserved.
 */
package qiniu

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/zjguoxin/gosuploader/config"
)

type qiniuUploader struct {
	mac    *qbox.Mac
	cfg    storage.Config
	bucket string
	domain string
}

func New(cfg config.QiniuConfig) (*qiniuUploader, error) {
	if cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Bucket == "" {
		return nil, errors.New("qiniu config is incomplete")
	}

	mac := qbox.NewMac(cfg.AccessKey, cfg.SecretKey)
	Region, _ := storage.GetZone(cfg.AccessKey, cfg.Bucket)

	return &qiniuUploader{
		mac:    mac,
		cfg:    storage.Config{Region: Region, Zone: Region, UseHTTPS: true, UseCdnDomains: false},
		bucket: cfg.Bucket,
		domain: cfg.Domain,
	}, nil
}

// getUpToken 获取上传凭证
func (h *qiniuUploader) getUpToken() string {
	// 上传策略
	putPolicy := storage.PutPolicy{
		Scope: h.bucket,
	}
	// 设置凭证有效期
	putPolicy.Expires = 3600 // 1小时
	return putPolicy.UploadToken(h.mac)
}

// generateUniqueKey 生成唯一的文件key
func (h *qiniuUploader) generateUniqueKey(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().UnixNano()
	randomStr := uuid.New().String()[:8]
	return fmt.Sprintf("%d_%s%s", timestamp, randomStr, ext)
}

// getFileURL 获取文件访问URL
func (h *qiniuUploader) getFileURL(key string) string {
	return fmt.Sprintf("https://%s/%s", h.domain, key)
}

// UploadBase64 上传Base64编码的文件
func (h *qiniuUploader) UploadBase64(fileName string, base64Code string) (string, error) {
	if fileName == "" {
		return "", errors.New("文件名不能为空")
	}
	if base64Code == "" {
		return "", errors.New("base64编码不能为空")
	}

	// 生成唯一文件名
	key := h.generateUniqueKey(fileName)

	// 解码Base64数据
	data, err := base64.StdEncoding.DecodeString(base64Code)
	if err != nil {
		return "", fmt.Errorf("base64解码失败: %v", err)
	}

	// 获取上传凭证
	upToken := h.getUpToken()

	// 创建表单上传对象
	formUploader := storage.NewFormUploader(&h.cfg)
	ret := storage.PutRet{}

	// 上传文件
	err = formUploader.Put(context.Background(), &ret, upToken, key, bytes.NewReader(data), int64(len(data)), nil)
	if err != nil {
		return "", fmt.Errorf("七牛云上传失败: %v", err)
	}

	return h.getFileURL(ret.Key), nil
}

// UploadBinary 上传二进制数据
func (h *qiniuUploader) UploadBinary(fileName string, content []byte) (string, error) {
	if fileName == "" {
		return "", errors.New("文件名不能为空")
	}
	if len(content) == 0 {
		return "", errors.New("文件内容不能为空")
	}

	// 生成唯一文件名
	key := h.generateUniqueKey(fileName)

	// 获取上传凭证
	upToken := h.getUpToken()

	// 创建表单上传对象
	formUploader := storage.NewFormUploader(&h.cfg)
	ret := storage.PutRet{}

	// 上传文件
	err := formUploader.Put(context.Background(), &ret, upToken, key, bytes.NewReader(content), int64(len(content)), nil)
	if err != nil {
		return "", fmt.Errorf("七牛云上传失败: %v", err)
	}

	return h.getFileURL(ret.Key), nil
}

// Delete 删除七牛云文件
func (h *qiniuUploader) Delete(filePath string) error {
	if filePath == "" {
		return errors.New("文件路径不能为空")
	}

	// 创建BucketManager
	bucketManager := storage.NewBucketManager(h.mac, &h.cfg)

	// 删除文件
	err := bucketManager.Delete(h.bucket, filePath)
	if err != nil {
		return fmt.Errorf("删除七牛云文件失败: %v", err)
	}

	return nil
}

// UploadFile 上传multipart文件
func (h *qiniuUploader) UploadFile(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", errors.New("文件头不能为空")
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 读取文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	// 使用二进制上传方法
	return h.UploadBinary(fileHeader.Filename, fileBytes)
}

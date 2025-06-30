/**
 * @Author: guxline zjguoxin@163.com
 * @Date: 2025/7/1 00:16:21
 * @LastEditors: guxline zjguoxin@163.com
 * @LastEditTime: 2025/7/1 00:16:21
 * Description:
 * Copyright: Copyright (©) 2025 中易综服. All rights reserved.
 */
package config

// LocalConfig 本地存储配置
type LocalConfig struct {
	BasePath string // 存储基础路径
}

// QiniuConfig 七牛云配置
type QiniuConfig struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
	Region    string // 存储区域
}

// AliyunConfig 阿里云OSS配置
type AliyunConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
	Domain          string
}

// TencentConfig 腾讯云COS配置
type TencentConfig struct {
	SecretID   string
	SecretKey  string
	BucketName string
	Region     string
	Domain     string
}

type ErrInvalidConfig struct {
	error
}

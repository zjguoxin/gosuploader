# GoSUploader - 统一多存储上传工具

![Go Version](https://img.shields.io/github/go-mod/go-version/zjguoxin/gosuploader)
![License](https://img.shields.io/github/license/zjguoxin/gosuploader)
![Tests](https://img.shields.io/github/actions/workflow/status/zjguoxin/gosuploader/go.yml)

GoSUploader 是一个统一的文件上传接口库，支持多种存储后端，包括本地存储、七牛云、阿里云 OSS 和腾讯云 COS。

## 功能特性

- **统一的上传接口**：一致的 API 支持多种上传方式
- **多存储支持**：
  - 本地文件系统
  - 七牛云存储
  - 阿里云 OSS
  - 腾讯云 COS
- **多种上传方式**：
  - 文件上传（`multipart.FileHeader`）
  - 二进制数据上传
  - Base64 字符串上传
- **线程安全**：所有方法都是并发安全的
- **完善的错误处理**：清晰的错误类型和提示信息

## 安装

```bash
go get github.com/zjguoxin/gosuploader
```

## 快速开始

### 基本使用

```go
package main

import (
	"fmt"
	"log"
	"mime/multipart"

	"github.com/zjguoxin/gosuploader"
	"github.com/zjguoxin/gosuploader/config"
)

func main() {
	// 初始化本地存储上传器
	uploader, err := gosuploader.NewUploader(gosuploader.Local, config.LocalConfig{
		BasePath: "./uploads",
	})
	if err != nil {
		log.Fatal(err)
	}

	// 上传文件示例（实际使用时fileHeader来自HTTP请求）
	fileHeader := &multipart.FileHeader{
		Filename: "example.txt",
		Size:     1024,
	}

	filePath, err := uploader.UploadFile(fileHeader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("文件上传成功: %s\n", filePath)

	// 删除文件
	if err := uploader.Delete(filePath); err != nil {
		log.Printf("删除文件失败: %v", err)
	}
}

```

## 配置说明

### 本地存储配置

```go
localCfg := config.LocalConfig{
	BasePath: "./uploads",
}
```

### 本地存储配置

```go
qiniuCfg := config.QiniuConfig{
	AccessKey: "your_access_key",
	SecretKey: "your_secret_key",
	Bucket:    "your_bucket",
	Domain:    "your_domain",
}
```

### 阿里云 OSS 配置

```go
	aliCfg := config.AliyunConfig{
	AccessKeyID: "your_access_key_id",
	AccessKeySecret: "your_access_key_secret",
	Endpoint: "your_endpoint",
	BucketName: "your_bucket",
}
```

### 腾讯云 COS 配置

```go
	txCfg := config.TencentConfig{
	SecretID: "your_secret_id",
	SecretKey: "your_secret_key",
	Region: "your_region",
	BucketName: "your_bucket",
}
```

## API 文档

### 上传器接口

```go
type Uploader interface {
	// 上传文件（来自HTTP请求的multipart.FileHeader）
	UploadFile(file *multipart.FileHeader) (string, error)

	// 上传二进制数据
	UploadBinary(filename string, content []byte) (string, error)

	// 上传Base64编码的数据
	UploadBase64(filename string, base64Str string) (string, error)

	// 删除文件
	Delete(filepath string) error
}
```

## 使用示例

### 七牛云上传器示例

```go
// 创建七牛云配置
qiniuCfg := config.QiniuConfig{
	AccessKey: os.Getenv("QINIU_ACCESS_KEY"),
	SecretKey: os.Getenv("QINIU_SECRET_KEY"),
	Bucket:    os.Getenv("QINIU_BUCKET"),
	Domain:    os.Getenv("QINIU_DOMAIN"),
}
// 初始化七牛云上传器
uploader, err := gosuploader.NewUploader(gosuploader.Qiniu, qiniuCfg)
if err != nil {
	log.Fatal(err)
}
filepath, err := uploader.UploadBinary("test.txt", []byte("upload test"))
if err != nil {
	log.Fatal(err)
}
```

### 阿里云 oss 上传器示例

```go
// 创建阿里云OSS配置
aliCfg := config.AliyunConfig{
	AccessKey: os.Getenv("ALIYUN_ACCESS_KEY"),
	SecretKey: os.Getenv("ALIYUN_SECRET_KEY"),
	Endpoint:  os.Getenv("ALIYUN_ENDPOINT"),
	Bucket:    os.Getenv("ALIYUN_BUCKET"),
	Domain:    os.Getenv("ALIYUN_DOMAIN"),
}
aliUploader, err := uploader.NewAliyunUploader(aliCfg)
if err != nil {
	log.Fatal(err)
}
filepath, err := aliUploader.UploadBinary("test.txt", []byte("upload test"))
if err != nil {
	log.Fatal(err)
}
```

### 腾讯云 cos 上传器示例

```go
// 创建腾讯云COS配置
txCfg := config.TencentConfig{
	SecretID:  os.Getenv("TENCENT_SECRET_ID"),
	SecretKey: os.Getenv("TENCENT_SECRET_KEY"),
	Region:    os.Getenv("TENCENT_REGION"),
	Bucket:    os.Getenv("TENCENT_BUCKET"),
	Domain:    os.Getenv("TENCENT_DOMAIN"),
}
uploader, err := upload.NewTencentUploader(txCfg)
if err != nil {
	log.Fatal(err)
}
filepath, err := uploader.UploadBinary("test.txt", []byte("upload test"))
if err != nil {
	log.Fatal(err)
}
```

## 测试

```bash
# 运行所有测试
go test -v ./...

# 设置环境变量后测试特定存储
export QINIU_ACCESS_KEY=your_key
export QINIU_SECRET_KEY=your_secret
go test -v -run TestQiniuUploader
```

## 许可证

[MIT](https://github.com/zjguoxin/gosuploader/blob/main/LICENSE)

### 作者

[zjguoxin@163.com](https://github.com/zjguoxin)

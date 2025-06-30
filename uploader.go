/**
 * @Author: guxline zjguoxin@163.com
 * @Date: 2025/7/1 00:12:30
 * @LastEditors: guxline zjguoxin@163.com
 * @LastEditTime: 2025/7/1 00:12:30
 * Description:
 * Copyright: Copyright (©) 2025 中易综服. All rights reserved.
 */
package uploader

import (
	"errors"
	"mime/multipart"

	"github.com/zjguoxin/gosuploader/aliyun"
	"github.com/zjguoxin/gosuploader/config"
	"github.com/zjguoxin/gosuploader/local"
	"github.com/zjguoxin/gosuploader/qiniu"
	"github.com/zjguoxin/gosuploader/tencent"
)

var (
	ErrInvalidConfig   = errors.New("invalid config for uploader")
	ErrUnsupportedType = errors.New("unsupported uploader type")
)

type UploadType string

const (
	Local   UploadType = "local"
	Qiniu   UploadType = "qiniu"
	Aliyun  UploadType = "aliyun"
	Tencent UploadType = "tencent"
)

// Uploader 统一上传接口
type Uploader interface {
	UploadFile(file *multipart.FileHeader) (string, error)
	UploadBinary(filename string, content []byte) (string, error)
	UploadBase64(filename string, base64Str string) (string, error)
	Delete(filepath string) error
}

// NewUploader 创建上传器
// 参数:
//   - t: 指定上传类型， (Local/Qiniu/Aliyun/Tencent)
//   - cfg: 是对应的配置结构体
//
// 返回:
//   - Uploader 实例
//   - error 如果创建失败，返回错误信息
func NewUploader(t UploadType, cfg interface{}) (Uploader, error) {
	switch t {
	case Local:
		localCfg, ok := cfg.(config.LocalConfig)
		if !ok {
			return nil, ErrInvalidConfig
		}
		return local.New(localCfg), nil
	case Qiniu:
		qiniuCfg, ok := cfg.(config.QiniuConfig)
		if !ok {
			return nil, ErrInvalidConfig
		}
		return qiniu.New(qiniuCfg)
	case Aliyun:
		aliCfg, ok := cfg.(config.AliyunConfig)
		if !ok {
			return nil, ErrInvalidConfig
		}
		return aliyun.New(aliCfg)
	case Tencent:
		txCfg, ok := cfg.(config.TencentConfig)
		if !ok {
			return nil, ErrInvalidConfig
		}
		return tencent.New(txCfg)
	default:
		return nil, ErrUnsupportedType
	}
}

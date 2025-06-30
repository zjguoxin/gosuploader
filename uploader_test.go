/**
 * @Author: guxline zjguoxin@163.com
 * @Date: 2025/7/1 00:12:30
 * Description: Uploader接口测试文件
 */
package uploader_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	uploader "github.com/zjguoxin/gosuploader"
	"github.com/zjguoxin/gosuploader/config"
)

// 测试辅助函数：创建一个模拟的multipart.FileHeader
func createTestFile(t *testing.T, filename string) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	assert.NoError(t, err)

	_, err = io.WriteString(part, "test file content")
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	file, header, err := req.FormFile("file")
	assert.NoError(t, err)
	defer file.Close()

	return header
}

// 测试本地存储上传
func TestLocalUploader(t *testing.T) {
	// 准备测试目录
	testDir := "./test_uploads"
	defer os.RemoveAll(testDir)

	// 创建本地存储配置
	localCfg := config.LocalConfig{
		BasePath: testDir,
	}

	// 创建上传器
	up, err := uploader.NewUploader(uploader.Local, localCfg)
	assert.NoError(t, err)

	// 测试上传文件
	t.Run("UploadFile", func(t *testing.T) {
		fileHeader := createTestFile(t, "testfile.txt")
		path, err := up.UploadFile(fileHeader)
		assert.NoError(t, err)
		assert.FileExists(t, filepath.Join(testDir, path))
	})

	// 测试上传二进制数据
	t.Run("UploadBinary", func(t *testing.T) {
		path, err := up.UploadBinary("binary.bin", []byte("binary data"))
		assert.NoError(t, err)
		assert.FileExists(t, filepath.Join(testDir, path))
	})

	/// 测试上传Base64数据
	t.Run("UploadBase64", func(t *testing.T) {
		base64Data := "dGVzdCBkYXRh" // "test data" in base64
		path, err := up.UploadBase64("base64.txt", base64Data)
		assert.NoError(t, err)
		assert.FileExists(t, filepath.Join(testDir, path))
	})

	// 测试删除文件
	t.Run("Delete", func(t *testing.T) {
		path, err := up.UploadBinary("todelete.txt", []byte("to be deleted"))
		assert.NoError(t, err)

		fullPath := filepath.Join(testDir, path)
		assert.FileExists(t, fullPath)

		err = up.Delete(path)
		assert.NoError(t, err, "Delete failed")

		// 检查文件是否真的被删除
		_, err = os.Stat(fullPath)
		assert.True(t, os.IsNotExist(err), "File still exists after deletion")
	})
}

// 测试七牛云存储上传
func TestQiniuUploader(t *testing.T) {
	// 创建七牛云配置
	qiniuCfg := config.QiniuConfig{
		AccessKey: os.Getenv("QINIU_ACCESS_KEY"),
		SecretKey: os.Getenv("QINIU_SECRET_KEY"),
		Bucket:    os.Getenv("QINIU_BUCKET"),
		Domain:    os.Getenv("QINIU_DOMAIN"),
	}

	// 如果缺少配置则跳过测试
	if qiniuCfg.AccessKey == "" || qiniuCfg.SecretKey == "" {
		t.Skip("Skipping Qiniu test due to missing environment variables")
	}

	// 创建上传器
	up, err := uploader.NewUploader(uploader.Qiniu, qiniuCfg)
	assert.NoError(t, err)

	// 测试上传文件
	t.Run("UploadFile", func(t *testing.T) {
		fileHeader := createTestFile(t, "qiniu_test.txt")
		path, err := up.UploadFile(fileHeader)

		assert.NoError(t, err)
		assert.NotEmpty(t, path)

		// 添加短暂延迟，确保处理完成
		time.Sleep(1 * time.Second)

		// 测试删除
		key := extractKey(path) // 提取文件key
		err = up.Delete(key)
		assert.NoError(t, err, "Delete failed")
	})
	//测试上传二进制数据
	t.Run("UploadBinary", func(t *testing.T) {
		path, err := up.UploadBinary("qiniu_test.bin", []byte("qiniu test data"))

		assert.NoError(t, err)
		assert.NotEmpty(t, path)

		// 添加短暂延迟，确保处理完成
		time.Sleep(1 * time.Second)

		// 测试删除
		key := extractKey(path) // 提取文件key
		err = up.Delete(key)
		assert.NoError(t, err, "Delete failed")
	})
	// 测试上传Base64数据
	t.Run("UploadBase64", func(t *testing.T) {
		base64Data := "dGVzdCBkYXRh" // "test data" in base64
		path, err := up.UploadBase64("qiniu_test.txt", base64Data)

		assert.NoError(t, err)
		assert.NotEmpty(t, path)

		// 测试删除
		key := extractKey(path) // 提取文件key
		err = up.Delete(key)
		assert.NoError(t, err, "Delete failed")
	})
}

// 测试阿里云OSS上传
func TestAliyunUploader(t *testing.T) {
	// 创建阿里云配置
	aliCfg := config.AliyunConfig{
		AccessKeyID:     os.Getenv("ALI_ACCESS_KEY"),
		AccessKeySecret: os.Getenv("ALI_SECRET_KEY"),
		Endpoint:        os.Getenv("ALI_ENDPOINT"),
		BucketName:      os.Getenv("ALI_BUCKET"),
		Domain:          os.Getenv("ALI_DOMAIN"),
	}

	// 如果缺少配置则跳过测试
	if aliCfg.AccessKeyID == "" || aliCfg.AccessKeySecret == "" {
		t.Skip("Skipping Aliyun test due to missing environment variables")
	}

	// 创建上传器
	up, err := uploader.NewUploader(uploader.Aliyun, aliCfg)
	assert.NoError(t, err)
	//// 测试上传文件
	t.Run("UploadFile", func(t *testing.T) {
		fileHeader := createTestFile(t, "qiniu_test.txt")
		path, err := up.UploadFile(fileHeader)

		assert.NoError(t, err)
		assert.NotEmpty(t, path)

		// 添加短暂延迟，确保处理完成
		time.Sleep(1 * time.Second)

		// 测试删除
		key := extractKey(path) // 提取文件key
		err = up.Delete(key)
		assert.NoError(t, err, "Delete failed")
	})
	//测试上传二进制数据
	t.Run("UploadBinary", func(t *testing.T) {
		path, err := up.UploadBinary("qiniu_test.bin", []byte("qiniu test data"))

		assert.NoError(t, err)
		assert.NotEmpty(t, path)

		// 添加短暂延迟，确保处理完成
		time.Sleep(1 * time.Second)

		// 测试删除
		key := extractKey(path) // 提取文件key
		err = up.Delete(key)
		assert.NoError(t, err, "Delete failed")
	})
	// 测试上传Base64数据
	t.Run("UploadBase64", func(t *testing.T) {
		base64Data := "dGVzdCBkYXRh" // "test data" in base64
		path, err := up.UploadBase64("qiniu_test.txt", base64Data)

		assert.NoError(t, err)
		assert.NotEmpty(t, path)

		// 测试删除
		key := extractKey(path) // 提取文件key
		err = up.Delete(key)
		assert.NoError(t, err, "Delete failed")
	})
}

// 测试腾讯云COS上传
func TestTencentUploader(t *testing.T) {
	// 创建腾讯云配置
	txCfg := config.TencentConfig{
		SecretID:   os.Getenv("TENCENT_SECRET_ID"),
		SecretKey:  os.Getenv("TENCENT_SECRET_KEY"),
		Region:     os.Getenv("TENCENT_REGION"),
		BucketName: os.Getenv("TENCENT_BUCKET"),
		Domain:     os.Getenv("TENCENT_DOMAIN"),
	}

	// 如果缺少配置则跳过测试
	if txCfg.SecretID == "" || txCfg.SecretKey == "" {
		t.Skip("Skipping Tencent test due to missing environment variables")
	}

	// 创建上传器
	up, err := uploader.NewUploader(uploader.Tencent, txCfg)
	assert.NoError(t, err)

	// 测试上传Base64数据
	t.Run("UploadFile", func(t *testing.T) {
		fileHeader := createTestFile(t, "qiniu_test.txt")
		path, err := up.UploadFile(fileHeader)

		assert.NoError(t, err)
		assert.NotEmpty(t, path)

		// 添加短暂延迟，确保处理完成
		time.Sleep(1 * time.Second)

		// 测试删除
		key := extractKey(path) // 提取文件key
		err = up.Delete(key)
		assert.NoError(t, err, "Delete failed")
	})
	//测试上传二进制数据
	t.Run("UploadBinary", func(t *testing.T) {
		path, err := up.UploadBinary("qiniu_test.bin", []byte("qiniu test data"))

		assert.NoError(t, err)
		assert.NotEmpty(t, path)

		// 添加短暂延迟，确保处理完成
		time.Sleep(1 * time.Second)

		// 测试删除
		key := extractKey(path) // 提取文件key
		err = up.Delete(key)
		assert.NoError(t, err, "Delete failed")
	})
	// 测试上传Base64数据
	t.Run("UploadBase64", func(t *testing.T) {
		base64Data := "dGVzdCBkYXRh" // "test data" in base64
		path, err := up.UploadBase64("qiniu_test.txt", base64Data)

		assert.NoError(t, err)
		assert.NotEmpty(t, path)

		// 测试删除
		key := extractKey(path) // 提取文件key
		err = up.Delete(key)
		assert.NoError(t, err, "Delete failed")
	})
}

// 测试无效配置
func TestInvalidConfig(t *testing.T) {
	// 测试本地存储无效配置
	_, err := uploader.NewUploader(uploader.Local, "invalid config")
	assert.EqualError(t, err, uploader.ErrInvalidConfig.Error())

	// 测试不支持的存储类型
	_, err = uploader.NewUploader("unsupported", nil)
	assert.EqualError(t, err, uploader.ErrUnsupportedType.Error())
}

// extractQiniuKey 从URL中提取七牛云文件key
func extractKey(url string) string {
	// 简单实现：去除http://和https://开头部分
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")

	// 去除域名部分
	parts := strings.SplitN(url, "/", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return url
}

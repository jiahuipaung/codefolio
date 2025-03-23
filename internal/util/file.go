package util

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// 上传目录
	UploadDir = "uploads"
	// 简历子目录
	ResumeDir = "resumes"
	// 允许的最大文件大小 (10MB)
	MaxFileSize = 10 * 1024 * 1024
	// 允许的文件类型
	AllowedFileType = "application/pdf"
)

// 文件相关错误
var (
	ErrFileTooLarge    = errors.New("文件大小超过限制")
	ErrInvalidFileType = errors.New("不支持的文件类型")
	ErrFileNotFound    = errors.New("文件不存在")
	ErrSaveFileFailed  = errors.New("保存文件失败")
)

// UploadFileResult 文件上传结果
type UploadFileResult struct {
	FilePath string
	FileName string
	FileType string
	FileSize int64
}

// SaveUploadedFile 保存上传的文件
func SaveUploadedFile(c *gin.Context, file *multipart.FileHeader, userID uint) (*UploadFileResult, error) {
	// 检查文件大小
	if file.Size > MaxFileSize {
		return nil, ErrFileTooLarge
	}

	// 检查文件类型
	if !strings.Contains(file.Header.Get("Content-Type"), "pdf") {
		return nil, ErrInvalidFileType
	}

	// 打开源文件
	src, err := file.Open()
	if err != nil {
		GetLogger().Error("打开上传文件失败", zap.Error(err))
		return nil, err
	}
	defer src.Close()

	// 创建目录结构 uploads/resumes/user_id/year_month/
	yearMonth := time.Now().Format("2006_01")
	dirPath := filepath.Join(UploadDir, ResumeDir, fmt.Sprintf("%d", userID), yearMonth)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		GetLogger().Error("创建上传目录失败", zap.Error(err), zap.String("path", dirPath))
		return nil, err
	}

	// 生成唯一文件名
	fileExt := filepath.Ext(file.Filename)
	originalName := strings.TrimSuffix(file.Filename, fileExt)
	newFileName := fmt.Sprintf("%s_%s%s", uuid.New().String()[:8], originalName, fileExt)
	filePath := filepath.Join(dirPath, newFileName)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		GetLogger().Error("创建目标文件失败", zap.Error(err), zap.String("path", filePath))
		return nil, err
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, src); err != nil {
		GetLogger().Error("复制文件内容失败", zap.Error(err))
		return nil, err
	}

	return &UploadFileResult{
		FilePath: filePath,
		FileName: file.Filename,
		FileType: file.Header.Get("Content-Type"),
		FileSize: file.Size,
	}, nil
}

// DeleteFile 删除文件
func DeleteFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return ErrFileNotFound
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		GetLogger().Error("删除文件失败", zap.Error(err), zap.String("path", filePath))
		return err
	}

	return nil
}

// GetFileURL 获取文件URL
func GetFileURL(c *gin.Context, filePath string) string {
	// 将文件路径转换为URL
	// 例如: uploads/resumes/1/2023_03/abc.pdf -> /api/v1/files/resumes/1/2023_03/abc.pdf
	relativePath := strings.TrimPrefix(filePath, UploadDir)
	relativePath = strings.TrimPrefix(relativePath, "/")

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/api/v1/files/%s",
		scheme,
		c.Request.Host,
		relativePath)
}

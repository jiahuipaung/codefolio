package util

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// 上传目录常量
var (
	// UploadDir 文件上传根目录
	UploadDir = "uploads"
	// ResumeDir 简历存储子目录
	ResumeDir = "resumes"
	// MaxFileSize 允许的最大文件大小 (10MB)
	MaxFileSize int64 = 10 * 1024 * 1024
	// AllowedFileType 允许的文件类型
	AllowedFileType = "application/pdf"
)

// 图片格式
const (
	// 默认图片格式
	DefaultImageFormat = "jpg"
	// 默认图片质量 (1-100)
	DefaultImageQuality = 90
	// 默认DPI
	DefaultDPI = 150
)

// 设置上传配置
func SetUploadConfig(uploadDir string, maxSize int64, allowedType string) {
	if uploadDir != "" {
		UploadDir = uploadDir
	}
	if maxSize > 0 {
		MaxFileSize = maxSize
	}
	if allowedType != "" {
		AllowedFileType = allowedType
	}
}

// 文件相关错误
var (
	ErrFileTooLarge      = errors.New("文件大小超过限制")
	ErrInvalidFileType   = errors.New("不支持的文件类型")
	ErrFileNotFound      = errors.New("文件不存在")
	ErrSaveFileFailed    = errors.New("保存文件失败")
	ErrConvertPDFFailed  = errors.New("PDF转换失败")
	ErrInvalidPageNumber = errors.New("无效的页码")
	ErrCommandNotFound   = errors.New("命令未找到")
)

// UploadFileResult 文件上传结果
type UploadFileResult struct {
	FilePath string
	FileName string
	FileType string
	FileSize int64
}

// ConvertPDFToImage 将PDF文件转换为图片
func ConvertPDFToImage(pdfPath string) (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return "", ErrFileNotFound
	}

	// 提取目录和文件名
	dir := filepath.Dir(pdfPath)
	base := filepath.Base(pdfPath)
	filename := strings.TrimSuffix(base, filepath.Ext(base))

	// 构建输出图片路径
	outputImageBase := filepath.Join(dir, filename)
	outputImagePath := fmt.Sprintf("%s.%s", outputImageBase, DefaultImageFormat)

	// 首先尝试使用pdftoppm（通常在Linux和macOS上可用）
	if _, err := exec.LookPath("pdftoppm"); err == nil {
		// pdftoppm命令，将PDF转换为JPG，只转换第一页
		cmd := exec.Command(
			"pdftoppm",
			"-jpeg",                             // 输出JPEG格式
			"-r", fmt.Sprintf("%d", DefaultDPI), // 设置DPI
			"-jpegopt", fmt.Sprintf("quality=%d", DefaultImageQuality), // 设置JPEG质量
			"-f", "1", "-l", "1", // 只处理第一页 (from 1, to 1)
			"-singlefile",   // 输出单个文件
			pdfPath,         // 输入PDF文件
			outputImageBase, // 输出基础名称
		)

		// 执行命令
		if err := cmd.Run(); err != nil {
			GetLogger().Error("PDF转图片失败 (pdftoppm)", zap.Error(err), zap.String("pdf", pdfPath))
			return "", ErrConvertPDFFailed
		}

		return outputImagePath, nil
	}

	// 如果pdftoppm不可用，尝试使用ghostscript
	if _, err := exec.LookPath("gs"); err == nil {
		// ghostscript命令，将PDF转换为JPG，只转换第一页
		cmd := exec.Command(
			"gs",
			"-sDEVICE=jpeg", // 设备类型为JPEG
			fmt.Sprintf("-dJPEGQ=%d", DefaultImageQuality), // JPEG质量
			fmt.Sprintf("-r%d", DefaultDPI),                // 设置DPI
			"-dBATCH",                                      // 批处理模式
			"-dNOPAUSE",                                    // 不暂停
			"-dFirstPage=1",                                // 从第一页开始
			"-dLastPage=1",                                 // 到第一页结束
			fmt.Sprintf("-sOutputFile=%s", outputImagePath), // 输出文件
			pdfPath, // 输入PDF文件
		)

		// 执行命令
		if err := cmd.Run(); err != nil {
			GetLogger().Error("PDF转图片失败 (ghostscript)", zap.Error(err), zap.String("pdf", pdfPath))
			return "", ErrConvertPDFFailed
		}

		return outputImagePath, nil
	}

	// 都不可用，尝试使用convert命令（ImageMagick）
	if _, err := exec.LookPath("convert"); err == nil {
		// ImageMagick命令，将PDF转换为JPG，只转换第一页
		cmd := exec.Command(
			"convert",
			"-density", fmt.Sprintf("%d", DefaultDPI), // 设置DPI
			"-quality", fmt.Sprintf("%d", DefaultImageQuality), // 设置质量
			fmt.Sprintf("%s[0]", pdfPath), // 输入PDF文件，只处理第一页
			outputImagePath,               // 输出文件
		)

		// 执行命令
		if err := cmd.Run(); err != nil {
			GetLogger().Error("PDF转图片失败 (ImageMagick)", zap.Error(err), zap.String("pdf", pdfPath))
			return "", ErrConvertPDFFailed
		}

		return outputImagePath, nil
	}

	// 所有工具都不可用
	GetLogger().Error("未找到PDF转图片工具", zap.String("pdf", pdfPath))
	return "", ErrCommandNotFound
}

// SaveUploadedPDF 保存上传的PDF并转换为图片
func SaveUploadedPDF(c *gin.Context, file *multipart.FileHeader, userID uint) (*UploadFileResult, error) {
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
	fileID := uuid.New().String()[:8]
	originalName := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))

	// 临时PDF文件路径
	tempPDFName := fmt.Sprintf("%s_%s.pdf", fileID, originalName)
	tempPDFPath := filepath.Join(dirPath, tempPDFName)

	// 创建临时PDF文件
	pdfFile, err := os.Create(tempPDFPath)
	if err != nil {
		GetLogger().Error("创建临时PDF文件失败", zap.Error(err), zap.String("path", tempPDFPath))
		return nil, err
	}
	defer pdfFile.Close()

	// 复制文件内容到临时PDF文件
	if _, err = io.Copy(pdfFile, src); err != nil {
		GetLogger().Error("复制PDF内容失败", zap.Error(err))
		_ = os.Remove(tempPDFPath) // 清理临时文件
		return nil, err
	}

	// 确保文件内容已写入磁盘
	if err = pdfFile.Sync(); err != nil {
		GetLogger().Error("同步PDF文件内容失败", zap.Error(err))
		_ = os.Remove(tempPDFPath) // 清理临时文件
		return nil, err
	}
	pdfFile.Close() // 关闭文件以便后续操作

	// 将PDF转换为图片
	imagePath, err := ConvertPDFToImage(tempPDFPath)
	if err != nil {
		_ = os.Remove(tempPDFPath) // 清理临时文件
		return nil, err
	}

	// 转换成功后删除临时PDF文件
	_ = os.Remove(tempPDFPath)

	// 记录文件路径
	GetLogger().Info("图片生成成功",
		zap.String("原PDF", tempPDFPath),
		zap.String("转换图片", imagePath))

	// 获取相对于项目根目录的相对路径
	relPath := "/" + imagePath

	// 返回结果
	return &UploadFileResult{
		FilePath: relPath, // 保存相对路径，方便构建URL
		FileName: file.Filename,
		FileType: "image/" + DefaultImageFormat,
		FileSize: file.Size,
	}, nil
}

// SaveUploadedFile 保存上传的文件（保留兼容性）
func SaveUploadedFile(c *gin.Context, file *multipart.FileHeader, userID uint) (*UploadFileResult, error) {
	// 检查文件大小
	if file.Size > MaxFileSize {
		return nil, ErrFileTooLarge
	}

	// 检查文件类型
	if !strings.Contains(file.Header.Get("Content-Type"), "pdf") {
		return nil, ErrInvalidFileType
	}

	// 如果是PDF文件，使用新方法进行转换
	if strings.Contains(file.Header.Get("Content-Type"), "pdf") {
		return SaveUploadedPDF(c, file, userID)
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
	// 例如: uploads/resumes/1/2023_03/abc.jpg -> /api/v1/files/resumes/1/2023_03/abc.jpg
	relativePath := strings.TrimPrefix(filePath, UploadDir)
	relativePath = strings.TrimPrefix(relativePath, "/")

	// 如果运行在Docker中，Host可能需要设置为外部访问地址
	host := c.Request.Host

	// 如果没有主机头，使用默认值
	if host == "" {
		host = "localhost:8080"
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	fileURL := fmt.Sprintf("%s://%s/api/v1/files/%s", scheme, host, relativePath)

	// 记录URL转换过程（方便调试）
	GetLogger().Debug("文件URL生成",
		zap.String("filePath", filePath),
		zap.String("relativePath", relativePath),
		zap.String("fileURL", fileURL))

	return fileURL
}

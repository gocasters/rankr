package middleware

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Config struct {
	FileSize int64    `koanf:"file_size"`
	FileType []string `koanf:"file_type"`
}

type Middleware struct {
	Config Config
}

func New(cfg Config) Middleware {
	return Middleware{Config: cfg}
}

func (m Middleware) CheckFile(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		fileHeader, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "fail to get file",
				"error":   err.Error(),
			})
		}

		srcFile, err := fileHeader.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": fmt.Sprintf("failed to open file: %v", err),
			})
		}
		defer srcFile.Close()

		buffer := make([]byte, 512)
		n, _ := srcFile.Read(buffer)
		contentType := http.DetectContentType(buffer[:n])

		if _, err := srcFile.Seek(0, 0); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": fmt.Sprintf("failed to reset file pointer: %v", err),
			})
		}

		if fileHeader.Size > m.Config.FileSize {
			return c.JSON(http.StatusRequestEntityTooLarge, map[string]interface{}{
				"message": fmt.Sprintf("file size exceeds %dMB limit", m.Config.FileSize),
			})
		}

		if !m.validFileType(contentType) {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message":  "invalid file type",
				"allowed":  m.Config.FileType,
				"received": contentType,
			})
		}

		c.Set("FileType", contentType)

		return next(c)
	}
}

func (m Middleware) validFileType(ext string) bool {
	for _, v := range m.Config.FileType {
		if v == ext {
			return true
		}
	}

	return false
}

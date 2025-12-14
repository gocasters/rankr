package middleware

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"path/filepath"
	"strings"
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

		file, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "file is required; send it via 'file' field",
				"error":   err.Error(),
			})
		}

		if file.Size > m.Config.FileSize {
			return c.JSON(http.StatusRequestEntityTooLarge, map[string]interface{}{
				"message": fmt.Sprintf("file size exceeds %dMB limit", m.Config.FileSize),
			})
		}

		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(file.Filename)), ".")
		if !m.validFileType(ext) {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message":  "invalid file type",
				"allowed":  m.Config.FileType,
				"received": ext,
			})
		}

		c.Set("FileType", ext)

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

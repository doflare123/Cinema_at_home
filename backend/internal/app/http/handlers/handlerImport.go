package handlers

import (
	"cinema/internal/app/services"
	appErrors "cinema/internal/errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const maxImportPayloadSize = 15 << 20

type ImportHandler interface {
	ImportCatalogExcel(c *gin.Context)
}

type importHandler struct {
	service services.ImportService
}

func NewImportHandler(service services.ImportService) ImportHandler {
	return &importHandler{service: service}
}

func (h *importHandler) ImportCatalogExcel(c *gin.Context) {
	adminID, ok := currentUserID(c)
	if !ok {
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxImportPayloadSize+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
		return
	}
	if len(data) > maxImportPayloadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is too large"})
		return
	}

	dryRun := parseBool(c.Query("dry_run")) || parseBool(c.PostForm("dry_run"))
	result, err := h.service.ImportCatalogExcel(adminID, fileHeader.Filename, data, dryRun)
	if err != nil {
		switch err {
		case appErrors.ErrImportEmptyFile,
			appErrors.ErrImportInvalidWorkbook,
			appErrors.ErrImportNoSheets,
			appErrors.ErrImportNoUsableRows:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"import": result})
}

func parseBool(raw string) bool {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if normalized == "" {
		return false
	}
	value, err := strconv.ParseBool(normalized)
	return err == nil && value
}

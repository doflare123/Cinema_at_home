package handlers

import (
	"bytes"
	"cinema/internal/models/dto"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubImportService struct {
	importCatalogExcel func(adminID uint, fileName string, payload []byte, dryRun bool) (dto.ImportCatalogResult, error)
}

func (s stubImportService) ImportCatalogExcel(adminID uint, fileName string, payload []byte, dryRun bool) (dto.ImportCatalogResult, error) {
	if s.importCatalogExcel != nil {
		return s.importCatalogExcel(adminID, fileName, payload, dryRun)
	}
	return dto.ImportCatalogResult{}, nil
}

func TestImportHandlerRequiresMultipartFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewImportHandler(stubImportService{})
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/admin/import/excel/catalog", strings.NewReader(""))
	ctx.Request.Header.Set("Content-Type", "multipart/form-data; boundary=x")
	ctx.Set("user_id", uint(2))

	handler.ImportCatalogExcel(ctx)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestImportHandlerPassesFileAndFlagsToService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	handler := NewImportHandler(stubImportService{
		importCatalogExcel: func(adminID uint, fileName string, payload []byte, dryRun bool) (dto.ImportCatalogResult, error) {
			called = true
			if adminID != 2 {
				t.Fatalf("expected adminID=2, got %d", adminID)
			}
			if fileName != "catalog.xlsm" {
				t.Fatalf("unexpected filename: %s", fileName)
			}
			if !dryRun {
				t.Fatalf("expected dryRun=true")
			}
			if string(payload) != "test-bytes" {
				t.Fatalf("unexpected payload: %q", string(payload))
			}
			runID := uint(7)
			return dto.ImportCatalogResult{RunID: &runID, RowsParsed: 1}, nil
		},
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "catalog.xlsm")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := part.Write([]byte("test-bytes")); err != nil {
		t.Fatalf("write form file failed: %v", err)
	}
	if err := writer.WriteField("dry_run", "true"); err != nil {
		t.Fatalf("write dry_run failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer failed: %v", err)
	}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/admin/import/excel/catalog", body)
	ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())
	ctx.Set("user_id", uint(2))

	handler.ImportCatalogExcel(ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected service to be called")
	}
}

func TestImportHandlerMapsServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewImportHandler(stubImportService{
		importCatalogExcel: func(adminID uint, fileName string, payload []byte, dryRun bool) (dto.ImportCatalogResult, error) {
			return dto.ImportCatalogResult{}, errors.New("boom")
		},
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "catalog.xlsm")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := part.Write([]byte("test-bytes")); err != nil {
		t.Fatalf("write form file failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer failed: %v", err)
	}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/admin/import/excel/catalog", body)
	ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())
	ctx.Set("user_id", uint(2))

	handler.ImportCatalogExcel(ctx)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

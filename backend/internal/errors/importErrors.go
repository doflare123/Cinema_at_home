package errors

import "errors"

var (
	ErrImportEmptyFile          = errors.New("import file is empty")
	ErrImportInvalidWorkbook    = errors.New("invalid excel workbook")
	ErrImportNoSheets           = errors.New("excel workbook has no worksheets")
	ErrImportNoUsableRows       = errors.New("excel workbook has no usable rows")
	ErrImportInvalidUser        = errors.New("invalid import user")
	ErrImportFailedPersistAudit = errors.New("failed to persist import audit")
)

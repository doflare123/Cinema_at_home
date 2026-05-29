package services

import (
	"archive/zip"
	"bytes"
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"cinema/internal/repository"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

const (
	fallbackImportedDescription = "Imported from legacy Excel catalog"
	fallbackImportedCountry     = "Unknown"
	fallbackImportedPoster      = "imported://poster"
	fallbackImportedDuration    = int32(90)
	fallbackImportedReleaseYear = 1970
)

var (
	reLeadingOrdinal = regexp.MustCompile(`^\s*\d+\s*[\.\,\)\-\:]\s*`)
	reOnlyDigits     = regexp.MustCompile(`^\d+$`)
)

type ImportService interface {
	ImportCatalogExcel(adminID uint, fileName string, payload []byte, dryRun bool) (dto.ImportCatalogResult, error)
}

type importService struct {
	rep repository.Repository
}

func NewImportService(rep repository.Repository) ImportService {
	return &importService{rep: rep}
}

func (s *importService) ImportCatalogExcel(adminID uint, fileName string, payload []byte, dryRun bool) (dto.ImportCatalogResult, error) {
	if adminID == 0 {
		return dto.ImportCatalogResult{}, appErrors.ErrImportInvalidUser
	}
	if len(payload) == 0 {
		return dto.ImportCatalogResult{}, appErrors.ErrImportEmptyFile
	}

	titles, rowsTotal, rowsInvalid, warnings, err := extractCatalogTitlesFromExcel(payload)
	if err != nil {
		return dto.ImportCatalogResult{}, err
	}
	if len(titles) == 0 {
		return dto.ImportCatalogResult{}, appErrors.ErrImportNoUsableRows
	}

	created := 0
	skippedExisting := 0
	parsed := len(titles)
	var runID uint

	err = s.rep.Transaction(func(tx repository.Repository) error {
		for _, title := range titles {
			var existing models.Film
			findErr := tx.Where("LOWER(title) = LOWER(?)", title).First(&existing).Error
			if findErr == nil {
				skippedExisting++
				continue
			}
			if !errors.Is(findErr, gorm.ErrRecordNotFound) {
				return findErr
			}

			if dryRun {
				continue
			}

			film := models.Film{
				Title:            title,
				Description:      fallbackImportedDescription,
				SmallDescription: fallbackImportedDescription,
				Duration:         fallbackImportedDuration,
				ReleaseDate:      fallbackImportedReleaseYear,
				Country:          fallbackImportedCountry,
				Poster:           fallbackImportedPoster,
				RatingKp:         0,
			}
			if createErr := tx.Create(&film).Error; createErr != nil {
				if strings.Contains(strings.ToLower(createErr.Error()), "unique") {
					skippedExisting++
					continue
				}
				return createErr
			}
			created++
		}

		summary := fmt.Sprintf("warnings=%d", len(warnings))
		if len(warnings) > 0 {
			summary = strings.Join(warnings, "\n")
		}

		run := models.ImportRun{
			SourceFileName:      strings.TrimSpace(fileName),
			ImportedByUserID:    adminID,
			DryRun:              dryRun,
			RowsTotal:           rowsTotal,
			RowsParsed:          parsed,
			RowsCreated:         created,
			RowsSkippedExisting: skippedExisting,
			RowsSkippedInvalid:  rowsInvalid,
			Notes:               summary,
		}
		if run.SourceFileName == "" {
			run.SourceFileName = "uploaded.xlsx"
		}
		if createErr := tx.Create(&run).Error; createErr != nil {
			return appErrors.ErrImportFailedPersistAudit
		}
		runID = run.ID
		return nil
	})
	if err != nil {
		return dto.ImportCatalogResult{}, err
	}

	sort.Strings(titles)
	sample := titles
	if len(sample) > 15 {
		sample = sample[:15]
	}

	return dto.ImportCatalogResult{
		RunID:               &runID,
		FileName:            strings.TrimSpace(fileName),
		DryRun:              dryRun,
		RowsTotal:           rowsTotal,
		RowsParsed:          parsed,
		RowsCreated:         created,
		RowsSkippedExisting: skippedExisting,
		RowsSkippedInvalid:  rowsInvalid,
		SampleTitles:        sample,
		Warnings:            warnings,
	}, nil
}

type workbookXML struct {
	Sheets []sheetXML `xml:"sheets>sheet"`
}

type sheetXML struct {
	Name  string `xml:"name,attr"`
	RelID string `xml:"id,attr"`
}

type relationshipsXML struct {
	Relationships []relationshipXML `xml:"Relationship"`
}

type relationshipXML struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
}

type sharedStringsXML struct {
	Items []sharedStringItemXML `xml:"si"`
}

type sharedStringItemXML struct {
	Text string               `xml:"t"`
	Runs []sharedStringRunXML `xml:"r"`
}

type sharedStringRunXML struct {
	Text string `xml:"t"`
}

type worksheetXML struct {
	Rows []worksheetRowXML `xml:"sheetData>row"`
}

type worksheetRowXML struct {
	Cells []worksheetCellXML `xml:"c"`
}

type worksheetCellXML struct {
	Ref       string                `xml:"r,attr"`
	Type      string                `xml:"t,attr"`
	Value     string                `xml:"v"`
	InlineStr worksheetInlineStrXML `xml:"is"`
}

type worksheetInlineStrXML struct {
	Text string `xml:"t"`
}

func extractCatalogTitlesFromExcel(payload []byte) ([]string, int, int, []string, error) {
	reader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return nil, 0, 0, nil, appErrors.ErrImportInvalidWorkbook
	}

	workbookRaw, err := readZipFile(reader, "xl/workbook.xml")
	if err != nil {
		return nil, 0, 0, nil, appErrors.ErrImportInvalidWorkbook
	}
	relsRaw, err := readZipFile(reader, "xl/_rels/workbook.xml.rels")
	if err != nil {
		return nil, 0, 0, nil, appErrors.ErrImportInvalidWorkbook
	}

	var workbook workbookXML
	if err := xml.Unmarshal(workbookRaw, &workbook); err != nil {
		return nil, 0, 0, nil, appErrors.ErrImportInvalidWorkbook
	}
	if len(workbook.Sheets) == 0 {
		return nil, 0, 0, nil, appErrors.ErrImportNoSheets
	}

	var rels relationshipsXML
	if err := xml.Unmarshal(relsRaw, &rels); err != nil {
		return nil, 0, 0, nil, appErrors.ErrImportInvalidWorkbook
	}
	relMap := make(map[string]string, len(rels.Relationships))
	for _, rel := range rels.Relationships {
		relMap[rel.ID] = strings.TrimPrefix(rel.Target, "/")
	}

	shared := make([]string, 0)
	if sharedRaw, err := readZipFile(reader, "xl/sharedStrings.xml"); err == nil {
		var sst sharedStringsXML
		if xml.Unmarshal(sharedRaw, &sst) == nil {
			shared = make([]string, 0, len(sst.Items))
			for _, item := range sst.Items {
				if item.Text != "" {
					shared = append(shared, item.Text)
					continue
				}
				var parts strings.Builder
				for _, run := range item.Runs {
					parts.WriteString(run.Text)
				}
				shared = append(shared, parts.String())
			}
		}
	}

	titleSet := make(map[string]struct{})
	titles := make([]string, 0, 256)
	warnings := make([]string, 0, 8)
	rowsTotal := 0
	rowsInvalid := 0

	for _, sheet := range prioritizeSheets(workbook.Sheets) {
		titlesBeforeSheet := len(titles)
		target, ok := relMap[sheet.RelID]
		if !ok || strings.TrimSpace(target) == "" {
			continue
		}
		if !strings.HasPrefix(target, "xl/") {
			target = "xl/" + target
		}

		worksheetRaw, err := readZipFile(reader, target)
		if err != nil {
			warnings = append(warnings, "failed to read worksheet "+sheet.Name)
			continue
		}

		var ws worksheetXML
		if err := xml.Unmarshal(worksheetRaw, &ws); err != nil {
			warnings = append(warnings, "failed to parse worksheet "+sheet.Name)
			continue
		}

		for _, row := range ws.Rows {
			raw, ok := firstColumnCellValue(row.Cells, shared)
			if !ok {
				continue
			}
			rowsTotal++
			normalized, valid := normalizeImportedTitle(raw)
			if !valid {
				rowsInvalid++
				continue
			}
			key := strings.ToLower(normalized)
			if _, exists := titleSet[key]; exists {
				continue
			}
			titleSet[key] = struct{}{}
			titles = append(titles, normalized)
		}

		if len(titles) > titlesBeforeSheet {
			break
		}
	}

	return titles, rowsTotal, rowsInvalid, warnings, nil
}

func prioritizeSheets(sheets []sheetXML) []sheetXML {
	if len(sheets) <= 1 {
		return sheets
	}
	priority := func(name string) int {
		n := strings.ToLower(strings.TrimSpace(name))
		switch n {
		case "\u043b\u0438\u0441\u04422", "sheet3":
			return 0
		case "\u043b\u0438\u0441\u04421", "sheet2":
			return 1
		default:
			return 10
		}
	}

	out := make([]sheetXML, len(sheets))
	copy(out, sheets)
	sort.SliceStable(out, func(i, j int) bool {
		return priority(out[i].Name) < priority(out[j].Name)
	})
	return out
}

func readZipFile(reader *zip.Reader, name string) ([]byte, error) {
	for _, f := range reader.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}
	return nil, fmt.Errorf("zip entry not found: %s", name)
}

func firstColumnCellValue(cells []worksheetCellXML, shared []string) (string, bool) {
	for _, cell := range cells {
		ref := strings.ToUpper(strings.TrimSpace(cell.Ref))
		if ref != "" && !strings.HasPrefix(ref, "A") {
			continue
		}
		value := decodeWorksheetCellValue(cell, shared)
		if strings.TrimSpace(value) == "" {
			return "", false
		}
		return value, true
	}
	return "", false
}

func decodeWorksheetCellValue(cell worksheetCellXML, shared []string) string {
	if strings.TrimSpace(cell.InlineStr.Text) != "" {
		return cell.InlineStr.Text
	}

	value := strings.TrimSpace(cell.Value)
	if value == "" {
		return ""
	}
	if strings.EqualFold(cell.Type, "s") {
		idx, err := strconv.Atoi(value)
		if err != nil || idx < 0 || idx >= len(shared) {
			return ""
		}
		return shared[idx]
	}
	return value
}

func normalizeImportedTitle(raw string) (string, bool) {
	title := strings.TrimSpace(raw)
	if title == "" {
		return "", false
	}
	title = reLeadingOrdinal.ReplaceAllString(title, "")
	title = strings.TrimSpace(title)
	if title == "" {
		return "", false
	}

	low := strings.ToLower(title)
	if low == "1" || low == "-" || low == "skip" || low == "\u0441\u043a\u0438\u043f" {
		return "", false
	}
	if reOnlyDigits.MatchString(title) {
		year, err := strconv.Atoi(title)
		if err == nil && year >= 1890 && year <= 2100 {
			return title, true
		}
		return "", false
	}
	return title, true
}

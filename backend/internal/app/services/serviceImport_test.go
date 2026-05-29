package services

import (
	"archive/zip"
	"bytes"
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"encoding/xml"
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestImportServiceDryRunParsesButDoesNotCreateFilms(t *testing.T) {
	rep := newImportTestRepository(t)
	svc := NewImportService(rep)

	workbook := buildWorkbookForImportTest(t, []string{
		"1. The Matrix",
		"2. Fight Club",
		"1",
		"skip",
		"2. Fight Club",
	})

	result, err := svc.ImportCatalogExcel(2, "catalog.xlsm", workbook, true)
	if err != nil {
		t.Fatalf("ImportCatalogExcel failed: %v", err)
	}
	if !result.DryRun {
		t.Fatalf("expected dry-run result")
	}
	if result.RowsParsed != 2 {
		t.Fatalf("expected RowsParsed=2, got %d", result.RowsParsed)
	}
	if result.RowsCreated != 0 {
		t.Fatalf("expected RowsCreated=0 in dry-run, got %d", result.RowsCreated)
	}

	var films []models.Film
	if err := rep.Find(&films).Error; err != nil {
		t.Fatalf("find films failed: %v", err)
	}
	if len(films) != 0 {
		t.Fatalf("expected no created films, got %d", len(films))
	}
}

func TestImportServiceCreatesFilmsAndSkipsExisting(t *testing.T) {
	rep := newImportTestRepository(t)
	if err := rep.Create(&models.Film{
		Title:            "Fight Club",
		Description:      "existing",
		SmallDescription: "existing",
		Duration:         100,
		ReleaseDate:      1999,
		Country:          "US",
		Poster:           "x",
		RatingKp:         8.8,
	}).Error; err != nil {
		t.Fatalf("seed film failed: %v", err)
	}

	svc := NewImportService(rep)
	workbook := buildWorkbookForImportTest(t, []string{
		"1. The Matrix",
		"2. Fight Club",
		"3. Interstellar",
	})

	result, err := svc.ImportCatalogExcel(2, "catalog.xlsm", workbook, false)
	if err != nil {
		t.Fatalf("ImportCatalogExcel failed: %v", err)
	}
	if result.RowsCreated != 2 {
		t.Fatalf("expected RowsCreated=2, got %d", result.RowsCreated)
	}
	if result.RowsSkippedExisting != 1 {
		t.Fatalf("expected RowsSkippedExisting=1, got %d", result.RowsSkippedExisting)
	}
	if result.RunID == nil || *result.RunID == 0 {
		t.Fatalf("expected persisted run id")
	}

	var runs []models.ImportRun
	if err := rep.Find(&runs).Error; err != nil {
		t.Fatalf("find runs failed: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected one import run, got %d", len(runs))
	}
}

func TestImportServiceRejectsInvalidWorkbook(t *testing.T) {
	rep := newImportTestRepository(t)
	svc := NewImportService(rep)

	_, err := svc.ImportCatalogExcel(2, "catalog.xlsm", []byte("broken"), true)
	if !errors.Is(err, appErrors.ErrImportInvalidWorkbook) {
		t.Fatalf("expected ErrImportInvalidWorkbook, got %v", err)
	}
}

func TestNormalizeImportedTitleKeepsNumericMovieName(t *testing.T) {
	title, ok := normalizeImportedTitle("1917")
	if !ok {
		t.Fatal("expected title to stay valid")
	}
	if title != "1917" {
		t.Fatalf("expected 1917, got %q", title)
	}
}

func TestExtractCatalogTitlesUsesFirstUsablePrioritySheet(t *testing.T) {
	workbook := buildWorkbookForImportTestTwoSheets(t,
		[]string{"1. First Sheet Movie"},
		[]string{"1. Second Sheet Movie"},
	)

	titles, _, _, _, err := extractCatalogTitlesFromExcel(workbook)
	if err != nil {
		t.Fatalf("extractCatalogTitlesFromExcel failed: %v", err)
	}
	if len(titles) != 1 {
		t.Fatalf("expected 1 title, got %d", len(titles))
	}
	if titles[0] != "First Sheet Movie" {
		t.Fatalf("unexpected title: %v", titles)
	}
}

func newImportTestRepository(t *testing.T) *testRepository {
	t.Helper()

	rep := &testRepository{db: openTestSQLiteDB(t)}
	for _, model := range []interface{}{
		&models.Role{},
		&models.User{},
		&models.Film{},
		&models.ImportRun{},
	} {
		if err := rep.AutoMigrate(model); err != nil {
			t.Fatalf("migrate failed: %v", err)
		}
	}
	if err := rep.Create(&models.Role{ID: 1, Name: "member"}).Error; err != nil {
		t.Fatalf("seed role member failed: %v", err)
	}
	if err := rep.Create(&models.Role{ID: 2, Name: "admin"}).Error; err != nil {
		t.Fatalf("seed role admin failed: %v", err)
	}
	if err := rep.Create(&models.User{ID: 2, Username: "admin", Password: "x", DisplayName: "Admin", RoleID: 2, Status: "active"}).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	return rep
}

func buildWorkbookForImportTest(t *testing.T, titles []string) []byte {
	t.Helper()

	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)

	addZipFile(t, zw, "[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet3.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/sharedStrings.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"/>
</Types>`)

	addZipFile(t, zw, "_rels/.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`)

	addZipFile(t, zw, "xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Лист2" sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`)

	addZipFile(t, zw, "xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet3.xml"/>
</Relationships>`)

	sharedBuilder := strings.Builder{}
	sharedBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sharedBuilder.WriteString(`<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
	for _, title := range titles {
		var escaped bytes.Buffer
		if escErr := xml.EscapeText(&escaped, []byte(title)); escErr != nil {
			t.Fatalf("xml escape shared string failed: %v", escErr)
		}
		sharedBuilder.WriteString("<si>")
		sharedBuilder.WriteString("<t>")
		sharedBuilder.WriteString(escaped.String())
		sharedBuilder.WriteString("</t>")
		sharedBuilder.WriteString("</si>")
	}
	sharedBuilder.WriteString("</sst>")
	addZipFile(t, zw, "xl/sharedStrings.xml", sharedBuilder.String())

	sheetBuilder := strings.Builder{}
	sheetBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sheetBuilder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	for i := range titles {
		row := i + 1
		sheetBuilder.WriteString(`<row r="`)
		sheetBuilder.WriteString(strconv.Itoa(row))
		sheetBuilder.WriteString(`"><c r="A`)
		sheetBuilder.WriteString(strconv.Itoa(row))
		sheetBuilder.WriteString(`" t="s"><v>`)
		sheetBuilder.WriteString(strconv.Itoa(i))
		sheetBuilder.WriteString(`</v></c></row>`)
	}
	sheetBuilder.WriteString(`</sheetData></worksheet>`)
	addZipFile(t, zw, "xl/worksheets/sheet3.xml", sheetBuilder.String())

	if err := zw.Close(); err != nil {
		t.Fatalf("close zip failed: %v", err)
	}
	return buf.Bytes()
}

func buildWorkbookForImportTestTwoSheets(t *testing.T, firstSheetTitles, secondSheetTitles []string) []byte {
	t.Helper()

	allTitles := make([]string, 0, len(firstSheetTitles)+len(secondSheetTitles))
	allTitles = append(allTitles, firstSheetTitles...)
	allTitles = append(allTitles, secondSheetTitles...)

	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)

	addZipFile(t, zw, "[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet2.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/worksheets/sheet3.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/sharedStrings.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"/>
</Types>`)

	addZipFile(t, zw, "_rels/.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`)

	addZipFile(t, zw, "xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Лист2" sheetId="1" r:id="rId1"/>
    <sheet name="Лист1" sheetId="2" r:id="rId2"/>
  </sheets>
</workbook>`)

	addZipFile(t, zw, "xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet3.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/>
</Relationships>`)

	addZipFile(t, zw, "xl/sharedStrings.xml", buildSharedStringsXML(t, allTitles))
	addZipFile(t, zw, "xl/worksheets/sheet3.xml", buildSheetXMLFromSharedIndex(0, len(firstSheetTitles)))
	addZipFile(t, zw, "xl/worksheets/sheet2.xml", buildSheetXMLFromSharedIndex(len(firstSheetTitles), len(secondSheetTitles)))

	if err := zw.Close(); err != nil {
		t.Fatalf("close zip failed: %v", err)
	}
	return buf.Bytes()
}

func buildSharedStringsXML(t *testing.T, titles []string) string {
	t.Helper()

	sharedBuilder := strings.Builder{}
	sharedBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sharedBuilder.WriteString(`<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
	for _, title := range titles {
		var escaped bytes.Buffer
		if escErr := xml.EscapeText(&escaped, []byte(title)); escErr != nil {
			t.Fatalf("xml escape shared string failed: %v", escErr)
		}
		sharedBuilder.WriteString("<si><t>")
		sharedBuilder.WriteString(escaped.String())
		sharedBuilder.WriteString("</t></si>")
	}
	sharedBuilder.WriteString("</sst>")
	return sharedBuilder.String()
}

func buildSheetXMLFromSharedIndex(startIdx, count int) string {
	sheetBuilder := strings.Builder{}
	sheetBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sheetBuilder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	for i := 0; i < count; i++ {
		row := i + 1
		sheetBuilder.WriteString(`<row r="`)
		sheetBuilder.WriteString(strconv.Itoa(row))
		sheetBuilder.WriteString(`"><c r="A`)
		sheetBuilder.WriteString(strconv.Itoa(row))
		sheetBuilder.WriteString(`" t="s"><v>`)
		sheetBuilder.WriteString(strconv.Itoa(startIdx + i))
		sheetBuilder.WriteString(`</v></c></row>`)
	}
	sheetBuilder.WriteString(`</sheetData></worksheet>`)
	return sheetBuilder.String()
}

func addZipFile(t *testing.T, zw *zip.Writer, name, content string) {
	t.Helper()
	w, err := zw.Create(name)
	if err != nil {
		t.Fatalf("zip create %s failed: %v", name, err)
	}
	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatalf("zip write %s failed: %v", name, err)
	}
}

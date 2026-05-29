package dto

type ImportCatalogResult struct {
	RunID               *uint    `json:"run_id,omitempty"`
	FileName            string   `json:"file_name"`
	DryRun              bool     `json:"dry_run"`
	RowsTotal           int      `json:"rows_total"`
	RowsParsed          int      `json:"rows_parsed"`
	RowsCreated         int      `json:"rows_created"`
	RowsSkippedExisting int      `json:"rows_skipped_existing"`
	RowsSkippedInvalid  int      `json:"rows_skipped_invalid"`
	SampleTitles        []string `json:"sample_titles"`
	Warnings            []string `json:"warnings"`
}

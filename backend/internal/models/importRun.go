package models

import "time"

type ImportRun struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	SourceFileName      string    `gorm:"not null" json:"source_file_name"`
	ImportedByUserID    uint      `gorm:"not null;index" json:"imported_by_user_id"`
	DryRun              bool      `gorm:"not null;default:false" json:"dry_run"`
	RowsTotal           int       `gorm:"not null;default:0" json:"rows_total"`
	RowsParsed          int       `gorm:"not null;default:0" json:"rows_parsed"`
	RowsCreated         int       `gorm:"not null;default:0" json:"rows_created"`
	RowsSkippedExisting int       `gorm:"not null;default:0" json:"rows_skipped_existing"`
	RowsSkippedInvalid  int       `gorm:"not null;default:0" json:"rows_skipped_invalid"`
	Notes               string    `gorm:"type:text" json:"notes"`
	CreatedAt           time.Time `json:"created_at"`

	ImportedBy User `gorm:"foreignKey:ImportedByUserID" json:"imported_by,omitempty"`
}

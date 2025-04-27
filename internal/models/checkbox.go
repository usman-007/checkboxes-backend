package models

import "time"

// User represents a user in the system
type Checkbox struct {
	RowID     uint32 `json:"row_id"`
	ColumnID uint32 `json:"column_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

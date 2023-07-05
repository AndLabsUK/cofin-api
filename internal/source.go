package internal

import (
	"database/sql/driver"
)

// SourceKind describes a category of source data for API response -- can be an
// SEC filing, an investor call transcript, etc. SourceType implements the
// Scanner interface and the Stringer interface.
type SourceKind string

const (
	Q10 SourceKind = "10-Q"
	K10 SourceKind = "10-K"
)

type Quarter uint8

const (
	Q1 Quarter = iota + 1
	Q2
	Q3
	Q4
)

func (st *SourceKind) Scan(value interface{}) error {
	*st = SourceKind(value.(string))
	return nil
}

func (st SourceKind) Value() (driver.Value, error) {
	return string(st), nil
}

func (st SourceKind) String() string {
	return string(st)
}

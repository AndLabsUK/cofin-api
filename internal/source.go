package internal

// SourceType describes a category of source data for API response -- can be an
// SEC filing, an investor call transcript, etc.
type SourceType string

const (
	Q10 SourceType = "10-Q"
	K10 SourceType = "10-K"
)

type Quarter uint8

const (
	Q1 Quarter = iota + 1
	Q2 Quarter = iota
	Q3 Quarter = iota
	Q4 Quarter = iota
)

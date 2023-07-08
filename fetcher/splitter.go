package fetcher

import "fmt"

// Splitter splits text into chunks.
type Splitter struct {
	chunkLength  int
	chunkOverlap int
}

// NewSplitter returns a new Splitter.
func NewSplitter(chunkLength, chunkOverlap int) (*Splitter, error) {
	if chunkLength < chunkOverlap {
		return &Splitter{}, fmt.Errorf("chunkLength must be greater than chunkOverlap")
	}

	if chunkLength == 0 {
		return &Splitter{}, fmt.Errorf("chunkLength must be greater than 0")
	}

	return &Splitter{
		chunkLength:  chunkLength,
		chunkOverlap: chunkOverlap,
	}, nil
}

// SplitText splits text into chunks.
func (s *Splitter) SplitText(t string) ([]string, error) {
	chunks := make([]string, 0)

	for i := 0; i < len(t); i += s.chunkLength - s.chunkOverlap {
		end := i + s.chunkLength
		if end > len(t) {
			end = len(t)
		}
		chunks = append(chunks, t[i:end])
	}

	return chunks, nil
}

package iosupport

type Line struct {
	Id     int
	Offset int
	Limit  int
}

type Indexer struct {
	sc        *Scanner
	NbOfLines int
	Lines     []Line
}

func NewIndexer(sc *Scanner) *Indexer {
	sc.Reset()
	sc.KeepNewlineSequence(true)
	return &Indexer{
		sc: sc,
	}
}

func (i *Indexer) Analyze() error {
	offset := 0
	for i.sc.ScanLine() {
		if i.sc.Err() != nil {
			return i.sc.Err()
		}
		limit := len(i.sc.Bytes())
		i.Lines = append(i.Lines, Line{i.NbOfLines, offset, limit})
		i.NbOfLines++
		offset = limit + 1
	}
	return nil
}

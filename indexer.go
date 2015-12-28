package iosupport

// Line had variables off indexed line
type Line struct {
	ID     int
	Offset int64
	Limit  int
}

// Indexer conatins all stuff for indexinf lines from a file
type Indexer struct {
	sc        *Scanner
	NbOfLines int
	Lines     []Line
}

// NewIndexer inatanciates a new Indexer
func NewIndexer(sc *Scanner) *Indexer {
	sc.Reset()
	sc.KeepNewlineSequence(true)
	return &Indexer{
		sc: sc,
	}
}

// Analyze creates line's index
func (i *Indexer) Analyze(fns ...func([]byte, int)) error {
	var offset int64
	for i.sc.ScanLine() {
		if i.sc.Err() != nil {
			return i.sc.Err()
		}
		limit := len(i.sc.Bytes())
		i.Lines = append(i.Lines, Line{i.NbOfLines, offset, limit})
		if i.hasReadFunction(fns) {
			fns[0](i.sc.Bytes(), i.NbOfLines)
		}
		i.NbOfLines++
		offset += int64(limit)
	}
	return nil
}

func (i *Indexer) hasReadFunction(fns []func([]byte, int)) bool {
	switch len(fns) {
	case 0:
		return false
	case 1:
		return true
	default:
		panic("Too many arguments (0 or 1 accepted)")
	}
}

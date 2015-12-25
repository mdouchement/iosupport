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

func (i *Indexer) Analyze(fns ...func([]byte, int)) error {
	offset := 0
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
		offset = limit + 1
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

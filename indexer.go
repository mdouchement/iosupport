package iosupport

// Line had variables off indexed line
type Line struct {
	Index  int
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
// Called without arguments, it appends values in Indexer#Lines
// Called with func as argument, it passes values to the func with this following signature fn(line []byte, index int, offset int64, limit int)
func (i *Indexer) Analyze(fns ...func([]byte, int, int64, int) error) error {
	var offset int64
	var limit int
	for i.sc.ScanLine() {
		if i.sc.Err() != nil {
			return i.sc.Err()
		}
		limit = len(i.sc.Bytes())
		if i.hasReadFunction(fns) {
			if err := fns[0](i.sc.Bytes(), i.NbOfLines, offset, limit); err != nil {
				return err
			}
		} else {
			i.Lines = append(i.Lines, Line{i.NbOfLines, offset, limit})
		}
		i.NbOfLines++
		offset += int64(limit)
	}
	return nil
}

func (i *Indexer) hasReadFunction(fns []func([]byte, int, int64, int) error) bool {
	switch len(fns) {
	case 0:
		return false
	case 1:
		return true
	default:
		panic("Too many arguments (0 or 1 accepted)")
	}
}

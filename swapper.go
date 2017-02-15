package iosupport

import (
	"errors"
	"fmt"
	"sync"
)

const (
	arrayAllocRatio float64 = 1.25      // cf. append() mechanism.
	reservedMemory  uint64  = 200 << 20 // Reserved memory (~200MB) for internal stuff.
)

type (
	// A Swapper dumps to filesystem the current index when memory limit is reached.
	Swapper struct {
		limit uint64
		// Storage is the handler that stores all the dumps.
		Storage StorageService
		// Chunksize is the number of elements per chunk within a dump.
		ChunkSize func(nbOfElements int) int
		dumps     []*dump
		tsvLines  TsvLines // Used when no memory dump is needed
	}

	dump struct {
		chunks []string
	}
)

// NewNullSwapper inatanciates a new Swapper without memory limit.
func NewNullSwapper() *Swapper {
	return &Swapper{
		Storage: NewMemStorageService(),
	}
}

// NewSwapper inatanciates a new Swapper with a memory limit in bytes.
func NewSwapper(limit uint64, basepath string) *Swapper {
	chunksize := func() func(nbOfElements int) int {
		// K is the chunksize ratio for a given limit and number of indexed line
		// for a wanted chunksize.
		//   K =        limit     *   noe   / wanted chunksize
		K := float64((1024 << 20) * 4140032 / 500000)
		l := float64(limit)
		return func(nbOfElements int) int {
			noe := float64(nbOfElements)
			return int(noe*l/K) + 1 // Avoid 0 in case of division
		}
	}
	return &Swapper{
		limit:     limit - reservedMemory,
		Storage:   NewHDDStorageService(basepath),
		dumps:     make([]*dump, 0, 4),
		ChunkSize: chunksize(),
	}
}

// IsTimeToSwap returns true when the memory limit is reached.
func (s *Swapper) IsTimeToSwap(elements TsvLines) bool {
	// NullSwapper
	if s.limit == 0 {
		return false
	}

	if len(elements) == cap(elements) {
		mem := GetMemoryUsage().SysKb * 1000
		// current allocation || next allocation
		return mem >= s.limit || uint64(float64(mem)*arrayAllocRatio) >= s.limit
	}

	return false
}

// HasSwapped returns true if there is at least one dump.
func (s *Swapper) HasSwapped() bool {
	return len(s.dumps) > 0
}

// NbOfDumps returns the number of processed dumps.
func (s *Swapper) NbOfDumps() int {
	return len(s.dumps)
}

// Swap dumps the given TsvLines to the configured StorageService.
func (s *Swapper) Swap(elements TsvLines) error {
	// NullSwapper
	if s.limit == 0 {
		return nil
	}

	chunks := s.chunkList(len(elements))
	chunksName := make([]string, 0, len(chunks))
	chunkSize := s.ChunkSize(len(elements))
	for i, size := range chunks {
		offset := i * chunkSize

		name := s.chunkName(len(s.dumps), i)
		err := s.writeChunk(elements[offset:offset+size], name)
		if err != nil {
			return err
		}
		chunksName = append(chunksName, name)
	}
	s.dumps = append(s.dumps, &dump{chunksName})

	return nil
}

// KeepWithoutSwap is used to track TsvLines from the caller and provide an in memory read iterator.
// Use this function instead of Swap when you do not want to dump memory.
func (s *Swapper) KeepWithoutSwap(elements TsvLines) {
	if s.HasSwapped() {
		panic(errors.New("KeepWithoutSwap should not be called when Swap has been used"))
	}
	s.tsvLines = elements
}

// ReadIterator returns an iterator on stored dumps.
func (s *Swapper) ReadIterator() LineIterator {
	if s.limit == 0 || !s.HasSwapped() {
		// NullSwapper or Swapper without any dumps.
		return newTsvLinesIterator(s.tsvLines)
	}
	return newDumpIterator(s.dumps, s.Storage)
}

// EraseAll removes all stored data.
func (s *Swapper) EraseAll() error {
	return s.Storage.EraseAll()
}

// ------------------ //
// Swap stuff         //
// ------------------ //

func (s *Swapper) writeChunk(chunk TsvLines, name string) error {
	return s.Storage.Marshal(name, chunk)
}

// chunkList returns a slice of all chunk size for the given size.
func (s *Swapper) chunkList(size int) []int {
	chunkSize := s.ChunkSize(size)
	nbChunks := size / chunkSize
	if nbChunks == 0 {
		return []int{size}
	}

	sl := make([]int, nbChunks, nbChunks+1)
	for i := 0; i < nbChunks; i++ {
		sl[i] = chunkSize
	}

	if lastChunk := size % chunkSize; lastChunk != 0 {
		sl = append(sl, lastChunk)
	}

	return sl
}

func (s *Swapper) chunkName(dumpKey, chunkKey int) string {
	return fmt.Sprintf("%d-%d.chunk", dumpKey, chunkKey)
}

// ------------------ //
// ReadIterator stuff //
// ------------------ //

// +----------------------+                                 A dump contains one or several sorted chunks.
// | Dump              #1 |                                 A chunk contains severals sorted TsvLines.
// |                      |
// | +----------------+   |                                 lineIterator#Next() iterates over lines.
// | | Chunk       #1 |   |  lineIterator
// | |                +---------------------+               chunkIterators#Next() iterates over chunks.
// | |  +---------+   |   |                 |
// | |  | TsvLine |   |   |                 |               dumpIterator#Next() iterates over dumps
// | |  +---------+   |   |                 |               by finding the best TsvLine
// | |                |   |                 |               according to the CompareFunc used by sort.
// | |  +---------+   |   |                 |
// | |  | TsvLine |   |   |                 |
// | |  +---------+   |   |                 |  chunkIterator
// | |                |   |                 +-----------------------+
// | |  +---------+   |   |                 |                       |
// | |  | TsvLine |   |   |                 |                       |
// | |  +---------+   |   |                 |                       |
// | |                |   |                 |                       |
// | +----------------+   |                 |                       |
// |                      |                 |                       |
// | +----------------+   |                 |                       |
// | | Chunk       #2 |   |  lineIterator   |                       |
// | |                +---------------------+                       |
// | |      ...       |   |                                         |
// | +----------------+   |                                         |
// |                      |                                         |
// +----------------------+                                         |  dumpIterator
//                                                                  +---------------->  Current TsvLine
// +----------------------+                                         |
// | Dump              #2 |                                         |
// |                      |                                         |
// | +----------------+   |                                         |
// | | Chunk       #1 |   |  lineIterator                           |
// | |                +---------------------+                       |
// | |  +---------+   |   |                 |                       |
// | |  | TsvLine |   |   |                 |                       |
// | |  +---------+   |   |                 |                       |
// | |                |   |                 |                       |
// | |  +---------+   |   |                 |                       |
// | |  | TsvLine |   |   |                 |                       |
// | |  +---------+   |   |                 |  chunkIterator        |
// | |                |   |                 +-----------------------+
// | |  +---------+   |   |                 |
// | |  | TsvLine |   |   |                 |
// | |  +---------+   |   |                 |
// | |                |   |                 |
// | +----------------+   |                 |
// |                      |                 |
// | +----------------+   |                 |
// | | Chunk       #2 |   |  lineIterator   |
// | |                +---------------------+
// | |      ...       |   |
// | +----------------+   |
// |                      |
// +----------------------+

type (
	// A LineIterator allows to iterate across TSV lines structure.
	LineIterator interface {
		// Next returns true if an next element is found.
		Next() bool
		// Value returns the current TsvLine.
		Value() TsvLine
		// Error allows to check if an error has occurred.
		Error() error
	}

	// used when there is no generated dump
	tsvLinesIterator struct {
		data    TsvLines
		current int
	}

	// used when there are some generated dumps
	dumpIterator struct {
		current int
		cit     []*chunkIterator
	}
)

func newTsvLinesIterator(lines TsvLines) *tsvLinesIterator {
	return &tsvLinesIterator{
		current: -1,
		data:    lines,
	}
}

// Next returns true if a next element is found.
func (it *tsvLinesIterator) Next() bool {
	it.current++
	return it.current < len(it.data)
}

// Value returns the current TsvLine.
func (it *tsvLinesIterator) Value() TsvLine {
	return it.data[it.current]
}

// Error allows to check if an error has occurred.
func (it *tsvLinesIterator) Error() error {
	return nil
}

// ---

func newDumpIterator(data []*dump, Storage StorageService) *dumpIterator {
	cit := make([]*chunkIterator, len(data))
	for i, d := range data {
		cit[i] = newChunkIterator(d.chunks, Storage)
	}
	return &dumpIterator{
		current: -42,
		cit:     cit,
	}
}

// Next returns true if an next element is found.
func (it *dumpIterator) Next() bool {
	defer it.selectCurrent()

	if it.current == -42 {
		// Iterators initialization
		for _, d := range it.cit {
			d.Next()
		}

		return true
	}

	if it.cit[it.current].Next() {
		return true
	}

	// No longer data for this dump
	copy(it.cit[it.current:], it.cit[it.current+1:])
	it.cit[len(it.cit)-1] = nil
	it.cit = it.cit[:len(it.cit)-1]

	return len(it.cit) > 0
}

// Value returns the current TsvLine.
func (it *dumpIterator) Value() TsvLine {
	return it.cit[it.current].Value()
}

// Error allows to check if an error has occurred.
func (it *dumpIterator) Error() error {
	if it.current == -42 {
		// In case of initialization error
		for _, chunk := range it.cit {
			if chunk.Error() != nil {
				return chunk.Error()
			}
		}
		return nil
	}
	return it.cit[it.current].Error()
}

func (it *dumpIterator) selectCurrent() {
	current := 0

	for i, dump := range it.cit {
		if i == current {
			continue
		}

		if CompareFunc(dump.Value(), it.cit[current].Value()) {
			current = i
		}
	}

	it.current = current
}

// ---- //

type chunkIterator struct {
	current int
	data    []string
	Storage StorageService
	lit     *lineIterator
}

func newChunkIterator(data []string, storage StorageService) *chunkIterator {
	return &chunkIterator{
		data:    data,
		Storage: storage,
		lit:     newLineIterator(storage, data[0]), // Take the first chunk element.
	}
}

// Next returns true if an next element is found.
func (it *chunkIterator) Next() bool {
	if it.lit.Next() {
		// The current chunk has remaining lines to be read
		return true
	}

	it.current++
	if it.current < len(it.data) {
		// Load the next chunk
		it.lit = newLineIterator(it.Storage, it.data[it.current])
		return it.lit.Next()
	}

	return false
}

// Value returns the current TsvLine.
func (it *chunkIterator) Value() TsvLine {
	return it.lit.Value()
}

// Error allows to check if an error has occurred.
func (it *chunkIterator) Error() error {
	return it.lit.Error()
}

// ---- //

var tsvLinePool = sync.Pool{
	New: func() interface{} {
		return TsvLines{}
	},
}

type lineIterator struct {
	current int
	err     error
	data    TsvLines
}

func newLineIterator(storage StorageService, key string) *lineIterator {
	data := tsvLinePool.Get().(TsvLines)
	err := storage.Unmarshal(key, &data)

	return &lineIterator{
		current: -1,
		err:     err,
		data:    data,
	}
}

// Next returns true if an next element is found.
func (it *lineIterator) Next() bool {
	it.current++

	if it.current < len(it.data) {
		return true
	}

	tsvLinePool.Put(it.data)
	return false
}

// Value returns the current TsvLine.
func (it *lineIterator) Value() TsvLine {
	return it.data[it.current]
}

// Error allows to check if an error has occurred.
func (it *lineIterator) Error() error {
	return it.err
}

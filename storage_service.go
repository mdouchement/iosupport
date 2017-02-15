package iosupport

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"
	"sync"

	"github.com/peterbourgon/diskv"
)

// A StorageService allows to store blobs
type StorageService interface {
	// Marshal writes the blob encoding of v to the given key.
	Marshal(key string, v interface{}) error
	// Unmarshal parses the blob-encoded data and stores the result in the value pointed to by v.
	Unmarshal(key string, v interface{}) error
	// EraseAll removes all stored data.
	EraseAll() error
}

// A HDDStorageService allows to reads and writes blobs on filesystem.
type HDDStorageService struct {
	disk *diskv.Diskv
}

// NewHDDStorageService inatanciates a new StorageService.
func NewHDDStorageService(basepath string) *HDDStorageService {
	return &HDDStorageService{
		disk: diskv.New(diskv.Options{
			BasePath:     basepath,
			Compression:  diskv.NewGzipCompression(),
			Transform:    func(s string) []string { return strings.Split(s, "-") },
			CacheSizeMax: 0,
		}),
	}
}

// Marshal writes the blob encoding of v to the given key.
func (s *HDDStorageService) Marshal(key string, v interface{}) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("StorageService: Marshal: %s", err.Error())
	}

	if err := s.disk.Write(key, buf.Bytes()); err != nil {
		return fmt.Errorf("StorageService: Marshal: %s", err.Error())
	}

	return nil
}

// Unmarshal parses the blob-encoded data and stores the result in the value pointed to by v.
func (s *HDDStorageService) Unmarshal(key string, v interface{}) error {
	serialized, err := s.disk.Read(key)
	if err != nil {
		return fmt.Errorf("StorageService: Unmarshal: %s", err.Error())
	}

	r := bytes.NewReader(serialized)
	dec := gob.NewDecoder(r)

	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("StorageService: Unmarshal: %s", err.Error())
	}

	return nil
}

// EraseAll removes all stored data.
func (s *HDDStorageService) EraseAll() error {
	err := s.disk.EraseAll()
	if err != nil {
		return fmt.Errorf("StorageService: EraseAll: %s", err.Error())
	}

	return nil
}

// ------------------ //
// MemStorageService  //
// ------------------ //

// A StorageService allows to reads and writes blobs on filesystem.
type MemStorageService struct {
	sync.RWMutex
	compression diskv.Compression
	registry    map[string][]byte
}

// NewStorageService inatanciates a new StorageService.
func NewMemStorageService() *MemStorageService {
	return &MemStorageService{
		compression: diskv.NewGzipCompression(),
		registry:    make(map[string][]byte, 0),
	}
}

// Marshal writes the blob encoding of v to the given key.
func (s *MemStorageService) Marshal(key string, v interface{}) error {
	var buf bytes.Buffer

	// Attach compression
	bufc, err := s.compression.Writer(&buf)
	if err != nil {
		return fmt.Errorf("MemStorageService: Marshal: %s", err.Error())
	}

	enc := gob.NewEncoder(bufc)
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("MemStorageService: Marshal: %s", err.Error())
	}

	bufc.Close()

	s.Lock()
	defer s.Unlock()
	s.registry[key] = buf.Bytes()

	return nil
}

// Unmarshal parses the blob-encoded data and stores the result in the value pointed to by v.
func (s *MemStorageService) Unmarshal(key string, v interface{}) error {
	s.RLock()
	defer s.RUnlock()

	bufc := bytes.NewBuffer(s.registry[key])
	buf, err := s.compression.Reader(bufc)
	if err != nil {
		return fmt.Errorf("MemStorageService: Unmarshal: %s", err.Error())
	}
	defer buf.Close()

	dec := gob.NewDecoder(buf)
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("MemStorageService: Unmarshal: %s", err.Error())
	}

	return nil
}

// EraseAll removes all stored data.
func (s *MemStorageService) EraseAll() error {
	s.registry = nil
	s.registry = make(map[string][]byte, 0)

	return nil
}

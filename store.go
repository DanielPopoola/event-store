package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

type indexEntry struct {
	offset int64
	length int64
}

type FileEngine struct {
	file  *os.File
	index map[string]indexEntry
	mu    sync.RWMutex
}

func NewFileEngine(path string) (*FileEngine, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	engine := &FileEngine{
		file:  file,
		index: make(map[string]indexEntry),
	}

	if err := engine.replay(); err != nil {
		return nil, err
	}

	return engine, nil
}

func (e *FileEngine) replay() error {
	if _, err := e.file.Seek(0, 0); err != nil {
		return err
	}

	var offset int64
	reader := bufio.NewReader(e.file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		line = strings.TrimSuffix(line, "\n")
		length := int64(len(line)) + 1 //  +1 for newline

		var event struct {
			ID string `json:"id"`
		}

		if err := json.Unmarshal([]byte(line), &event); err != nil || event.ID == "" {
			offset += length
			continue
		}

		e.index[event.ID] = indexEntry{offset: offset, length: length}
		offset += length
	}

	log.Printf("recovered %d events\n", len(e.index))
	return nil
}

func (e *FileEngine) Set(id, line string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	offset, err := e.file.Seek(0, 2)
	if err != nil {
		return err
	}

	length, err := e.file.WriteString(line + "\n")
	if err != nil {
		return err
	}

	if err := e.file.Sync(); err != nil {
		return err
	}

	e.index[id] = indexEntry{
		offset: offset,
		length: int64(length),
	}
	return nil
}

func (e *FileEngine) Get(id string) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	entry, ok := e.index[id]
	if !ok {
		return "", fmt.Errorf("No event found")
	}

	buf := make([]byte, entry.length)
	if _, err := e.file.ReadAt(buf, entry.offset); err != nil {
		return "", err
	}

	return string(buf), nil
}

func (e *FileEngine) Stats() (int, int, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	info, err := e.file.Stat()
	if err != nil {
		return 0, 0, err
	}
	return len(e.index), int(info.Size()), nil
}

func (e *FileEngine) Close() error {
	return e.file.Close()
}

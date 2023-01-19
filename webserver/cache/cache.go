package cache

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var ErrNoSuchFile = errors.New("no such file")

type FileEntry struct {
	Key  string
	Name string
}

// FileCache is a fairly simple file storage handler, that stores files and times them out
type FileCache struct {
	sync.RWMutex
	storageFolder string
	fileTimeout   time.Duration
	files         map[string]fileCacheEntry
	workerContext context.Context
	workerCancel  context.CancelFunc
}

func NewFileCache() *FileCache {
	tempFolder, err := os.MkdirTemp("", "PostBox_*")
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	f := &FileCache{
		storageFolder: tempFolder,
		fileTimeout:   time.Hour,
		files:         make(map[string]fileCacheEntry),
		workerContext: ctx,
		workerCancel:  cancel,
	}
	go f.cleanupWorker()
	return f
}

func (cache *FileCache) Close() {
	cache.workerCancel()
}

func (cache *FileCache) ListFiles() []FileEntry {
	cache.RLock()
	defer cache.RUnlock()
	output := make([]FileEntry, 0, len(cache.files))
	for _, value := range cache.files {
		output = append(output, FileEntry{
			Key:  value.key(),
			Name: value.fileName,
		})
	}
	return output
}

func (cache *FileCache) GetFile(key string) (io.ReadSeekCloser, string, error) {
	cache.RLock()
	defer cache.RUnlock()
	entry, ok := cache.files[key]
	if !ok {
		return nil, "", ErrNoSuchFile
	}
	reader, err := entry.getReader()
	return reader, entry.fileName, err
}
func (cache *FileCache) IngestFile(fileName string, r io.Reader) error {
	cache.Lock()
	defer cache.Unlock()
	fileObject, err := os.CreateTemp(cache.storageFolder, "*")
	if err != nil {
		log.Error().Err(err).Msg("Failed to create temp cache file ")
		return err
	}
	defer fileObject.Close()
	record := fileCacheEntry{
		valid:     true,
		fileName:  fileName,
		filePath:  fileObject.Name(),
		eraseTime: time.Now().Add(cache.fileTimeout),
	}
	_, err = io.Copy(fileObject, r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write temp file ")
		os.Remove(fileObject.Name())
		return err
	}
	//Otherwise we are all good, save and continue on
	cache.files[record.key()] = record
	return nil
}

func (cache *FileCache) cleanupWorker() {
	//Spin forever, cleanup any expired files

	for {
		select {
		case <-time.After(time.Minute * 5):
			log.Info().Msg("Running cleanup scan")
			cache.Lock()
			for s, entry := range cache.files {
				if entry.checkCleanup() {
					log.Info().Str("name", entry.fileName).Msg("Cleaning up temp file after expiry")
					delete(cache.files, s)
				}
			}
			cache.Unlock()
		case <-cache.workerContext.Done():
			log.Warn().Msg("Cache cleanup task exiting")
			return
		}
	}
}

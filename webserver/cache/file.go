package cache

import (
	"crypto/sha1"
	"encoding/base64"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"time"
)

type fileCacheEntry struct {
	valid     bool
	fileName  string
	filePath  string
	eraseTime time.Time
}

func (ce *fileCacheEntry) getReader() (io.ReadSeekCloser, error) {
	return os.Open(ce.filePath)
}

func (ce *fileCacheEntry) key() string {
	hash := sha1.New()
	hash.Write([]byte(ce.filePath))
	sha := base64.URLEncoding.EncodeToString(hash.Sum(nil))
	return sha
}

// checkCleanup checks if this file has expired, and if it has returns true
func (ce *fileCacheEntry) checkCleanup() bool {
	if !ce.valid {
		return true
	}
	if ce.eraseTime.After(time.Now()) {
		if err := os.Remove(ce.fileName); err != nil {
			log.Error().Err(err).Str("path", ce.filePath).Msg("File cleanup failed")
		}
		ce.valid = false
	}
	return !ce.valid
}
func (ce *fileCacheEntry) forceClean() {
	if !ce.valid {
		return
	}
	if err := os.Remove(ce.fileName); err != nil {
		log.Error().Err(err).Str("path", ce.filePath).Msg("File cleanup failed")
	}
	ce.valid = false
}

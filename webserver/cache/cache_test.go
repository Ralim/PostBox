package cache_test

import (
	"github.com/ralim/PostBox/webserver/cache"
	"io"
	"strings"
	"testing"
)

func TestFileCache_IngestFile(t *testing.T) {
	cacheObj := cache.NewFileCache()
	defer cacheObj.Close()
	r := strings.NewReader("TestData")
	err := cacheObj.IngestFile("test.file", r)
	if err != nil {
		t.Errorf("Should ingest data correctly %v", err)
	}
	if len(cacheObj.ListFiles()) != 1 {
		t.Errorf("should return our file, %v", cacheObj.ListFiles())
	}
	reader, _, err := cacheObj.GetFile(cacheObj.ListFiles()[0].Key)
	if err != nil {
		t.Errorf("Should get file reader %v", err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Errorf("Should be able to read file ok %v", err)
	}
	datas := string(data)
	if datas != "TestData" {
		t.Errorf("Should read back contents >%s< != >%s<", datas, "TestData")
	}
}

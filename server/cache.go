package server

import (
	"github.com/allegro/bigcache"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type CachedInstance struct {
	RelativePath string
	AbsolutePath string
	Data         []byte
	ContentType  string
	Get          func(cache *CachedInstance) (string, []byte)
	GetExpiry    func(cache *CachedInstance, big *bigcache.BigCache) (string, []byte)
}

func (cache *CachedInstance) Reset() {
	cache.Get = LoadCache
	cache.GetExpiry = LoadCacheExpiry
}

func IndexCache(cache map[string]*CachedInstance) error {
	for _, dirPath := range viper.GetStringSlice("paths") {
		cleanPath := path.Clean(dirPath)
		if err := indexPath(cache, cleanPath); err != nil {
			return errors.Wrap(err, "error indexing path "+cleanPath)
		}
	}
	return nil
}

func indexPath(cache map[string]*CachedInstance, dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		cleanedPath := strings.ReplaceAll(filepath.Clean(path), "\\", "/")
		cleanedPath = cleanedPath[len(dirPath)-1:]

		if !strings.HasPrefix(cleanedPath, "/") {
			cleanedPath = "/" + cleanedPath
		}

		// TODO Optional storing in memory on index
		absPath, _ := filepath.Abs(path)
		cache[cleanedPath] = &CachedInstance{
			RelativePath: cleanedPath,
			AbsolutePath: absPath,
			Get:          LoadCache,
			GetExpiry:    LoadCacheExpiry,
		}

		return nil
	})
}

// TODO Streaming
func load(instance *CachedInstance) (string, []byte) {
	fileName := filepath.Base(instance.AbsolutePath)
	fileType := mime.TypeByExtension(filepath.Ext(fileName))

	data, err := ioutil.ReadFile(instance.AbsolutePath)
	if err != nil {
		log.Error(errors.Wrap(err, "error reading file"))
		return "", nil
	}

	if fileType == "" {
		fileType = http.DetectContentType(data[:512])
	}

	log.Debugf("Loaded into cache: %s", instance.AbsolutePath)

	return fileType, data
}

func LoadCache(instance *CachedInstance) (string, []byte) {
	fileType, data := load(instance)

	instance.Data = data
	instance.ContentType = fileType
	instance.Get = func(cache *CachedInstance) (string, []byte) {
		return cache.ContentType, cache.Data
	}

	return fileType, data
}

func LoadCacheExpiry(instance *CachedInstance, big *bigcache.BigCache) (string, []byte) {
	fileType, data := load(instance)

	if err := big.Set(instance.RelativePath, data); err != nil {
		log.Error(errors.Wrap(err, "error setting cache"))
		return "", nil
	}

	instance.ContentType = fileType
	instance.GetExpiry = func(cache *CachedInstance, _ *bigcache.BigCache) (string, []byte) {
		data, _ := big.Get(cache.RelativePath)
		return cache.ContentType, data
	}

	return fileType, data
}

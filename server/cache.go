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

func IndexCache(cache map[string]*CachedInstance) (int64, error) {
	totalSize := int64(0)
	for _, dirPath := range viper.GetStringSlice("paths") {
		cleanPath := path.Clean(dirPath)
		pathSize, err := indexPath(cache, cleanPath)
		if err != nil {
			return 0, errors.Wrap(err, "error indexing path "+cleanPath)
		}
		totalSize += pathSize
	}
	return totalSize, nil
}

func indexPath(cache map[string]*CachedInstance, dirPath string) (int64, error) {
	trimmed := strings.Trim(dirPath, "/")
	toRemove := len(strings.Split(trimmed, "/"))

	if trimmed == "." || trimmed == "" {
		toRemove = 0
	}

	totalSize := int64(0)

	if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		filePath := path

		if info.IsDir() {
			indexFile := viper.GetString("index.file")
			if indexFile != "" {
				joined := filepath.Join(path, indexFile)
				_, err := os.Stat(joined)
				if err != nil && !os.IsNotExist(err) {
					return err
				} else if err != nil {
					return nil
				}
				filePath = joined
			} else {
				return nil
			}
		}

		absPath, _ := filepath.Abs(filePath)

		cleanedPath := strings.ReplaceAll(filepath.Clean(path), "\\", "/")

		// Remove the initial path
		cleanedPath = strings.Join(strings.Split(cleanedPath, "/")[toRemove:], "/")

		if !strings.HasPrefix(cleanedPath, "/") {
			cleanedPath = "/" + cleanedPath
		}

		log.Tracef("Indexed: %s -> %s", cleanedPath, absPath)

		instance := &CachedInstance{
			RelativePath: cleanedPath,
			AbsolutePath: absPath,
			Get:          LoadCache,
			GetExpiry:    LoadCacheExpiry,
		}

		cache[cleanedPath] = instance

		if viper.GetBool("warmup") {
			if viper.GetBool("expiry") {
				panic("expiry not supported if warmup is enabled")
			}

			instance.Get(instance)
			totalSize += int64(len(instance.Data))
		}

		return nil
	}); err != nil {
		return 0, err
	}

	return totalSize, nil
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

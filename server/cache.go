package server

import (
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
	AbsolutePath string
	Data         []byte
	ContentType  string
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
		cleanedPath = cleanedPath[len(dirPath):]

		if !strings.HasPrefix(cleanedPath, "/") {
			cleanedPath = "/" + cleanedPath
		}

		// TODO Optional storing in memory on index
		absPath, _ := filepath.Abs(path)
		cache[cleanedPath] = &CachedInstance{
			AbsolutePath: absPath,
		}

		return nil
	})
}

// TODO Streaming
func LoadCache(cache *map[string]*CachedInstance, path []byte) *CachedInstance {
	instance := (*cache)[string(path)]

	fileName := filepath.Base(instance.AbsolutePath)
	fileType := mime.TypeByExtension(filepath.Ext(fileName))

	data, err := ioutil.ReadFile(instance.AbsolutePath)
	if err != nil {
		log.Error(errors.Wrap(err, "error reading file"))
		instance.Data = []byte{}
		return instance
	}

	if fileType == "" {
		fileType = http.DetectContentType(data[:512])
	}

	instance.Data = data
	instance.ContentType = fileType

	log.Debugf("Loaded into cache: %s", path)

	return instance
}

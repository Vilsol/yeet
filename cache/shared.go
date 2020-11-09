package cache

import (
	"github.com/Vilsol/yeet/utils"
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

type IndexFunc = func(absolutePath string, cleanedPath string) int64

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

func index(f IndexFunc) (int64, error) {
	totalSize := int64(0)
	totalCount := int64(0)
	for _, dirPath := range viper.GetStringSlice("paths") {
		cleanPath := path.Clean(dirPath)
		pathSize, pathCount, err := indexPath(cleanPath, f)
		if err != nil {
			return 0, errors.Wrap(err, "error indexing path "+cleanPath)
		}
		totalSize += pathSize
		totalCount += pathCount
	}

	if viper.GetBool("warmup") {
		log.Infof("Indexed %d files with %s of memory usage", totalCount, utils.ByteCountToHuman(totalSize))
	} else {
		log.Infof("Indexed %d files", totalCount)
	}

	return totalSize, nil
}

func indexPath(dirPath string, f IndexFunc) (int64, int64, error) {
	totalSize := int64(0)
	totalCount := int64(0)

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
		cleanedPath := cleanPath(path, dirPath)
		totalSize += f(absPath, cleanedPath)
		totalCount++

		log.Tracef("Indexed: %s -> %s", cleanedPath, absPath)

		return nil
	}); err != nil {
		return 0, 0, err
	}

	return totalSize, totalCount, nil
}

func cleanPath(path string, dirPath string) string {
	trimmed := strings.Trim(strings.ReplaceAll(filepath.Clean(dirPath), "\\", "/"), "/")
	toRemove := len(strings.Split(trimmed, "/"))

	if trimmed == "." || trimmed == "" {
		toRemove = 0
	}

	cleanedPath := strings.ReplaceAll(filepath.Clean(path), "\\", "/")

	// Remove the initial path
	cleanedPath = strings.Join(strings.Split(cleanedPath, "/")[toRemove:], "/")

	if !strings.HasPrefix(cleanedPath, "/") {
		cleanedPath = "/" + cleanedPath
	}

	return cleanedPath
}

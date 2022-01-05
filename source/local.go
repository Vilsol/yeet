package source

import (
	"github.com/Vilsol/yeet/utils"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

var _ Source = (*Local)(nil)

type Local struct {
}

func (l Local) Get(path string, host []byte) *utils.StreamHijacker {
	file, err := os.OpenFile(path, os.O_RDONLY, 0664)
	if err != nil {
		log.Err(err).Msg("error reading file")
		return nil
	}

	stat, err := file.Stat()
	if err != nil {
		log.Err(err).Msg("error reading file")
		return nil
	}

	fileType := mime.TypeByExtension(filepath.Ext(filepath.Base(path)))

	return utils.NewStreamHijacker(int(stat.Size()), fileType, file)
}

func (l Local) IndexPath(dir string, f IndexFunc) (int64, int64, error) {
	totalSize := int64(0)
	totalCount := int64(0)
	fileNames := make([]string, 0)

	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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
		cleanedPath := cleanPath(path, dir)
		totalSize += f(absPath, cleanedPath)
		totalCount++

		log.Trace().Msgf("Indexed: %s -> %s", cleanedPath, absPath)

		fileNames = append(fileNames, cleanedPath)

		return nil
	}); err != nil {
		return 0, 0, err
	}

	filter := bloom.NewWithEstimates(uint(totalCount), 0.001)
	for _, fileName := range fileNames {
		filter.AddString(fileName)
	}

	log.Trace().Msgf("Created bloom filter of size %d and hash count %d", filter.Cap(), filter.K())

	return totalSize, totalCount, nil
}

func (l Local) Watch() (<-chan WatchEvent, error) {
	events := make(chan WatchEvent, 0)

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, err
	}

	go func() {
		for event := range watcher.Events {
			dirPath := ""
			for _, path := range viper.GetStringSlice("paths") {
				if strings.HasPrefix(event.Name, filepath.Clean(path)) {
					dirPath = path
				}
			}

			if dirPath == "" {
				log.Warn().Msgf("Received update about an unknown path: %s", event.Name)
				continue
			}

			clean := cleanPath(event.Name, dirPath)
			absPath, _ := filepath.Abs(event.Name)

			switch {
			case event.Op&fsnotify.Rename == fsnotify.Rename:
				log.Trace().Msgf("Received rename: %s", clean)
				events <- WatchEvent{
					CleanPath: clean,
					AbsPath:   absPath,
					Op:        WatchRename,
				}
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				log.Trace().Msgf("Received remove: %s", clean)
				events <- WatchEvent{
					CleanPath: clean,
					AbsPath:   absPath,
					Op:        WatchDelete,
				}
			case event.Op&fsnotify.Create == fsnotify.Create:
				log.Trace().Msgf("Received create: %s", clean)
				events <- WatchEvent{
					CleanPath: clean,
					AbsPath:   absPath,
					Op:        WatchCreate,
				}
			case event.Op&fsnotify.Write == fsnotify.Write:
				log.Trace().Msgf("Received write: %s", clean)
				events <- WatchEvent{
					CleanPath: clean,
					AbsPath:   absPath,
					Op:        WatchModify,
				}
			}
		}
	}()

	for _, dirPath := range viper.GetStringSlice("paths") {
		log.Debug().Msgf("Watching path: %s", dirPath)
		err = watcher.Add(dirPath)
		if err != nil {
			return nil, err
		}
	}

	return events, nil
}

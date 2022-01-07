package cache

import (
	"bytes"
	"github.com/Vilsol/yeet/source"
	"github.com/Vilsol/yeet/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"io"
	"path"
	"time"
)

func indexBase(c Cache) (int64, error) {
	return index(c.Source(), func(absolutePath string, cleanedPath string) int64 {
		instance := &commonInstance{
			Instance: &CachedInstance{
				RelativePath: cleanedPath,
				AbsolutePath: absolutePath,
			},
			Get: load(c),
		}

		// Host is not used when indexing is supported
		c.Store(utils.UnsafeGetBytes(cleanedPath), nil, instance)

		if viper.GetBool("warmup") {
			// Host is not used when indexing is supported
			instance.Get(instance, nil)
			return int64(len(instance.Instance.Data))
		}

		return 0
	})
}

func index(source source.Source, f source.IndexFunc) (int64, error) {
	totalSize := int64(0)
	totalCount := int64(0)
	for _, dirPath := range viper.GetStringSlice("paths") {
		cleanPath := path.Clean(dirPath)
		pathSize, pathCount, err := source.IndexPath(cleanPath, f)
		if err != nil {
			return 0, errors.Wrap(err, "error indexing path "+cleanPath)
		}
		totalSize += pathSize
		totalCount += pathCount
	}

	if viper.GetBool("warmup") {
		log.Info().Msgf("Indexed %d files with %s of memory usage", totalCount, utils.ByteCountToHuman(totalSize))
	} else {
		log.Info().Msgf("Indexed %d files", totalCount)
	}

	return totalSize, nil
}

func load(c Cache) func(*commonInstance, []byte) (string, io.Reader, int, bool) {
	return func(instance *commonInstance, host []byte) (string, io.Reader, int, bool) {
		hijacker, failed := c.Source().Get(instance.Instance.AbsolutePath, host)

		if hijacker == nil {
			return "", nil, 0, failed
		}

		log.Debug().Msgf("Loaded file [%d][%s]: %s", hijacker.Size, hijacker.FileType(), instance.Instance.AbsolutePath)

		hijacker.OnClose = func(h *utils.StreamHijacker) {
			instance.Instance.LoadTime = time.Now()
			instance.Instance.Data = hijacker.Buffer
			instance.Instance.ContentType = hijacker.FileType()
			instance.Get = func(cache *commonInstance, _ []byte) (string, io.Reader, int, bool) {
				return cache.Instance.ContentType, bytes.NewReader(cache.Instance.Data), len(cache.Instance.Data), false
			}
		}

		return hijacker.FileType(), hijacker, hijacker.Size, false
	}
}

func expiry(c Cache) {
	go func(c Cache) {
		ticker := time.NewTicker(viper.GetDuration("expiry.interval"))
		defer ticker.Stop()
		for range ticker.C {
			cleanBefore := time.Now().Add(viper.GetDuration("expiry.time") * -1)
			for keyVal := range c.Iter() {
				if keyVal.Value.Instance.Data != nil && keyVal.Value.Instance.LoadTime.Before(cleanBefore) {
					keyVal.Value.Get = load(c)
					keyVal.Value.Instance.Data = nil
					keyVal.Value.Instance.ContentType = ""
					log.Trace().Msgf("Evicted from cache: %s", keyVal.Key)
				}
			}
		}
	}(c)
}

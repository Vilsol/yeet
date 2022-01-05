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
	"reflect"
	"time"
	"unsafe"
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
		c.Store(unsafeGetBytes(cleanedPath), nil, instance)

		if viper.GetBool("warmup") {
			instance.Get(instance)
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

func load(c Cache) func(*commonInstance) (string, io.Reader, int) {
	return func(instance *commonInstance) (string, io.Reader, int) {
		hijacker := c.Source().Get(instance.Instance.AbsolutePath, nil)

		log.Debug().Msgf("Loaded file [%d][%s]: %s", hijacker.Size, hijacker.FileType(), instance.Instance.AbsolutePath)

		hijacker.OnClose = func(h *utils.StreamHijacker) {
			instance.Instance.LoadTime = time.Now()
			instance.Instance.Data = hijacker.Buffer
			instance.Instance.ContentType = hijacker.FileType()
			instance.Get = func(cache *commonInstance) (string, io.Reader, int) {
				return cache.Instance.ContentType, bytes.NewReader(cache.Instance.Data), len(cache.Instance.Data)
			}
		}

		return hijacker.FileType(), hijacker, hijacker.Size
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

func unsafeGetBytes(s string) []byte {
	return (*[0x7fff0000]byte)(unsafe.Pointer(
		(*reflect.StringHeader)(unsafe.Pointer(&s)).Data),
	)[:len(s):len(s)]
}

func byteSliceToString(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

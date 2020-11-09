package cache

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

type readOnlyInstance struct {
	Instance *CachedInstance
	Get      func(instance *readOnlyInstance) (string, []byte)
}

type readOnlyCache struct {
	Cache
	data map[string]*readOnlyInstance
}

func NewReadOnlyCache() (*readOnlyCache, error) {
	if viper.GetBool("watch") {
		return nil, errors.New("read only cache doesn't support file watching")
	}

	c := &readOnlyCache{
		data: make(map[string]*readOnlyInstance),
	}

	if viper.GetBool("expiry") {
		go func(c *readOnlyCache) {
			ticker := time.NewTicker(viper.GetDuration("expiry.interval"))
			defer ticker.Stop()
			for range ticker.C {
				cleanBefore := time.Now().Add(viper.GetDuration("expiry.time") * -1)
				for key, instance := range c.data {
					if instance.Instance.Data != nil && instance.Instance.LoadTime.Before(cleanBefore) {
						instance.Get = loadReadOnly
						instance.Instance.Data = nil
						instance.Instance.ContentType = ""
						log.Tracef("Evicted from cache: %s", key)
					}
				}
			}
		}(c)
	}

	return c, nil
}

func (c *readOnlyCache) Index() (int64, error) {
	return index(func(absolutePath string, cleanedPath string) int64 {
		instance := &readOnlyInstance{
			Instance: &CachedInstance{
				RelativePath: cleanedPath,
				AbsolutePath: absolutePath,
			},
			Get: loadReadOnly,
		}

		c.data[cleanedPath] = instance

		if viper.GetBool("warmup") {
			if viper.GetBool("expiry") {
				panic("expiry not supported if warmup is enabled")
			}

			instance.Get(instance)
			return int64(len(instance.Instance.Data))
		}

		return 0
	})
}

func (c *readOnlyCache) Get(path []byte) (string, []byte) {
	if instance, ok := c.data[string(path)]; ok {
		return instance.Get(instance)
	}

	return "", nil
}

func loadReadOnly(instance *readOnlyInstance) (string, []byte) {
	fileType, data := load(instance.Instance)

	instance.Instance.LoadTime = time.Now()
	instance.Instance.Data = data
	instance.Instance.ContentType = fileType
	instance.Get = func(cache *readOnlyInstance) (string, []byte) {
		return cache.Instance.ContentType, cache.Instance.Data
	}

	return fileType, data
}

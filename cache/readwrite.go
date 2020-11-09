package cache

import (
	"github.com/cornelk/hashmap"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
	"time"
)

type readWriteInstance struct {
	Instance *CachedInstance
	Get      func(instance *readWriteInstance) (string, []byte)
}

type readWriteCache struct {
	Cache
	data *hashmap.HashMap
}

func NewReadWriteCache() (*readWriteCache, error) {
	data := &hashmap.HashMap{}

	c := &readWriteCache{
		data: data,
	}

	if viper.GetBool("expiry") {
		go func(c *readWriteCache) {
			ticker := time.NewTicker(viper.GetDuration("expiry.interval"))
			defer ticker.Stop()
			for range ticker.C {
				cleanBefore := time.Now().Add(viper.GetDuration("expiry.time") * -1)
				for keyVal := range c.data.Iter() {
					instance := keyVal.Value.(*readWriteInstance)
					if instance.Instance.Data != nil && instance.Instance.LoadTime.Before(cleanBefore) {
						instance.Get = loadReadWrite
						instance.Instance.Data = nil
						instance.Instance.ContentType = ""
						log.Tracef("Evicted from cache: %s", keyVal.Key)
					}
				}
			}
		}(c)
	}

	return c, nil
}

func (c *readWriteCache) Index() (int64, error) {
	indexFunc := func(absolutePath string, cleanedPath string) int64 {
		instance := &readWriteInstance{
			Instance: &CachedInstance{
				RelativePath: cleanedPath,
				AbsolutePath: absolutePath,
			},
			Get: loadReadWrite,
		}

		c.data.Set(cleanedPath, instance)

		if viper.GetBool("warmup") {
			if viper.GetBool("expiry") {
				panic("expiry not supported if warmup is enabled")
			}

			instance.Get(instance)
			return int64(len(instance.Instance.Data))
		}

		return 0
	}

	totalCount, err := index(indexFunc)

	if viper.GetBool("watch") {
		watcher, err := fsnotify.NewWatcher()

		if err != nil {
			return 0, err
		}

		go func(c *readWriteCache) {
			for event := range watcher.Events {
				dirPath := ""
				for _, path := range viper.GetStringSlice("paths") {
					if strings.HasPrefix(event.Name, filepath.Clean(path)) {
						dirPath = path
					}
				}

				if dirPath == "" {
					log.Warnf("Received update about an unknown path: %s", event.Name)
					continue
				}

				cleanPath := cleanPath(event.Name, dirPath)

				switch {
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					fallthrough
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					c.data.Del(cleanPath)
					log.Tracef("File removed from disk and cache: %s", cleanPath)
				case event.Op&fsnotify.Create == fsnotify.Create:
					absPath, _ := filepath.Abs(event.Name)
					indexFunc(absPath, cleanPath)
					log.Tracef("File created on disk and added to cache: %s", cleanPath)
				case event.Op&fsnotify.Write == fsnotify.Write:
					instance, ok := c.data.GetStringKey(cleanPath)
					if ok {
						instance.(*readWriteInstance).Get = loadReadWrite
						instance.(*readWriteInstance).Instance.Data = nil
						instance.(*readWriteInstance).Instance.ContentType = ""
						log.Tracef("File updated on disk and evicted from cache: %s", cleanPath)
					}
				}
			}
		}(c)

		for _, dirPath := range viper.GetStringSlice("paths") {
			log.Debugf("Watching path: %s", dirPath)
			err = watcher.Add(dirPath)
			if err != nil {
				return 0, err
			}
		}
	}

	return totalCount, err
}

func (c *readWriteCache) Get(path []byte) (string, []byte) {
	if instance, ok := c.data.GetStringKey(string(path)); ok {
		return instance.(*readWriteInstance).Get(instance.(*readWriteInstance))
	}

	return "", nil
}

func loadReadWrite(instance *readWriteInstance) (string, []byte) {
	fileType, data := load(instance.Instance)

	instance.Instance.LoadTime = time.Now()
	instance.Instance.Data = data
	instance.Instance.ContentType = fileType
	instance.Get = func(cache *readWriteInstance) (string, []byte) {
		return cache.Instance.ContentType, cache.Instance.Data
	}

	return fileType, data
}

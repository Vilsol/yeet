package cache

import (
	"github.com/Vilsol/yeet/source"
	"github.com/Vilsol/yeet/utils"
	"github.com/cornelk/hashmap"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"io"
)

var _ Cache = (*HashMapCache)(nil)

type HashMapCache struct {
	data   *hashmap.HashMap
	source source.Source
	hosts  bool
}

func NewHashMapCache(source source.Source, hosts bool) (*HashMapCache, error) {
	data := &hashmap.HashMap{}

	c := &HashMapCache{
		data:   data,
		source: source,
		hosts:  hosts,
	}

	if viper.GetBool("expiry") {
		expiry(c)
	}

	return c, nil
}

func (c *HashMapCache) Index() (int64, error) {
	totalCount, err := indexBase(c)

	if viper.GetBool("watch") {
		events, err := c.source.Watch()
		if err != nil {
			return 0, err
		}

		go func(c *HashMapCache, events <-chan source.WatchEvent) {
			for event := range events {
				switch event.Op {
				case source.WatchRename:
					fallthrough
				case source.WatchDelete:
					c.data.Del(event.CleanPath)
					log.Trace().Msgf("Removed from cache: %s", event.CleanPath)
				case source.WatchCreate:
					instance := &commonInstance{
						Instance: &CachedInstance{
							RelativePath: event.CleanPath,
							AbsolutePath: event.AbsPath,
						},
						Get: load(c),
					}

					c.data.Set(event.CleanPath, instance)

					if viper.GetBool("warmup") {
						instance.Get(instance, nil)
					}

					log.Trace().Msgf("Added to cache: %s", event.CleanPath)
				case source.WatchModify:
					instance, ok := c.data.GetStringKey(event.CleanPath)
					if ok {
						instance.(*commonInstance).Get = load(c)
						instance.(*commonInstance).Instance.Data = nil
						instance.(*commonInstance).Instance.ContentType = ""
						log.Trace().Msgf("Evicted from cache: %s", event.CleanPath)
					}
				}
			}
		}(c, events)
	}

	return totalCount, err
}

func (c *HashMapCache) Store(path []byte, host []byte, instance *commonInstance) {
	if !c.hosts {
		c.data.Set(path, instance)
	} else {
		c.data.Set(append(host, path...), instance)
	}
}

func (c *HashMapCache) Get(path []byte, host []byte) (string, io.Reader, int, bool) {
	if !c.hosts {
		if instance, ok := c.data.Get(path); ok {
			return instance.(*commonInstance).Get(instance.(*commonInstance), nil)
		}
	} else {
		key := append(host, path...)
		instance, ok := c.data.Get(key)
		if !ok {
			pathStr := utils.ByteSliceToString(path)
			instance = &commonInstance{
				Instance: &CachedInstance{
					RelativePath: pathStr,
					AbsolutePath: pathStr,
				},
				Get: load(c),
			}
			c.data.Set(key, instance)
		}

		if instance != nil {
			return instance.(*commonInstance).Get(instance.(*commonInstance), host)
		}
	}

	return "", nil, 0, false
}

func (c *HashMapCache) Source() source.Source {
	return c.source
}

func (c *HashMapCache) Iter() <-chan KeyValue {
	ch := make(chan KeyValue)
	go func(c *HashMapCache, ch chan KeyValue) {
		for keyVal := range c.data.Iter() {
			ch <- KeyValue{
				Key:   keyVal.Key.(string),
				Value: keyVal.Value.(*commonInstance),
			}
		}
		close(ch)
	}(c, ch)
	return ch
}

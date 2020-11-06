package server

import (
	"github.com/allegro/bigcache"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
)

type ExpiryWebserver struct {
	Cache  map[string]*CachedInstance
	Expiry *bigcache.BigCache
}

func (h *ExpiryWebserver) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	if cache, ok := h.Cache[string(ctx.Path())]; ok {
		ctx.Success(cache.GetExpiry(cache, h.Expiry))
	} else {
		ctx.SetStatusCode(404)
	}
}

func GetExpiryWebserver(cache map[string]*CachedInstance) (*ExpiryWebserver, error) {
	bigCache, err := bigcache.NewBigCache(bigcache.Config{
		Shards:           viper.GetInt("expiry.shards"),
		LifeWindow:       viper.GetDuration("expiry.time"),
		CleanWindow:      viper.GetDuration("expiry.interval"),
		HardMaxCacheSize: viper.GetInt("expiry.memory"),
		OnRemove: func(key string, entry []byte) {
			cache[key].Reset()
		},
	})

	if err != nil {
		return nil, err
	}

	return &ExpiryWebserver{
		Cache:  cache,
		Expiry: bigCache,
	}, nil
}

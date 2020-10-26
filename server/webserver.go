package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
)

type Webserver struct {
	Cache map[string]*CachedInstance
}

func (h *Webserver) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	data, ok := h.Cache[string(ctx.Path())]

	if !ok {
		ctx.SetStatusCode(404)
		return
	}

	if data.Data == nil {
		data = LoadCache(&h.Cache, ctx.Path())
	}

	ctx.SetContentType(data.ContentType)
	ctx.SetBody(data.Data)
	ctx.SetStatusCode(200)
}

func RunWebserver() error {
	cache := make(map[string]*CachedInstance)

	if err := IndexCache(cache); err != nil {
		return err
	}

	log.Infof("Indexed %d files", len(cache))

	ws := &Webserver{
		Cache: cache,
	}

	address := fmt.Sprintf("%s:%d", viper.GetString("host"), viper.GetInt("port"))

	log.Infof("Starting webserver on %s", address)

	if err := fasthttp.ListenAndServe(address, ws.HandleFastHTTP); err != nil {
		return err
	}

	return nil
}

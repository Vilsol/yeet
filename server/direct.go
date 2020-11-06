package server

import (
	"github.com/valyala/fasthttp"
)

type DirectWebserver struct {
	Cache map[string]*CachedInstance
}

func (h *DirectWebserver) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	if data, ok := h.Cache[string(ctx.Path())]; ok {
		ctx.Success(data.Get(data))
	} else {
		ctx.SetStatusCode(404)
	}
}

func GetWebserver(cache map[string]*CachedInstance) Webserver {
	return &DirectWebserver{
		Cache: cache,
	}
}

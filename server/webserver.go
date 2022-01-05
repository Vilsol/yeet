package server

import (
	"github.com/Vilsol/yeet/cache"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"net"
	"strconv"
)

func Run(c cache.Cache) error {
	ws := &Webserver{
		Cache: c,
	}

	address := viper.GetString("host") + ":" + strconv.Itoa(viper.GetInt("port"))

	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	log.Info().Msgf("Starting webserver on %s", address)

	if err := fasthttp.Serve(ln, ws.HandleFastHTTP); err != nil {
		return err
	}

	return nil
}

type Webserver struct {
	Cache cache.Cache
}

func (h *Webserver) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	if fileType, stream, size := h.Cache.Get(ctx.Path(), ctx.Host()); size > 0 {
		ctx.SetContentType(fileType)
		ctx.SetBodyStream(stream, size)
	} else {
		ctx.SetStatusCode(404)
	}
}

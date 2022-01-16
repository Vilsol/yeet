package server

import (
	"github.com/Vilsol/yeet/cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
)

func Run(c cache.Cache) error {
	ws := &Webserver{
		Cache: c,
	}

	if viper.GetBool("404-index") {
		ws.Index404 = []byte("/" + viper.GetString("index-file"))
	}

	address := viper.GetString("host") + ":" + strconv.Itoa(viper.GetInt("port"))

	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	handler := ws.HandleFastHTTP
	if viper.GetString("bot.proxy") != "" {
		r, err := regexp.Compile(`(?i)` + viper.GetString("bot.agents"))
		if err != nil {
			return errors.Wrap(err, "failed to compile bot proxy regex")
		}

		proxy, err := url.Parse(viper.GetString("bot.proxy"))
		if err != nil {
			return errors.Wrap(err, "failed to parse bot proxy url")
		}

		reverseProxy := httputil.NewSingleHostReverseProxy(proxy)
		proxyHandler := fasthttpadaptor.NewFastHTTPHandler(reverseProxy)

		handler = ws.HandleFastHTTPWithBotProxy(r, proxyHandler)
	}

	if viper.GetString("tls.cert") != "" && viper.GetString("tls.key") != "" {
		log.Info().Msgf("Starting webserver with TLS on %s", address)
		return fasthttp.ServeTLS(ln, viper.GetString("tls.cert"), viper.GetString("tls.key"), handler)
	}

	log.Info().Msgf("Starting webserver on %s", address)
	return fasthttp.Serve(ln, handler)
}

type Webserver struct {
	Cache    cache.Cache
	Index404 []byte
}

func (h *Webserver) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	fileType, stream, size, failed := h.Cache.Get(ctx.Path(), ctx.Host())
	if size > 0 {
		ctx.SetContentType(fileType)
		ctx.SetBodyStream(stream, size)
		return
	}

	if failed {
		ctx.SetStatusCode(500)
		return
	}

	if h.Index404 != nil {
		fileType, stream, size, failed := h.Cache.Get(h.Index404, ctx.Host())
		if size > 0 {
			ctx.SetContentType(fileType)
			ctx.SetBodyStream(stream, size)
			return
		} else if failed {
			ctx.SetStatusCode(500)
			return
		}
	}

	ctx.SetStatusCode(404)
}

func (h *Webserver) HandleFastHTTPWithBotProxy(botHeaderRegex *regexp.Regexp, proxy fasthttp.RequestHandler) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		if ctx.UserAgent() == nil || !botHeaderRegex.Match(ctx.UserAgent()) {
			h.HandleFastHTTP(ctx)
		} else {
			proxy(ctx)
		}
	}
}

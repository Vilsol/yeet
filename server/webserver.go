package server

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"strconv"
)

type Webserver interface {
	HandleFastHTTP(ctx *fasthttp.RequestCtx)
}

func Run() error {
	cache := make(map[string]*CachedInstance)

	if err := IndexCache(cache); err != nil {
		return err
	}

	log.Infof("Indexed %d files", len(cache))

	var err error
	var ws Webserver
	if viper.GetBool("expiry.enabled") {
		ws, err = GetExpiryWebserver(cache)
	} else {
		ws = GetWebserver(cache)
	}

	if err != nil {
		return err
	}

	address := viper.GetString("host") + ":" + strconv.Itoa(viper.GetInt("port"))

	log.Infof("Starting webserver on %s", address)

	if err := fasthttp.ListenAndServe(address, ws.HandleFastHTTP); err != nil {
		return err
	}

	return nil
}

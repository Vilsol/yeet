package server

import (
	"github.com/Vilsol/yeet/cache"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"strconv"
)

func Run() error {
	ws, err := GetWebserver()

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

type Webserver struct {
	Cache cache.Cache
}

func (h *Webserver) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	if fileType, b := h.Cache.Get(ctx.Path()); b != nil {
		ctx.Success(fileType, b)
	} else {
		ctx.SetStatusCode(404)
	}
}

func GetWebserver() (*Webserver, error) {
	var c cache.Cache
	var err error

	if viper.GetBool("watch") {
		c, err = cache.NewReadWriteCache()
	} else {
		c, err = cache.NewReadOnlyCache()
	}

	if err != nil {
		return nil, err
	}

	if _, err := c.Index(); err != nil {
		return nil, err
	}

	ws := &Webserver{
		Cache: c,
	}

	return ws, nil
}

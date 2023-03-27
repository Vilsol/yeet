package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"time"
)

const botAgents = "bot|crawl|spider|external|meta|scrap|archive|discourse"

func init() {
	ServeCMD.PersistentFlags().String("host", "", "Hostname to bind the webserver")
	ServeCMD.PersistentFlags().Int("port", 8080, "Port to run the webserver on")

	ServeCMD.PersistentFlags().Bool("warmup", false, "Load all files into memory on startup")

	ServeCMD.PersistentFlags().Bool("expiry", false, "Use cache expiry")
	ServeCMD.PersistentFlags().Duration("expiry-time", time.Minute*60, "Lifetime of a cache entry")
	ServeCMD.PersistentFlags().Duration("expiry-interval", time.Minute*10, "Interval between cache GC's")

	ServeCMD.PersistentFlags().String("index-file", "index.html", "The directory default index file")

	ServeCMD.PersistentFlags().StringSliceP("paths", "p", []string{"./www"}, "Paths to serve on the webserver")
	ServeCMD.PersistentFlags().BoolP("watch", "w", false, "Watch filesystem for changes")

	ServeCMD.PersistentFlags().String("tls-cert", "", "TLS Certificate file path")
	ServeCMD.PersistentFlags().String("tls-key", "", "TLS Key file path")

	ServeCMD.PersistentFlags().String("bot-proxy", "", "Bot proxy URL")
	ServeCMD.PersistentFlags().String("bot-agents", botAgents, "Bot User-Agent header regex")

	ServeCMD.PersistentFlags().Bool("404-index", false, "Redirect any 404 to the index file")
	ServeCMD.PersistentFlags().String("404-fallback", "", "Redirect any 404 to the provided fallback file")

	_ = viper.BindPFlag("paths", ServeCMD.PersistentFlags().Lookup("paths"))
	_ = viper.BindPFlag("watch", ServeCMD.PersistentFlags().Lookup("watch"))

	_ = viper.BindPFlag("host", ServeCMD.PersistentFlags().Lookup("host"))
	_ = viper.BindPFlag("port", ServeCMD.PersistentFlags().Lookup("port"))

	_ = viper.BindPFlag("warmup", ServeCMD.PersistentFlags().Lookup("warmup"))

	_ = viper.BindPFlag("expiry", ServeCMD.PersistentFlags().Lookup("expiry"))
	_ = viper.BindPFlag("expiry.time", ServeCMD.PersistentFlags().Lookup("expiry-time"))
	_ = viper.BindPFlag("expiry.interval", ServeCMD.PersistentFlags().Lookup("expiry-interval"))

	_ = viper.BindPFlag("index.file", ServeCMD.PersistentFlags().Lookup("index-file"))

	_ = viper.BindPFlag("tls.cert", ServeCMD.PersistentFlags().Lookup("tls-cert"))
	_ = viper.BindPFlag("tls.key", ServeCMD.PersistentFlags().Lookup("tls-key"))

	_ = viper.BindPFlag("bot.proxy", ServeCMD.PersistentFlags().Lookup("bot-proxy"))
	_ = viper.BindPFlag("bot.agents", ServeCMD.PersistentFlags().Lookup("bot-agents"))

	_ = viper.BindPFlag("404-index", ServeCMD.PersistentFlags().Lookup("404-index"))
	_ = viper.BindPFlag("404-fallback", ServeCMD.PersistentFlags().Lookup("404-fallback"))

	RootCMD.AddCommand(ServeCMD)
}

var ServeCMD = &cobra.Command{
	Use:   "serve",
	Short: "Serve files with yeet",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("warmup") && viper.GetBool("expiry") {
			return errors.New("expiry not supported if warmup is enabled")
		}
		return nil
	},
}

package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "yeet",
	Short: "yeet is an in-memory indexed static file webserver",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.SetEnvPrefix("yeet")
		viper.AutomaticEnv()

		_ = viper.ReadInConfig()

		level, err := log.ParseLevel(viper.GetString("log"))

		if err != nil {
			panic(err)
		}

		log.SetFormatter(&log.TextFormatter{
			ForceColors: viper.GetBool("colors"),
		})
		log.SetOutput(os.Stdout)
		log.SetLevel(level)
	},
}

func Execute() {
	// Allow running from explorer
	cobra.MousetrapHelpText = ""

	// Execute serve command as default
	cmd, _, err := rootCmd.Find(os.Args[1:])
	if (len(os.Args) <= 1 || os.Args[1] != "help") && (err != nil || cmd == rootCmd) {
		args := append([]string{"serve"}, os.Args[1:]...)
		rootCmd.SetArgs(args)
	}

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.PersistentFlags().String("log", "info", "The log level to output")
	rootCmd.PersistentFlags().Bool("colors", false, "Force output with colors")

	rootCmd.PersistentFlags().String("host", "0.0.0.0", "Hostname to bind the webserver")
	rootCmd.PersistentFlags().Int("port", 8080, "Port to run the webserver on")

	rootCmd.PersistentFlags().StringSlice("paths", []string{"./www"}, "Paths to serve on the webserver")
	rootCmd.PersistentFlags().Bool("warmup", false, "Load all files into memory on startup")
	rootCmd.PersistentFlags().Bool("watch", false, "Watch filesystem for changes")

	rootCmd.PersistentFlags().Bool("expiry", false, "Use cache expiry")
	rootCmd.PersistentFlags().Duration("expiry-time", time.Minute*60, "Lifetime of a cache entry")
	rootCmd.PersistentFlags().Duration("expiry-interval", time.Minute*10, "Interval between cache GC's")

	rootCmd.PersistentFlags().String("index-file", "index.html", "The directory default index file")

	_ = viper.BindPFlag("log", rootCmd.PersistentFlags().Lookup("log"))
	_ = viper.BindPFlag("colors", rootCmd.PersistentFlags().Lookup("colors"))

	_ = viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	_ = viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))

	_ = viper.BindPFlag("paths", rootCmd.PersistentFlags().Lookup("paths"))
	_ = viper.BindPFlag("warmup", rootCmd.PersistentFlags().Lookup("warmup"))
	_ = viper.BindPFlag("watch", rootCmd.PersistentFlags().Lookup("watch"))

	_ = viper.BindPFlag("expiry", rootCmd.PersistentFlags().Lookup("expiry"))
	_ = viper.BindPFlag("expiry.time", rootCmd.PersistentFlags().Lookup("expiry-time"))
	_ = viper.BindPFlag("expiry.interval", rootCmd.PersistentFlags().Lookup("expiry-interval"))

	_ = viper.BindPFlag("index.file", rootCmd.PersistentFlags().Lookup("index-file"))
}

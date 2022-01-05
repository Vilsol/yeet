package cmd

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
	"time"
)

var RootCMD = &cobra.Command{
	Use:   "yeet",
	Short: "yeet is an in-memory indexed static file webserver",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.SetEnvPrefix("yeet")
		viper.AutomaticEnv()

		_ = viper.ReadInConfig()

		level, err := zerolog.ParseLevel(viper.GetString("log"))
		if err != nil {
			panic(err)
		}

		zerolog.SetGlobalLevel(level)

		writers := make([]io.Writer, 0)
		if !viper.GetBool("quiet") {
			writers = append(writers, zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC3339,
			})
		}

		if viper.GetString("log-file") != "" {
			logFile, err := os.OpenFile(viper.GetString("log-file"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

			if err != nil {
				return errors.Wrap(err, "failed to open log file")
			}

			writers = append(writers, logFile)
		}

		log.Logger = zerolog.New(io.MultiWriter(writers...)).With().Timestamp().Logger()

		return nil
	},
}

func Execute() {
	// Allow running from explorer
	cobra.MousetrapHelpText = ""

	// Execute serve local command as default
	cmd, _, err := RootCMD.Find(os.Args[1:])
	if (len(os.Args) <= 1 || os.Args[1] != "help") && (err != nil || cmd == RootCMD) {
		args := append([]string{"serve", "local"}, os.Args[1:]...)
		RootCMD.SetArgs(args)
	}

	if err := RootCMD.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	RootCMD.PersistentFlags().String("log", "info", "The log level to output")
	RootCMD.PersistentFlags().String("log-file", "", "File to output logs to")
	RootCMD.PersistentFlags().Bool("quiet", false, "Do not log anything to console")

	_ = viper.BindPFlag("log", RootCMD.PersistentFlags().Lookup("log"))
	_ = viper.BindPFlag("log-file", RootCMD.PersistentFlags().Lookup("log-file"))
	_ = viper.BindPFlag("quiet", RootCMD.PersistentFlags().Lookup("quiet"))
}

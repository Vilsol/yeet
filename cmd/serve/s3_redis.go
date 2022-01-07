package serve

import (
	"github.com/Vilsol/yeet/cache"
	"github.com/Vilsol/yeet/cmd"
	"github.com/Vilsol/yeet/server"
	"github.com/Vilsol/yeet/source"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	s3RedisCmd.Flags().String("network", "tcp", "The network type, either tcp or unix")
	s3RedisCmd.Flags().StringP("address", "a", "localhost:6379", "host:port address of Redis")
	s3RedisCmd.Flags().String("user", "", "Username of Redis")
	s3RedisCmd.Flags().String("pass", "", "Password of Redis")
	s3RedisCmd.Flags().Int("db", 0, "DB of Redis")

	_ = s3RedisCmd.MarkFlagRequired("address")

	_ = viper.BindPFlag("network", s3RedisCmd.Flags().Lookup("network"))
	_ = viper.BindPFlag("address", s3RedisCmd.Flags().Lookup("address"))
	_ = viper.BindPFlag("username", s3RedisCmd.Flags().Lookup("username"))
	_ = viper.BindPFlag("password", s3RedisCmd.Flags().Lookup("password"))
	_ = viper.BindPFlag("db", s3RedisCmd.Flags().Lookup("db"))

	cmd.ServeCMD.AddCommand(s3RedisCmd)
}

var s3RedisCmd = &cobra.Command{
	Use:   "s3-redis",
	Short: "Serve Redis-backed S3 buckets",
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("watch") {
			return errors.New("watch is not supported for s3-redis")
		}

		src, err := source.NewS3Redis(
			viper.GetString("network"),
			viper.GetString("address"),
			viper.GetString("username"),
			viper.GetString("password"),
			viper.GetInt("db"),
		)

		if err != nil {
			return err
		}

		c, err := cache.NewHashMapCache(src, true)

		if err != nil {
			return err
		}

		if _, err := c.Index(); err != nil {
			return err
		}

		return server.Run(c)
	},
}

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
	s3Cmd.Flags().StringP("bucket", "b", "", "S3 Bucket to fetch from")
	s3Cmd.Flags().StringP("key", "k", "", "S3 Key of the account")
	s3Cmd.Flags().StringP("secret", "s", "", "S3 Secret of the account")
	s3Cmd.Flags().StringP("endpoint", "e", "", "S3 Endpoint")
	s3Cmd.Flags().String("region", "us-west-002", "S3 Region of the bucket")

	_ = s3Cmd.MarkFlagRequired("bucket")
	_ = s3Cmd.MarkFlagRequired("key")
	_ = s3Cmd.MarkFlagRequired("secret")
	_ = s3Cmd.MarkFlagRequired("endpoint")

	_ = viper.BindPFlag("bucket", s3Cmd.Flags().Lookup("bucket"))
	_ = viper.BindPFlag("key", s3Cmd.Flags().Lookup("key"))
	_ = viper.BindPFlag("secret", s3Cmd.Flags().Lookup("secret"))
	_ = viper.BindPFlag("endpoint", s3Cmd.Flags().Lookup("endpoint"))
	_ = viper.BindPFlag("region", s3Cmd.Flags().Lookup("region"))

	cmd.ServeCMD.AddCommand(s3Cmd)
}

var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "Serve an S3 bucket",
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("watch") {
			return errors.New("watch is not supported for s3")
		}

		src, err := source.NewS3(
			viper.GetString("bucket"),
			viper.GetString("key"),
			viper.GetString("secret"),
			viper.GetString("endpoint"),
			viper.GetString("region"),
		)

		if err != nil {
			return err
		}

		c, err := cache.NewHashMapCache(src, false)

		if err != nil {
			return err
		}

		if _, err := c.Index(); err != nil {
			return err
		}

		return server.Run(c)
	},
}

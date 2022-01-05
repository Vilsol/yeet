package serve

import (
	"github.com/Vilsol/yeet/cache"
	"github.com/Vilsol/yeet/cmd"
	"github.com/Vilsol/yeet/server"
	"github.com/Vilsol/yeet/source"
	"github.com/spf13/cobra"
)

func init() {
	cmd.ServeCMD.AddCommand(localCmd)
}

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Serve a local directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		src := source.Local{}
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

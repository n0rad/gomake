package gomake

import (
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
	"os"
)

type StepClean struct {
}

func (c *StepClean) Init() error {
	return nil
}

func (c *StepClean) Name() string {
	return "clean"
}

func (c *StepClean) GetCommand() *cobra.Command {
	var tools bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "clean build directory",
		RunE: commandDurationWrapper(func(cmd *cobra.Command, args []string) error {
			if tools {
				logs.Info("Cleaning tools")
				return os.RemoveAll("./dist-tools/")
			} else {
				logs.Info("Cleaning")
				return os.RemoveAll("./dist/")
			}
		}),
	}

	cmd.Flags().BoolVarP(&tools, "tools", "t", false, "clean build tools")

	return cmd
}

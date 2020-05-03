package gomake

import (
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
	"os"
)

type StepClean struct {
	project *Project
}

func (c *StepClean) Init(project *Project) error {
	c.project = project
	return nil
}

func (c *StepClean) Name() string {
	return "clean"
}

func (c *StepClean) GetCommand() *cobra.Command {
	var tools bool

	cmd := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,
		Use:           "clean",
		Short:         "clean build directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := CommandDurationWrapper(cmd, func() error {
				ColorPrintln("Cleaning", HGreen)
				if err := os.RemoveAll("./dist/"); err != nil {
					return err
				}

				if tools {
					logs.Info("Cleaning tools")
					if err := os.RemoveAll("./dist-tools/"); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				return err
			}
			return c.project.processArgs(args)
		},
	}

	cmd.Flags().BoolVarP(&tools, "tools", "t", false, "Also clean build tools")

	return cmd
}

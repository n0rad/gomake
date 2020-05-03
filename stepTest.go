package gomake

import (
	"github.com/spf13/cobra"
)

type StepTest struct {
	project *Project
}

func (c *StepTest) Init(project *Project) error {
	c.project = project
	return nil
}

func (c *StepTest) Name() string {
	return "test"
}

func (c *StepTest) GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,
		Use:           "test",
		Short:         "run tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := commandDurationWrapper(cmd, func() error {
				ColorPrintln("Testing", HGreen)
				err := ExecShell("go test $(go list ./... | grep -v '/vendor/')")
				if err != nil {
					return err
				}
				return nil
			}); err != nil {
				return err
			}
			return c.project.processArgs(args)
		},
	}
	return cmd
}

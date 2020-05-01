package gomake

import (
	"github.com/spf13/cobra"
)

type StepQuality struct {
	Lint        *bool
	Vet         *bool
	Misspell    *bool
	Ineffassign *bool
	Gocyclo     *bool
}

func (c *StepQuality) Init() error {
	return nil
}

func (c *StepQuality) Name() string {
	return "quality"
}

func (c *StepQuality) GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quality",
		Short: "check code quality",
		RunE: commandDurationWrapper(func(cmd *cobra.Command, args []string) error {
			return nil
		}),
	}
	return cmd
}

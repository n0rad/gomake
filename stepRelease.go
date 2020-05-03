package gomake

import (
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
)

type StepRelease struct {
	project       *Project
	OsArchRelease []string
	Upx           bool
}

func (c *StepRelease) Init(project *Project) error {
	c.project = project

	if len(c.OsArchRelease) == 0 {
		c.OsArchRelease = append(c.OsArchRelease, "linux-amd64")
		c.OsArchRelease = append(c.OsArchRelease, "darwin-amd64")
		c.OsArchRelease = append(c.OsArchRelease, "windows-amd64")
	}

	return nil
}

func (c *StepRelease) Name() string {
	return "release"
}

func (c *StepRelease) GetCommand() *cobra.Command {
	var token string
	var version string

	cmd := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,
		Use:           "release",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := commandDurationWrapper(cmd, func() error {
				ColorPrintln("Releasing", HGreen)

				logs.WithField("token", token).WithField("version", version).Info("Release")

				return nil
			}); err != nil {
				return err
			}
			return c.project.processArgs(args)
		},
	}

	cmd.Flags().StringVarP(&token, "token", "t", "", "token")
	cmd.Flags().StringVarP(&version, "version", "v", "", "version")

	return cmd
}

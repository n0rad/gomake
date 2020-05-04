package gomake

import (
	"github.com/n0rad/go-erlog/errs"
	"github.com/spf13/cobra"
	"os"
)

type StepRelease struct {
	project         *Project
	OsArchRelease   []string
	Upx             *bool
	PostReleaseHook func(StepRelease) error // upload

}

func (c *StepRelease) Init(project *Project) error {
	c.project = project

	if len(c.OsArchRelease) == 0 {
		c.OsArchRelease = append(c.OsArchRelease, "linux-amd64")
		c.OsArchRelease = append(c.OsArchRelease, "darwin-amd64")
		c.OsArchRelease = append(c.OsArchRelease, "windows-amd64")
	}

	if c.Upx == nil {
		c.Upx = False
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
			if err := CommandDurationWrapper(cmd, func() error {
				ColorPrintln("Releasing", HGreen)

				// clean
				cleanCommand := c.project.steps["clean"].GetCommand()
				cleanCommand.SetArgs([]string{})
				if err := cleanCommand.Execute(); err != nil {
					return errs.WithE(err, "Cannot release, clean failed")
				}

				// build
				build := c.project.steps["build"].(*StepBuild)
				build.Upx = c.Upx
				build.Programs = []Program{}
				for _, osArch := range c.OsArchRelease {
					build.Programs = append(build.Programs, Program{OsArch: osArch})
				}
				if err := build.Init(c.project); err != nil {
					return errs.WithE(err, "Failed to re-init build for release")
				}
				buildCommand := build.GetCommand()
				buildCommand.SetArgs([]string{})
				if err := buildCommand.Execute(); err != nil {
					return errs.WithE(err, "Cannot release, build failed")
				}

				// test
				testCommand := c.project.steps["test"].GetCommand()
				testCommand.SetArgs([]string{})
				if err := testCommand.Execute(); err != nil {
					return errs.WithE(err, "Cannot release, test failed")
				}

				// check
				checkCommand := c.project.steps["check"].GetCommand()
				checkCommand.SetArgs([]string{})
				if err := checkCommand.Execute(); err != nil {
					return errs.WithE(err, "Cannot release, check failed")
				}

				if err := IsGitWorkTreeClean(); err != nil {
					return errs.WithE(err, "git repository is not clean")
				}

				// compressing
				for _, p := range build.Programs {
					if err := c.compressRelease(p); err != nil {
						return errs.WithE(err, "Compression failed")
					}
				}

				if c.PostReleaseHook != nil {
					if err := c.PostReleaseHook(*c); err != nil {
						return errs.WithE(err, "Post release hook failed")
					}
				}

				return nil
			}); err != nil {
				return err
			}
			return c.project.processArgs(args)
		},
	}

	cmd.Flags().StringVarP(&token, "token", "t", "", "token")
	cmd.Flags().StringVarP(&version, "version", "v", "", "version")
	RegisterLogLevelParser(cmd)

	return cmd
}

func (c StepRelease) compressRelease(p Program) error {
	fileToWrite, err := os.OpenFile("./dist/"+c.project.name+"-"+p.OsArch+".tar.gz", os.O_CREATE|os.O_RDWR, os.FileMode(600))
	if err != nil {
		return errs.WithE(err, "Failed to open compressed release file") // TODO
	}
	defer fileToWrite.Close()
	if err := CompressToTarGzDirectory("./dist/"+c.project.name+"-"+p.OsArch, fileToWrite); err != nil {
		return errs.WithE(err, "Failed to compress dir to tar.gz")
	}
	return nil
}

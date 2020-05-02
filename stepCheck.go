package gomake

import (
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
	"os"
)

type StepCheck struct {
	Lint        *bool
	Vet         *bool
	Misspell    *bool
	Ineffassign *bool
	Gocyclo     *bool

	project *Project
}

func (c *StepCheck) Init(project *Project) error {
	c.project = project
	return nil
}

func (c *StepCheck) Name() string {
	return "check"
}

func (c *StepCheck) GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,
		Use:           "check",
		Short:         "check code quality",
		RunE: func(cmd *cobra.Command, args []string) error {
			logs.Info("Checking")
			// golint
			if err := ensureTool("golint", "golang.org/x/lint/golint"); err != nil {
				return err
			}
			logs.Info("Running lint")
			if err := ExecShell("./dist-tools/golint $(go list ./... | grep -v '/vendor/') | grep -v 'should have comment or be unexported'"); err != nil {
				return errs.WithE(err, "misspell failed")
			}

			// vet
			logs.Info("Running vet")
			if err := Exec("go", "vet"); err != nil {
				return errs.WithE(err, "vet failed")
			}

			// misspell
			if err := ensureTool("misspell", "github.com/client9/misspell/cmd/misspell"); err != nil {
				return err
			}
			logs.Info("Running misspell")
			if err := ExecShell("./dist-tools/misspell -source=text $(go list ./... | grep -v '/vendor/')"); err != nil {
				return errs.WithE(err, "misspell failed")
			}

			// ineffassign
			if err := ensureTool("ineffassign", "github.com/gordonklaus/ineffassign"); err != nil {
				return err
			}
			logs.Info("Running ineffassign")
			if err := ExecShell("./dist-tools/ineffassign -n $(find . -name '*.go' ! -path './vendor/*')"); err != nil {
				return errs.WithE(err, "ineffassign failed")
			}

			// gocyclo
			if err := ensureTool("gocyclo", "github.com/fzipp/gocyclo"); err != nil {
				return err
			}
			logs.Info("Running gocyclo")
			if err := ExecShell("./dist-tools/gocyclo -over 15 $(find . -name '*.go' ! -path './vendor/*')"); err != nil {
				return errs.WithE(err, "gocyclo failed")
			}

			return c.project.processArgs(args)
		},
	}
	return cmd
}

func ensureTool(tool string, toolPackage string) error {
	if _, err := os.Stat("dist-tools/" + tool); err != nil {
		logs.WithEF(err, data.WithField("tool", tool)).Warn("Building tool")
		if err := os.MkdirAll("./dist-tools", 0755); err != nil {
			return errs.WithE(err, "Failed to create dist-tools directory")
		}

		return Exec("go", "build", "-mod", "vendor", "-o", "./dist-tools/"+tool, toolPackage)
	}

	return nil
}

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
	if c.Lint == nil {
		c.Lint = True
	}
	if c.Vet == nil {
		c.Vet = True
	}
	if c.Misspell == nil {
		c.Misspell = True
	}
	if c.Ineffassign == nil {
		c.Ineffassign = True
	}
	if c.Gocyclo == nil {
		c.Gocyclo = True
	}

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
			if err := CommandDurationWrapper(cmd, func() error {
				ColorPrintln("Checking", HGreen)
				// golint
				if *c.Lint {
					if err := EnsureTool("golint", "golang.org/x/lint/golint"); err != nil {
						return err
					}
					ColorPrintln("lint", Magenta)
					if err := ExecShell("./dist-tools/golint $(go list ./... | grep -v '/vendor/') | grep -v 'should have comment or be unexported' || true"); err != nil {
						return errs.WithE(err, "lint failed")
					}
				}

				// vet
				if *c.Vet {
					ColorPrintln("vet", Magenta)
					if err := Exec("go", "vet", "./..."); err != nil {
						//return errs.WithE(err, "vet failed")
					}
				}

				// misspell
				if *c.Misspell {
					if err := EnsureTool("misspell", "github.com/client9/misspell/cmd/misspell"); err != nil {
						return err
					}
					ColorPrintln("misspell", Magenta)
					if err := ExecShell("./dist-tools/misspell -source=text $(go list ./... | grep -v '/vendor/') || true"); err != nil {
						return errs.WithE(err, "misspell failed")
					}
				}

				// ineffassign
				if *c.Ineffassign {
					if err := EnsureTool("ineffassign", "github.com/gordonklaus/ineffassign"); err != nil {
						return err
					}
					ColorPrintln("ineffassign", Magenta)
					if err := ExecShell("./dist-tools/ineffassign -n $(find . -name '*.go' ! -path './vendor/*') || true"); err != nil {
						return errs.WithE(err, "ineffassign failed")
					}
				}

				// gocyclo
				if *c.Gocyclo {
					if err := EnsureTool("gocyclo", "github.com/fzipp/gocyclo/cmd/gocyclo"); err != nil {
						return err
					}
					ColorPrintln("gocyclo", Magenta)
					if err := ExecShell("./dist-tools/gocyclo -over 15 $(find . -name '*.go' ! -path './vendor/*') || true"); err != nil {
						return errs.WithE(err, "gocyclo failed")
					}
				}
				return nil
			}); err != nil {
				return err
			}
			return c.project.processArgs(args)
		},
	}

	RegisterLogLevelParser(cmd)

	return cmd
}

func EnsureTool(tool string, toolPackage string) error {
	if _, err := os.Stat("dist-tools/" + tool); err != nil {
		logs.WithEF(err, data.WithField("tool", tool)).Warn("Building tool")
		if err := os.MkdirAll("./dist-tools", 0755); err != nil {
			return errs.WithE(err, "Failed to create dist-tools directory")
		}

		args := []string{"build", "-o", "./dist-tools/" + tool}
		if _, err := os.Stat("dist-tools/" + tool); err == nil {
			args = append(args, "-mod", "vendor")
		}
		args = append(args, toolPackage)

		return Exec("go", args...)
	}

	return nil
}

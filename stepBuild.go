package gomake

import (
	"fmt"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/hard-disk-manager/pkg/runner"
	"github.com/spf13/cobra"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

type StepBuild struct {
	BinaryName string
	OsArch     string
	UseVendor  *bool
	Version    string
	Package    string
	Upx        *bool

	project     *Project
	packageName string
}

func (c *StepBuild) Name() string {
	return "build"
}

func (c *StepBuild) Init(project *Project) error {
	c.project = project
	if c.BinaryName == "" {
		wd, err := os.Getwd()
		if err != nil {
			return errs.WithE(err, "Failed to get working directory to build")
		}
		c.BinaryName = path.Base(wd)
	}

	if len(c.OsArch) == 0 {
		c.OsArch = runtime.GOOS + "-" + runtime.GOARCH
	}

	if c.Package == "" {
		c.Package = "./"
	}

	if c.Upx == nil {
		c.Upx = False
	}

	if c.UseVendor == nil {
		c.UseVendor = False
	}

	if c.Version == "" {
		githash, err := runner.Local.ExecGetStdout("git", "rev-parse", "--short", "HEAD")
		if err != nil {
			return errs.WithE(err, "Failed to get git commit hash")
		}
		now := time.Now()
		c.Version = fmt.Sprintf("%s.%s.%s-%s",
			"1",
			now.Format("20060102"),
			strings.TrimLeft(now.Format("150405"), "0"),
			githash)
	}

	return nil
}

func (c *StepBuild) GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,
		Use:           "build",
		Short:         "build program",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := CommandDurationWrapper(cmd, func() error {
				ColorPrintln("Building", HGreen)
				distBindataPath := "dist/bindata"
				if err := os.MkdirAll(distBindataPath, 0755); err != nil {
					return errs.WithEF(err, data.WithField("path", distBindataPath), "Failed to create bindata dist directory")
				}

				ColorPrintln("fmt", Magenta)
				if err := Exec("go", "fmt"); err != nil {
					return err
				}

				ColorPrintln("fix", Magenta)
				if err := Exec("go", "fix"); err != nil {
					return err
				}

				ColorPrintln(c.OsArch, Magenta)
				buildArgs := []string{"build"}
				if *c.UseVendor {
					buildArgs = append(buildArgs, "-mod", "vendor")
				}
				buildArgs = append(buildArgs, "-ldflags", "-s -w -X main.Version="+c.Version)

				packageName, err := ExecGetStdout("go", "list", "-f", "{{.Name}}", c.Package)
				if err != nil {
					return errs.WithE(err, "Failed to get package name")
				}
				if packageName == "main" {
					buildArgs = append(buildArgs, "-o", "dist/"+c.BinaryName+"-"+c.OsArch+"/"+c.BinaryName)
				}

				if c.Package != "" {
					buildArgs = append(buildArgs, c.Package)
				}

				if err := Exec("go", buildArgs...); err != nil {
					return errs.WithE(err, "go build failed")
				}

				if *c.Upx && packageName != "main" {
					return errs.With("Cannot upx a library package")
				}
				if *c.Upx {
					if std, err := ExecGetStd("which", "upx"); err != nil {
						return errs.WithEF(err, data.WithField("std", std), "upx binary not in path")
					}

					if err := Exec("upx", "--ultra-brute", "dist/"+c.BinaryName+"-"+c.OsArch+"/"+c.BinaryName); err != nil {
						return errs.WithE(err, "upx failed")
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

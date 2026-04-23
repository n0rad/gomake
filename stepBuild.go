package gomake

import (
	"os"
	"runtime"
	"strings"

	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
)

type Program struct {
	BinaryName string
	OsArch     string
	Package    string
	Cgo        *bool

	version string
}

func (c *Program) Init(project *Project) error {
	if c.BinaryName == "" {
		c.BinaryName = project.name
	}

	if len(c.OsArch) == 0 {
		c.OsArch = runtime.GOOS + "-" + runtime.GOARCH
	}

	if c.Package == "" {
		c.Package = "./"
	}

	return nil
}

type StepBuild struct {
	Programs     []Program
	Version      string
	UseVendor    *bool
	Upx          *bool
	Fix          *bool
	Fmt          *bool
	PreBuildHook func(StepBuild) error // prepare bindata files

	project *Project
}

func (c *StepBuild) Name() string {
	return "build"
}

func (c *StepBuild) Init(project *Project) error {
	c.project = project

	if len(c.Programs) == 0 {
		c.Programs = append(c.Programs, Program{})
	}

	if c.Upx == nil {
		b := false
		c.Upx = &b
	}

	if c.Fix == nil {
		b := true
		c.Fix = &b
	}

	if c.Fmt == nil {
		b := true
		c.Fmt = &b
	}

	if c.UseVendor == nil {
		b := false
		c.UseVendor = &b
	}

	for i := range c.Programs {
		c.Programs[i].version = c.Version
		if c.Programs[i].Cgo == nil {
			b := false
			c.Programs[i].Cgo = &b
		}
	}

	for i := range c.Programs {
		if err := c.Programs[i].Init(c.project); err != nil {
			return errs.WithE(err, "Failed to init a program")
		}
	}

	return nil
}

func (c *StepBuild) GetCommand() *cobra.Command {
	PrepareOnly := false
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

				if c.PreBuildHook != nil {
					if err := c.PreBuildHook(*c); err != nil {
						return errs.WithE(err, "Pre build hook failed")
					}
				}

				empty, _ := IsDirectoryEmpty("dist/bindata")
				if !empty {
					if err := EnsureTool("go-bindata", "github.com/go-bindata/go-bindata/go-bindata"); err != nil {
						return err
					}
					if err := Exec("./dist-tools/go-bindata", "-nomemcopy", "-pkg", "dist", "-prefix", "dist/bindata", "-o", "dist/bindata.go", "dist/bindata/..."); err != nil {
						return errs.WithE(err, "go-bindata failed")
					}
				}

				ColorPrintln("generate", Magenta)
				if err := Exec("go", "generate", "./..."); err != nil {
					return err
				}

				if *c.Fmt {
					ColorPrintln("fmt", Magenta)
					if err := Exec("go", "fmt", "./..."); err != nil {
						return err
					}
				}

				if *c.Fix {
					ColorPrintln("fix", Magenta)
					if err := Exec("go", "fix"); err != nil {
						return err
					}
				}

				if PrepareOnly {
					return nil
				}

				if c.Version == "" {
					version, err := c.project.versionFunc()
					if err != nil {
						return errs.WithE(err, "Failed to generate version")
					}
					c.Version = version
				}

				for _, program := range c.Programs {
					fields := data.WithField("package", program.Package)

					ColorPrintln(program.BinaryName+" : "+program.OsArch, Magenta)
					osArchSplit := strings.Split(program.OsArch, "-")
					buildArgs := []string{"GOOS=" + osArchSplit[0], "GOARCH=" + osArchSplit[1]}
					if !*program.Cgo {
						buildArgs = append(buildArgs, "CGO_ENABLED=0")
					}

					buildArgs = append(buildArgs, "go", "build")
					if *c.UseVendor {
						buildArgs = append(buildArgs, "-mod", "vendor")
					}
					buildArgs = append(buildArgs, "-ldflags", "'-s -w -X main.Version="+c.Version+"'")

					packageName, err := ExecGetStdout("go", "list", "-f", "{{.Name}}", program.Package)
					if err != nil {
						return errs.WithEF(err, fields, "Failed to get package name")
					}
					if packageName == "main" {
						if strings.HasPrefix(program.OsArch, "windows") {
							buildArgs = append(buildArgs, "-o", "dist/"+c.project.name+"-"+program.OsArch+"/"+program.BinaryName+".exe")
						} else {
							buildArgs = append(buildArgs, "-o", "dist/"+c.project.name+"-"+program.OsArch+"/"+program.BinaryName)
						}
					}

					if program.Package != "" {
						buildArgs = append(buildArgs, program.Package)
					}

					if err := ExecShell(strings.Join(buildArgs, " ")); err != nil {
						return errs.WithEF(err, fields, "go build failed")
					}

					if *c.Upx && packageName != "main" {
						return errs.WithF(fields, "Cannot upx a library package")
					}
					if *c.Upx {
						if strings.HasPrefix(program.OsArch, "darwin") {
							logs.WithField("osArch", program.OsArch).Info("Skipping upx: not supported on darwin")
						} else {
							if std, err := ExecGetStd("which", "upx"); err != nil {
								return errs.WithEF(err, fields.WithField("std", std), "upx binary not in path")
							}

							if err := Exec("upx", "--ultra-brute", "dist/"+c.project.name+"-"+program.OsArch+"/"+program.BinaryName); err != nil {
								return errs.WithEF(err, fields, "upx failed")
							}
						}
					}

				}

				return nil
			}); err != nil {
				return err
			}
			return c.project.processArgs(args)
		},
	}

	cmd.Flags().BoolVarP(&PrepareOnly, "prepare-only", "p", false, "Only prepare the build, do not build binaries")
	cmd.Flags().BoolVar(c.Fix, "fix", *c.Fix, "Run go fix before building")
	cmd.Flags().BoolVar(c.Fmt, "fmt", *c.Fmt, "Run go fmt before building")
	cmd.Flags().StringVarP(&c.Version, "version", "v", c.Version, "Version to build")

	RegisterLogLevelParser(cmd)

	return cmd
}

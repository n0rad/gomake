package gomake

import (
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/spf13/cobra"
	"os"
	"runtime"
	"strings"
)

type Program struct {
	BinaryName string
	OsArch     string
	Package    string

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
		c.Upx = False
	}

	if c.UseVendor == nil {
		c.UseVendor = False
	}

	if c.Version == "" {
		v, err := GeneratedVersion()
		if err != nil {
			return errs.WithE(err, "Failed to generate version")
		}
		c.Version = v
	}

	for i := range c.Programs {
		c.Programs[i].version = c.Version
	}

	for i := range c.Programs {
		if err := c.Programs[i].Init(c.project); err != nil {
			return errs.WithE(err, "Failed to init a program")
		}
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

				if c.PreBuildHook != nil {
					if err := c.PreBuildHook(*c); err != nil {
						return errs.WithE(err, "Pre build hook failed")
					}
				}

				empty, _ := IsDirectoryEmpty("dist/bindata")
				if !empty {
					if err := ensureTool("go-bindata", "github.com/go-bindata/go-bindata/go-bindata"); err != nil {
						return err
					}
					if err := Exec("./dist-tools/go-bindata", "-nomemcopy", "-pkg", "dist", "-prefix", "dist/bindata", "-o", "dist/bindata.go", "dist/bindata/..."); err != nil {
						return errs.WithE(err, "go-bindata failed")
					}
				}

				ColorPrintln("fmt", Magenta)
				if err := Exec("go", "fmt"); err != nil {
					return err
				}

				ColorPrintln("fix", Magenta)
				if err := Exec("go", "fix"); err != nil {
					return err
				}

				for _, program := range c.Programs {
					fields := data.WithField("package", program.Package)

					ColorPrintln(program.BinaryName+" : "+program.OsArch, Magenta)
					osArchSplit := strings.Split(program.OsArch, "-")
					buildArgs := []string{"GOOS=" + osArchSplit[0], "GOARCH=" + osArchSplit[1], "go", "build"}
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
						return errs.WithEF(err, fields,"go build failed")
					}

					if *c.Upx && packageName != "main" {
						return errs.WithF(fields, "Cannot upx a library package")
					}
					if *c.Upx {
						if std, err := ExecGetStd("which", "upx"); err != nil {
							return errs.WithEF(err, fields.WithField("std", std), "upx binary not in path")
						}

						if err := Exec("upx", "--ultra-brute", "dist/"+program.BinaryName+"-"+program.OsArch+"/"+program.BinaryName); err != nil {
							return errs.WithEF(err, fields, "upx failed")
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

	RegisterLogLevelParser(cmd)

	return cmd
}

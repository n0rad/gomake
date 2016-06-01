package src

import (
	"os"
	"runtime"
	"github.com/n0rad/go-erlog/data"
	"github.com/blablacar/dgr/bin-dgr/common"
	"path/filepath"
	"github.com/n0rad/go-erlog/logs"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/gomake/dist"
	"io/ioutil"
)

type Project struct {
	workPath string
	config *Config
	fields data.Fields
}

func newProject(workPath string) (*Project, error) {
	config := newConfig()

	if config.name == "" {
		abs, err := filepath.Abs(workPath)
		if err != nil {
			return nil, errs.WithF(data.WithField("path", workPath), "Failed to get absolute path of workPath")
		}
		config.name = filepath.Base(abs)
	}

	return &Project{
		workPath: workPath,
		config: config,
		fields: data.WithField("name", config.name),
	}, nil
}

func (p *Project) Install() error {
	logs.WithF(p.fields).Info("Installing app to $GOPATH/bin")
	return common.CopyFile(p.buildFullname() + "/" + p.config.name, os.Getenv("GOPATH") + "/bin" + "/" + p.config.name)
}

func (p *Project) Clean() error {
	return os.RemoveAll(workPath + p.config.targetDirectory)
}

func (p *Project) internalCommand(command string, args []string) error {
	os.Setenv("app", p.config.name)
	os.Setenv("github_repo", p.config.repository)
	os.Chdir(p.workPath)
	for _, asset := range []string{"build","release","test","quality", "clean"} {
		content, err := dist.Asset("internal-scripts/command-"+asset+".sh")
		if err != nil {
			return errs.WithE(err, "cannot found internal "+asset+" script. This is a bug")
		}
		if err := ioutil.WriteFile("/tmp/command-"+asset+".sh", content, 0777); err != nil {
			return errs.WithEF(err, p.fields, "failed to write command-"+asset+"")
		}
	}
	return common.ExecCmd("/tmp/command-"+command+".sh", args...)
}

///////////////////////

func (p *Project) buildFullname() string {
	return p.config.name + "-" + p.config.version + "-" + runtime.GOOS + "-" + runtime.GOARCH
}



package src

import (
	"os"
	"runtime"
	"github.com/n0rad/go-erlog/data"
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/logs"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/gomake/dist"
	"io/ioutil"
	"strings"
)

type Project struct {
	workPath string
	config *Config
	fields data.Fields
}

func newProject(workPath string) (*Project, error) {
	config, err := newConfig(workPath)
	if err != nil {
		return nil, err
	}

	return &Project{
		workPath: workPath,
		config: config,
		fields: data.WithField("name", config.Name),
	}, nil
}

func (p *Project) Install() error {
	logs.WithF(p.fields).Info("Installing app to $GOPATH/bin")
	return common.CopyFile(p.buildFullname() + "/" + p.config.Name, os.Getenv("GOPATH") + "/bin" + "/" + p.config.Name)
}

func (p *Project) Clean() error {
	return os.RemoveAll(workPath + p.config.TargetDirectory)
}

func (p *Project) internalCommand(command string, args []string) error {
	os.Setenv("app", p.config.Name)
	os.Setenv("github_repo", strings.Replace(p.config.Repository, "github.com/", "", 1))
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
	return p.config.Name + "-" + p.config.Version + "-" + runtime.GOOS + "-" + runtime.GOARCH
}



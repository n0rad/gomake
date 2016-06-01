package src

import (
	"io/ioutil"
	"github.com/n0rad/go-erlog/logs"
	"github.com/n0rad/go-erlog/data"
	"gopkg.in/yaml.v2"
	"github.com/n0rad/go-erlog/errs"
	"path/filepath"
)

type Config struct {
	Name            string
	Repository      string
	Version         string
	TargetDirectory string
}

const pathGomakeYml = "/gomake.yml"

func newConfig(workPath string) (*Config, error) {
	errFields := data.WithField("file", workPath + pathGomakeYml)
	config := Config{
		TargetDirectory: "/dist",
	}

	if source, err := ioutil.ReadFile(workPath + pathGomakeYml); err == nil {
		err = yaml.Unmarshal([]byte(source), &config)
		if err != nil {
			return nil, errs.WithEF(err, errFields, "Failed to process configuration file")
		}
	} else {
		logs.WithF(errFields).Debug("No configuration file found")
	}

	if config.Name == "" {
		abs, err := filepath.Abs(workPath)
		if err != nil {
			return nil, errs.WithF(data.WithField("path", workPath), "Failed to get absolute path of workPath")
		}
		config.Name = filepath.Base(abs)
	}

	logs.WithF(data.WithField("conf", config)).Trace("Configuration")

	return &config, nil
}

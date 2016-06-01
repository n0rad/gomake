package main

import "os"

type Project struct {
	config *Config
}

func newProject(workPath string) *Project {
	return &Project{
		config: newConfig(),
	}
}

func (p *Project) clean() {
	os.RemoveAll(workPath + p.config.targetDirectory)
}

func (p *Project) build() {
	//
}

package executor

import "os/exec"

//Executor :
type Executor interface {
	/**
	Build a command
	*/
	BuildCmd(string) (*exec.Cmd, error)

	/**
	Handles exit code and rewrites exit code
	*/
	HandleExit(exitCode int, pid int) int
}

//TemplateInput :
type TemplateInput struct {
	Baseimage            string
	HasNodeJSApplication bool
	NginxOverrides       map[string]string
	ConfigurableProxy    bool
	Static               string
	SPA                  bool
	ExtraStaticHeaders   map[string]string
	Path                 string
	Env                  map[string]string
}

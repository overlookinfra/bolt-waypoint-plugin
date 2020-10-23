package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/puppetlabs/bolt-waypoint-plugin/builder"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Deploy struct {
	config DeployConfig
}

type DeployConfig struct {
	Plan    string   `hcl:"plan"`
	Targets []string `hcl:"targets,optional"`
	Project string   `hcl:"project,optional"`
	Flags   []string `hcl:"flags,optional"`
}

// Implement Configurable
func (d *Deploy) Config() (interface{}, error) {
	return &d.config, nil
}

// Implement ConfigurableNotify
func (d *Deploy) ConfigSet(config interface{}) error {
	_, ok := config.(*DeployConfig)
	if !ok {
		// The Waypoint SDK should ensure this never gets hit
		// The HCL type should assert this is a string
		return fmt.Errorf("Expected plan as parameter")
	}

	return nil
}

// Implement Builder
func (d *Deploy) DeployFunc() interface{} {
	// return a function which will be called by Waypoint
	return d.deploy
}

func runCommand(command string) ([]byte, error) {
	var cmdargs []string

	if runtime.GOOS == "windows" {
		cmdargs = []string{"cmd", "/C"}
	} else {
		cmdargs = []string{"/bin/sh", "-c"}
	}
	cmdargs = append(cmdargs, command)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdargs[0], cmdargs[1:]...)
	return cmd.Output()
}

func (d *Deploy) deploy(
	ctx context.Context,
	log hclog.Logger,
	ui terminal.UI,
	resultset *builder.ResultSet) (*ResultSet, error) {
	u := ui.Status()
	defer u.Close()
	u.Update(fmt.Sprintf("Deploying application: Running Bolt plan %s", d.config.Plan))

	cmdargs := []string{"bolt", "plan", "run", d.config.Plan}

	if d.config.Project != "" {
		cmdargs = append(cmdargs, "--project", d.config.Project)
	}

	if d.config.Targets != nil {
		cmdargs = append(cmdargs, "-t", strings.Join(d.config.Targets, ","))
	}

	if d.config.Flags != nil {
		cmdargs = append(cmdargs, d.config.Flags...)
	}
	// This can be overidden by prior flags
	cmdargs = append(cmdargs, "--no-host-key-check", "--format", "json")

	out, err := runCommand(strings.Join(cmdargs, " "))
	result := new(ResultSet)

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err = json.Unmarshal(out, result); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return result, nil
}

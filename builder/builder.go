package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

type Builder struct {
	config BuildConfig
}

type BuildConfig struct {
	Plan    string   `hcl:"plan"`
	Targets []string `hcl:"targets,optional"`
	Project string   `hcl:"project,optional"`
	Flags   []string `hcl:"flags,optional"`
}

// Implement Configurable
func (b *Builder) Config() (interface{}, error) {
	return &b.config, nil
}

// Implement ConfigurableNotify
func (b *Builder) ConfigSet(config interface{}) error {
	_, ok := config.(*BuildConfig)
	if !ok {
		// The Waypoint SDK should ensure this never gets hit
		// The HCL type should assert this is a string
		return fmt.Errorf("Expected plan as parameter")
	}

	return nil
}

// Implement Builder
func (b *Builder) BuildFunc() interface{} {
	return b.build
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

func (b *Builder) build(
	ctx context.Context,
	ui terminal.UI,
) (*ResultSet, error) {
	u := ui.Status()
	defer u.Close()
	u.Update(fmt.Sprintf("Building application: Running Bolt plan %s", b.config.Plan))

	cmdargs := []string{"bolt", "plan", "run", b.config.Plan}

	if b.config.Project != "" {
		cmdargs = append(cmdargs, "--project", b.config.Project)
	}

	if b.config.Targets != nil {
		cmdargs = append(cmdargs, "-t", strings.Join(b.config.Targets, ","))
	}

	if b.config.Flags != nil {
		cmdargs = append(cmdargs, b.config.Flags...)
	}
	// This can be overidden by prior flags
	cmdargs = append(cmdargs, "--no-host-key-check", "--format", "json")

	out, err := runCommand(strings.Join(cmdargs, " "))
	result := new(ResultSet)

	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(out, result); err != nil {
		return nil, err
	}

	return result, nil
}

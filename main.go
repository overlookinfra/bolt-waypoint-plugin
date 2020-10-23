package main

import (
  "github.com/puppetlabs/bolt-waypoint-plugin/builder"
  "github.com/puppetlabs/bolt-waypoint-plugin/platform"
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
)


func main() {
	// sdk.Main allows you to register the components which should
	// be included in your plugin
	// Main sets up all the go-plugin requirements

	sdk.Main(sdk.WithComponents(
    &builder.Builder{},
    &platform.Deploy{},
	))
}

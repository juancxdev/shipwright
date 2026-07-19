package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"
)

var Version = "dev"
var VersionName = "Phase 10 — Distribution & Installer"

func VersionCommand(args []string) {
	if len(args) > 0 && args[0] == "--json" {
		payload := map[string]string{
			"version": Version,
			"name":    VersionName,
			"os":      runtime.GOOS,
			"arch":    runtime.GOARCH,
		}
		data, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Println(string(data))
		return
	}
	fmt.Printf("Shipwright %s (%s)\n", Version, VersionName)
}

package main

import (
	"fmt"
	"os"

	"shipwright/cmd"
)

func main() {
	if len(os.Args) < 2 {
		cmd.PrintUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "init":
		cmd.Init(args)
	case "start":
		cmd.Start(args)
	case "status":
		cmd.Status(args)
	case "next":
		cmd.Next(args)
	case "run":
		cmd.Run(args)
	case "approve":
		cmd.Approve(args)
	case "request-change":
		cmd.RequestChange(args)
	case "scaffold":
		cmd.Scaffold(args)
	case "generate":
		cmd.Generate(args)
	case "agents":
		cmd.Agents(args)
	case "contract":
		cmd.Contract(args)
	case "memory":
		cmd.Memory(args)
	case "integrations":
		cmd.Integrations(args)
	case "config":
		cmd.Config(args)
	case "executor":
		cmd.Executor(args)
	case "doctor":
		cmd.Doctor(args)
	case "design":
		cmd.Design(args)
	case "review":
		cmd.Review(args)
	case "skills":
		cmd.Skills(args)
	case "tdd":
		cmd.TDD(args)
	case "help", "-h", "--help":
		cmd.PrintUsage()
	case "version", "-v", "--version":
		cmd.VersionCommand(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", command)
		cmd.PrintUsage()
		os.Exit(1)
	}
}

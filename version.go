package main

import "fmt"

var (
	version 	 	   = "dev"
	commit  	 	   = "none"
	repoUrl	   = "unknown"
	buildDate    	   = "unknown"
	buildFlavor  	   = "unknown"
	builtBy    	 	   = "unknown"
	builtWithGoVersion = "unknown"
)

type VersionCommand struct {}

var versionCommand VersionCommand

func (v VersionCommand) Execute(args []string) error {
	fmt.Printf("supervisor %v (%v#%v)\n\n", version, repoUrl, commit);
	fmt.Printf("build:\n  flavor=%v\n  date=%v\n  by=%v\n  using=%v\n", buildFlavor, buildDate, builtBy, builtWithGoVersion);
	
	return nil
}

func init() {
	parser.AddCommand("version",
		"show the version of supervisor",
		"display the supervisor version",
		&versionCommand)
}

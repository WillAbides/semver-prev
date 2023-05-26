package main

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/alecthomas/kong"
)

var version = "unknown"

type cmd struct {
	Dir        string           `kong:"short=C,help='The directory containing the git repository.'"`
	Ref        string           `kong:"short=r,default=HEAD,help='Ref to the commit to start at.'"`
	Prefix     []string         `kong:"default='',help='Prefix the tag must start with. When specified multiple times, the tag must start with any of the prefixes.'"`
	Constraint string           `kong:"help='A semver constraint to filter tags by.'"`
	Fallback   *string          `kong:"help='If no version is found, output this instead of erroring.'"`
	Version    kong.VersionFlag `kong:"help='Print version and exit.'"`
}

func main() {
	ctx := context.Background()
	var cli cmd
	k := kong.Parse(&cli, kong.Vars{"version": version})
	opts := prevVersionOptions{
		head:     cli.Ref,
		repoDir:  cli.Dir,
		prefixes: cli.Prefix,
	}
	var err error
	if cli.Constraint != "" {
		opts.constraints, err = semver.NewConstraint(cli.Constraint)
		k.FatalIfErrorf(err)
	}
	ver, err := prevVersion(ctx, &opts)
	k.FatalIfErrorf(err)
	if ver != "" {
		fmt.Println(ver)
		return
	}
	if cli.Fallback != nil {
		fmt.Println(*cli.Fallback)
		return
	}
	if ver == "" {
		k.Fatalf("No previous version found.")
	}
	fmt.Println(ver)
}

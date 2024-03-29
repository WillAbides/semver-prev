package main

import (
	"bufio"
	"context"
	"os/exec"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type prevVersionOptions struct {
	head        string
	repoDir     string
	prefixes    []string
	constraints *semver.Constraints
}

func prevVersion(ctx context.Context, options *prevVersionOptions) (string, error) {
	if options == nil {
		options = &prevVersionOptions{}
	}
	head := options.head
	if head == "" {
		head = "HEAD"
	}
	prefixes := options.prefixes
	if len(prefixes) == 0 {
		prefixes = []string{""}
	}
	cmdLine := []string{"git", "rev-list", "--pretty=%D", head}
	type prefixedVersion struct {
		prefix string
		ver    *semver.Version
	}
	var versions []prefixedVersion
	done := false
	err := runCommandHandleLines(ctx, options.repoDir, cmdLine, func(line string, cancel context.CancelFunc) {
		if done {
			return
		}
		refs := strings.Split(line, ", ")
		for _, r := range refs {
			var ok bool
			r, ok = strings.CutPrefix(r, "tag: ")
			if !ok {
				continue
			}
			for _, prefix := range options.prefixes {
				r, ok = strings.CutPrefix(r, prefix)
				if !ok {
					continue
				}
				ver, err := semver.StrictNewVersion(r)
				if err != nil {
					continue
				}
				if options.constraints != nil && !options.constraints.Check(ver) {
					continue
				}
				versions = append(versions, prefixedVersion{prefix, ver})
			}
		}
		if len(versions) > 0 {
			cancel()
			done = true
		}
	})
	if err != nil {
		return "", err
	}
	// order first by version then by index of prefix in prefixes
	sort.Slice(versions, func(i, j int) bool {
		a, b := versions[i], versions[j]
		if !a.ver.Equal(b.ver) {
			return a.ver.GreaterThan(b.ver)
		}
		for _, prefix := range prefixes {
			if a.prefix == prefix {
				return b.prefix != prefix
			}
			if b.prefix == prefix {
				return false
			}
		}
		return false
	})
	if len(versions) == 0 {
		return "", nil
	}
	winner := versions[0]
	return winner.prefix + winner.ver.Original(), nil
}

func runCommandHandleLines(
	ctx context.Context,
	dir string,
	cmdLine []string,
	handleLine func(line string, cancel context.CancelFunc),
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	command := exec.CommandContext(ctx, cmdLine[0], cmdLine[1:]...)
	command.Dir = dir

	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	err = command.Start()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		handleLine(line, cancel)
	}
	err = command.Wait()
	if err == nil {
		return nil
	}
	if err != context.Canceled && err.Error() != "signal: killed" {
		return err
	}
	return nil
}

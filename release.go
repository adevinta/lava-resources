// Copyright 2023 Adevinta

/*
Release publishes a new GitHub release with a given set of files.

It expects an environment variable with the name GITHUB_REF_NAME. The
value of GITHUB_REF_NAME must be a git tag with the format
"dir/semver" (e.g. checktypes/v1.2.3).

For a given tag, it creates three releases:

  - dir/vMAJOR
  - dir/vMAJOR.MINOR
  - dir/vMAJOR.MINOR.PATCH

If the version in the tag name corresponds to a prerelease, only
vMAJOR.MINOR.PATCH-PRERELEASE is created.

The regular files in the directory specified in the tag are attached
to all the releases.

For instance, if the tag is checktypes/v1.2.3, the following releases
will be created:

  - checktypes/v1     (updated if it already exists)
  - checktypes/v1.2   (updated if it already exists)
  - checktypes/v1.2.3

The files under the directory "checktypes" are attached to all of
them.

This release schema allows users to pin versions depending on their
needs. In other words,

  - v1.2.3  :=  ==v1.2.3
  - v1.2    :=  >=v1.2.0, <v1.3.0
  - v1      :=  >=v1.0.0, <v2.0.0
  - v0.2.3  :=  ==v0.2.3
  - v0.2    :=  >=v0.2.0, <v0.3.0
  - v0      :=  >=v0.0.0, <v1.0.0
*/
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

func main() {
	log.SetFlags(0)

	refName := os.Getenv("GITHUB_REF_NAME")
	if refName == "" {
		log.Fatalf("error: missing env var GITHUB_REF_NAME")
	}

	dir, version, err := parseRef(refName)
	if err != nil {
		log.Fatalf("error: parse ref: %v", err)
	}

	files, err := readDir(dir)
	if err != nil {
		log.Fatalf("error: list files: %v", err)
	}

	hash, err := gitHash(refName)
	if err != nil {
		log.Fatalf("error: get hash: %v", err)
	}

	var releases []string

	// Do not create vMAJOR and vMAJOR.MINOR for pre-releases.
	if semver.Prerelease(version) == "" {
		releases = []string{semver.Major(version), semver.MajorMinor(version)}
	}

	releases = append(releases, version)

	for _, r := range releases {
		tag := dir + "/" + r

		// Do not update vMAJOR.MINOR.PATCH.
		update := tag != refName

		if err := ghRelease(tag, hash, update, files); err != nil {
			log.Fatalf("error: create GitHub release %q: %v", tag, err)
		}
	}
}

// parseRef parses a reference name with the format "dir/semver" (e.g.
// checktypes/v1.2.3). It returns the directory and version parts of
// the reference. An error is returned if the reference does not match
// the expected format or the version is not a valid semantic version.
func parseRef(ref string) (dir, version string, err error) {
	parts := strings.Split(ref, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid tag name %q", ref)
	}

	dir = parts[0]
	version = parts[1]

	if !semver.IsValid(version) {
		return "", "", fmt.Errorf("invalid version %q", version)
	}

	return dir, version, nil
}

// gitHash returns the hash corresponding to the specified reference
// using "git show-ref".
func gitHash(ref string) (string, error) {
	hash, err := cmdOutput("git", "show-ref", "--hash", ref)
	if err != nil {
		return "", fmt.Errorf("git show-ref: %w", err)
	}
	return hash, nil
}

// ghRelease creates a GitHub release using "gh release". If update is
// true, it first tries to delete any existing release with the same
// tag.
func ghRelease(tag, target string, update bool, files []string) error {
	if update {
		if _, err := cmdOutput("gh", "release", "delete", "--cleanup-tag", "--yes", tag); err != nil {
			log.Printf("warn: could not delete release %q", tag)
		}

		// BUG(rm): If a release is deleted and then created
		// again to update its reference, the new release is
		// created as draft. This happens because of a race
		// condition on the GitHub side.
		//
		// A 30s delay should mitigate the issue while it is
		// not fixed by GitHub.
		//
		// For more information, see
		// https://github.com/cli/cli/issues/8458
		time.Sleep(30 * time.Second)
	}

	args := []string{"release", "create", "--target", target, tag}
	args = append(args, files...)
	if _, err := cmdOutput("gh", args...); err != nil {
		return fmt.Errorf("gh release create: %w", err)
	}

	return nil
}

// readDir returns the list of files in the provided directory. The
// returned filenames are prefixed with dir.
func readDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			log.Printf("warn: skipping dir %q", entry.Name())
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}

	return files, nil
}

// cmdOutput runs the specified command and returns its standard
// output. The returned output is trimmed. In case of error, it
// returns stderr along with the error.
func cmdOutput(name string, arg ...string) (string, error) {
	stderr := &bytes.Buffer{}
	cmd := exec.Command(name, arg...)
	cmd.Stderr = stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("cmd output: %w: %#q", err, stderr)
	}
	return strings.TrimSpace(string(out)), nil
}

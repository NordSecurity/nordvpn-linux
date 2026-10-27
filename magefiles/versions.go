//go:build mage

package main

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"go.yaml.in/yaml/v3"
)

const versionsFile = "versions.yaml"

// markerPattern matches a marker comment holding the template of the line below it, e.g.:
//
//	# versions.yaml: FROM ubuntu:{ubuntu-tooling}
//	FROM ubuntu:24.04
//
// Parts of the pattern:
//   - ^ and $: match whole line
//   - ([[:blank:]]*): indentation (spaces or tabs), copied to the target line
//   - (#|//): comment prefix, '#' for Dockerfile, YAML and env files, '//' for Go and go.mod
//   - " versions\.yaml: ": the marker itself, with exactly one space after the comment prefix
//   - (.*): group 3, template of the managed line with {name} placeholders
const markerPattern = `^([[:blank:]]*)(#|//) versions\.yaml: (.*)$`

var markerRegexp = regexp.MustCompile(markerPattern)

// versionPlaceholder matches {name} in marker templates.
var versionPlaceholder = regexp.MustCompile(`\{[a-z][a-z0-9-]*\}`)

type Versions mg.Namespace

type fileUpdate struct {
	file       string
	lines      []string
	mismatches []versionMismatch
}

type versionMismatch struct {
	file  string
	line  int
	found string
	want  string
}

func (Versions) Sync() error {
	updates, err := syncVersions()
	if err != nil {
		return err
	}
	for _, update := range updates {
		if err := update.write(); err != nil {
			return err
		}
		for _, m := range update.mismatches {
			fmt.Println("updated", m)
		}
	}
	return tidyGoModules(updates)
}

func (Versions) Check() error {
	updates, err := syncVersions()
	if err != nil {
		return err
	}
	count := 0
	for _, update := range updates {
		for _, m := range update.mismatches {
			fmt.Println(m)
			count++
		}
	}
	if count == 0 {
		return nil
	}
	return fmt.Errorf("%d line(s) out of sync with %s, run `mage versions:sync`", count, versionsFile)
}

func syncVersions() ([]fileUpdate, error) {
	versions, err := loadVersions()
	if err != nil {
		return nil, err
	}
	files, err := findMarkerFiles()
	if err != nil {
		return nil, err
	}

	used := map[string]bool{}
	var updates []fileUpdate
	for _, file := range files {
		update, err := syncFile(file, versions, used)
		if err != nil {
			return nil, err
		}
		if len(update.mismatches) > 0 {
			updates = append(updates, update)
		}
	}
	if err := checkAllUsed(versions, used); err != nil {
		return nil, err
	}
	return updates, nil
}

func loadVersions() (map[string]string, error) {
	content, err := os.ReadFile(versionsFile)
	if err != nil {
		return nil, err
	}
	var config struct {
		Versions map[string]string `yaml:"versions"`
	}
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", versionsFile, err)
	}
	for name, version := range config.Versions {
		if version == "" {
			return nil, fmt.Errorf("%s: %s has no version", versionsFile, name)
		}
	}
	return config.Versions, nil
}

func findMarkerFiles() ([]string, error) {
	out, err := sh.Output("git", "grep", "-l", "-z", "-I", "--untracked", "-E", "-e", markerPattern)
	if sh.ExitStatus(err) == 1 {
		// Nothing found, reported later as unused versions.
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("searching for markers: %w", err)
	}
	return strings.Split(strings.TrimSuffix(out, "\x00"), "\x00"), nil
}

// XXX: Simplify or at least clean up this
func syncFile(file string, versions map[string]string, used map[string]bool) (fileUpdate, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return fileUpdate{}, err
	}
	update := fileUpdate{file: file, lines: strings.Split(string(content), "\n")}
	for i, line := range update.lines {
		marker := markerRegexp.FindStringSubmatch(line)
		if marker == nil {
			continue
		}
		if i+1 == len(update.lines) || markerRegexp.MatchString(update.lines[i+1]) {
			return fileUpdate{}, fmt.Errorf("%s:%d: marker must be followed by the line it manages", file, i+1)
		}

		indent, template := marker[1], marker[3]
		want, names, err := renderTemplate(template, versions)
		if err != nil {
			return fileUpdate{}, fmt.Errorf("%s:%d: %w", file, i+1, err)
		}
		for _, name := range names {
			used[name] = true
		}

		want = indent + want
		if found := update.lines[i+1]; found != want {
			update.mismatches = append(update.mismatches, versionMismatch{file: file, line: i + 2, found: found, want: want})
			update.lines[i+1] = want
		}
	}
	return update, nil
}

func renderTemplate(template string, versions map[string]string) (string, []string, error) {
	var names, unknown []string
	rendered := versionPlaceholder.ReplaceAllStringFunc(template, func(placeholder string) string {
		name := placeholder[1 : len(placeholder)-1]
		names = append(names, name)
		version, ok := versions[name]
		if !ok {
			unknown = append(unknown, name)
		}
		return version
	})
	if len(unknown) > 0 {
		return "", nil, fmt.Errorf("%s not defined in %s", strings.Join(unknown, ", "), versionsFile)
	}
	return rendered, names, nil
}

func checkAllUsed(versions map[string]string, used map[string]bool) error {
	var unused []string
	for _, name := range slices.Sorted(maps.Keys(versions)) {
		if !used[name] {
			unused = append(unused, name)
		}
	}
	if len(unused) > 0 {
		return fmt.Errorf("%s: %s not used by any marker", versionsFile, strings.Join(unused, ", "))
	}
	return nil
}

func (u fileUpdate) write() error {
	info, err := os.Stat(u.file)
	if err != nil {
		return err
	}
	return os.WriteFile(u.file, []byte(strings.Join(u.lines, "\n")), info.Mode().Perm())
}

func tidyGoModules(updates []fileUpdate) error {
	for _, update := range updates {
		if filepath.Base(update.file) != "go.mod" {
			continue
		}
		if err := sh.RunV(mg.GoCmd(), "-C", filepath.Dir(update.file), "mod", "tidy"); err != nil {
			return err
		}
	}
	return nil
}

func (m versionMismatch) String() string {
	return fmt.Sprintf("%s:%d\n  have: %s\n  want: %s", m.file, m.line, m.found, m.want)
}

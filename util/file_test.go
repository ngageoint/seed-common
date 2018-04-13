	package util

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func init() {
	InitPrinter(PrintErr)
}

func TestGetFullPath(t *testing.T) {
	curDir, _ := os.Getwd()
	cases := []struct {
		file     string
		dir      string
		ending   string
		errorMsg string
	}{
		{"file.go", ".", curDir + "/file.go", ""},
		{"seed-scale.zip", "../testdata/", "/testdata/seed-scale.zip", ""},
		{"/absolute/path", "", "/absolute/path", ""},
	}

	for _, c := range cases {
		full := GetFullPath(c.file, c.dir)
		if !strings.HasSuffix(full, c.ending) {
			t.Errorf("GetFullPath(%q, %q) == %v, expected string ending in %v", c.file, c.dir, full, c.ending)
		}
	}
}

func TestDockerfileBaseRegistry(t *testing.T) {
	cases := []struct {
		dir      string
		registry string
		errorMsg string
	}{
		{".", "", "no such file or directory"},
		{"../testdata/complete", "", ""},
		{"../testdata/remote-base-registry", "remoteRegistry:5000", ""},
	}

	for _, c := range cases {
		registry, err := DockerfileBaseRegistry(c.dir)
		if registry != c.registry {
			t.Errorf("DockerfileBaseRegistry(%q) == %v, expected %v", c.dir, registry, c.registry)
		}

		errMsg := ""
		if err != nil {
			errMsg = fmt.Sprintf("%s", err.Error())
		}
		if !strings.Contains(errMsg, c.errorMsg) {
			t.Errorf("DockerfileBaseRegistry(%q) == %v, expected %v", c.dir, errMsg, c.errorMsg)
		}
	}
}

func TestReadLinesFromFile(t *testing.T) {
	cases := []struct {
		dir       string
		lineCount int
		errorMsg  string
	}{
		{".", 0, "seed.manifest.json cannot be found"},
		{"../testdata/complete", 101, ""},
	}

	for _, c := range cases {
		file, err := SeedFileName(c.dir)
		num_lines := 0

		if err == nil {
			lines, _ := ReadLinesFromFile(file)
			num_lines = len(lines)
		}

		if num_lines != c.lineCount {
			t.Errorf("ReadLinesFromFile(%q) returned %d lines, expected %d", file, num_lines, c.lineCount)
		}

		errMsg := ""
		if err != nil {
			errMsg = fmt.Sprintf("%s", err.Error())
		}
		if !strings.Contains(errMsg, c.errorMsg) {
			t.Errorf("TestReadLinesFromFile(%q) == %v, expected %v", c.dir, errMsg, c.errorMsg)
		}
	}
}

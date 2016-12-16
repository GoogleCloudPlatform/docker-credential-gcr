// Copyright 2016 Google, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package util contains utilities which are shared between packages.
*/
package util

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// SdkConfigPath tries to return the directory where the gcloud config is
// located.
func SdkConfigPath() (string, error) {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "gcloud"), nil
	}
	homeDir := unixHomeDir()
	if homeDir == "" {
		return "", errors.New("unable to get current user home directory: os/user lookup failed; $HOME is empty")
	}
	return filepath.Join(homeDir, ".config", "gcloud"), nil
}

func unixHomeDir() string {
	usr, err := user.Current()
	if err == nil {
		return usr.HomeDir
	}
	return os.Getenv("HOME")
}

// DockerClientVersionStrings attempts to discover the version of the Docker client,
// returning the major, minor, and patch versions, or an error if unsuccessful.
func DockerClientVersionStrings() (string, string, string, error) {
	cmd := exec.Command("docker", "version", "--format", "'{{.Client.Version}}'")
	out, err := cmd.Output()
	if err != nil {
		return "", "", "", err
	}

	vstring := string(out)

	// Remove any leading/trailing '
	vstring = strings.Trim(vstring, "'")

	ver := strings.Split(vstring, ".")

	if len(ver) != 3 {
		return "", "", "", errors.New("version string not of the form '<major>.<minor>.<patch>': " + vstring)
	}
	return ver[0], ver[1], ver[2], nil
}

// DockerClientVersion attempts to discover the major and minor version
// numbers of the Docker client, returning <major number>, <minor number>,
// <patch number>, <patch suffix>, nil if successful, 0, 0, err otherwise.
// e.g.
// '1.12.0' => 1, 12, 0, "", nil
// '1.13.0-dev' => 1, 13, 0, "dev", nil
// '1.what.0' => 0, 0, 0, "", nil
func DockerClientVersion() (int, int, int, string, error) {
	majorstr, minorstr, patchstr, err := DockerClientVersionStrings()
	if err != nil {
		return 0, 0, 0, "", err
	}

	major, err := strconv.Atoi(majorstr)
	if err != nil {
		return 0, 0, 0, "", err
	}
	minor, err := strconv.Atoi(minorstr)
	if err != nil {
		return 0, 0, 0, "", err
	}

	patchSplit := strings.Split(patchstr, "-")
	patch, err := strconv.Atoi(patchSplit[0])
	if err != nil {
		return 0, 0, 0, "", err
	}

	return major, minor, patch, patchSplit[1], nil
}

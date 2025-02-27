// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023-Present The UDS Authors

// Package test contains e2e tests for UDS
package test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/defenseunicorns/zarf/src/pkg/utils/exec"
	"github.com/defenseunicorns/zarf/src/pkg/utils/helpers"
	"github.com/stretchr/testify/require"
)

// UDSE2ETest Struct holding common fields most of the tests will utilize.
type UDSE2ETest struct {
	UdsBinPath        string
	Arch              string
	ApplianceMode     bool
	ApplianceModeKeep bool
	RunClusterTests   bool
	CommandLog        []string
}

// GetCLIName looks at the OS and CPU architecture to determine which Zarf binary needs to be run.
func GetCLIName() string {
	var binaryName string
	if runtime.GOOS == "linux" {
		binaryName = "uds"
	} else if runtime.GOOS == "darwin" {
		if runtime.GOARCH == "arm64" {
			binaryName = "uds-mac-apple"
		} else {
			binaryName = "uds-mac-intel"
		}
	}
	return binaryName
}

var logRegex = regexp.MustCompile(`Saving log file to (?P<logFile>.*?\.log)`)

// SetupWithCluster performs actions for each test that requires a K8s cluster.
func (e2e *UDSE2ETest) SetupWithCluster(t *testing.T) {
	if !e2e.RunClusterTests {
		t.Skip("")
	}
	_ = exec.CmdWithPrint("sh", "-c", fmt.Sprintf("%s tools kubectl describe nodes | grep -A 99 Non-terminated", e2e.UdsBinPath))
}

// UDS executes a UDS command.
func (e2e *UDSE2ETest) UDS(args ...string) (string, string, error) {
	e2e.CommandLog = append(e2e.CommandLog, strings.Join(args, " "))
	return exec.CmdWithContext(context.TODO(), exec.PrintCfg(), e2e.UdsBinPath, args...)
}

// CleanFiles removes files and directories that have been created during the test.
func (e2e *UDSE2ETest) CleanFiles(files ...string) {
	for _, file := range files {
		_ = os.RemoveAll(file)
	}
}

// GetMismatchedArch determines what architecture our tests are running on,
// and returns the opposite architecture.
func (e2e *UDSE2ETest) GetMismatchedArch() string {
	switch e2e.Arch {
	case "arm64":
		return "amd64"
	default:
		return "arm64"
	}
}

// GetLogFileContents gets the log file contents from a given run's std error.
func (e2e *UDSE2ETest) GetLogFileContents(t *testing.T, stdErr string) string {
	get, err := helpers.MatchRegex(logRegex, stdErr)
	require.NoError(t, err)
	logFile := get("logFile")
	logContents, err := os.ReadFile(logFile)
	require.NoError(t, err)
	return string(logContents)
}

// SetupDockerRegistry uses the host machine's docker daemon to spin up a local registry for testing purposes.
func (e2e *UDSE2ETest) SetupDockerRegistry(t *testing.T, port int) {
	// spin up a local registry
	registryImage := "registry:2.8.2"
	err := exec.CmdWithPrint("docker", "run", "-d", "--restart=always", "-p", fmt.Sprintf("%d:5000", port), "--name", fmt.Sprintf("registry-%d", port), registryImage)
	require.NoError(t, err)
}

// TeardownRegistry removes the local registry.
func (e2e *UDSE2ETest) TeardownRegistry(t *testing.T, port int) {
	// remove the local registry
	err := exec.CmdWithPrint("docker", "rm", "-f", fmt.Sprintf("registry-%d", port))
	require.NoError(t, err)
}

// GetUdsVersion returns the current build version
func (e2e *UDSE2ETest) GetUdsVersion(t *testing.T) string {
	// Get the version of the CLI
	stdOut, stdErr, err := e2e.UDS("version")
	require.NoError(t, err, stdOut, stdErr)
	return strings.Trim(stdOut, "\n")
}

// DownloadZarfInitPkg downloads the zarf init pkg used for testing if it doesn't already exist (todo: makefile?)
func (e2e *UDSE2ETest) DownloadZarfInitPkg(t *testing.T, zarfVersion string) {
	filename := fmt.Sprintf("zarf-init-%s-%s.tar.zst", e2e.Arch, zarfVersion)
	zarfReleaseURL := fmt.Sprintf("https://github.com/defenseunicorns/zarf/releases/download/%s/%s", zarfVersion, filename)
	outputDir := "src/test/packages/zarf"

	// Check if the file already exists
	if _, err := os.Stat(outputDir + "/" + filename); err == nil {
		fmt.Println("Zarf init pkg already exists. Skipping download.")
		require.NoError(t, err)
		return
	}

	err := downloadFile(zarfReleaseURL, outputDir)
	require.NoError(t, err)
}

// CreateZarfPkg creates a Zarf in the given path (uses system Zarf binary) (todo: makefile?)
func (e2e *UDSE2ETest) CreateZarfPkg(t *testing.T, path string) {
	//  check if pkg already exists
	pattern := fmt.Sprintf("%s/*-%s-*.tar.zst", path, e2e.Arch)
	matches, err := filepath.Glob(pattern)
	require.NoError(t, err)
	if len(matches) > 0 {
		fmt.Println("Zarf pkg already exists, skipping create")
		return
	}
	cmd := "zarf"
	args := strings.Split(fmt.Sprintf("package create . --confirm"), " ")
	tmp := exec.PrintCfg()
	tmp.Dir = path
	_, _, err = exec.CmdWithContext(context.TODO(), tmp, cmd, args...)
	require.NoError(t, err)
}

func downloadFile(url string, outputDir string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	outputFileName := filepath.Base(url)
	outputFilePath := filepath.Join(outputDir, outputFileName)

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, response.Body)
	if err != nil {
		return err
	}

	return nil
}

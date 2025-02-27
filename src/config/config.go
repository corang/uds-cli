// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023-Present The UDS Authorspackage config

// Package config contains configuration strings for UDS-CLI
package config

import (
	"runtime"

	"github.com/corang/uds-cli/src/types"
	zarfConfig "github.com/defenseunicorns/zarf/src/config"
	zarfTypes "github.com/defenseunicorns/zarf/src/types"
)

const (
	// ZarfYAML is the string for zarf.yaml
	ZarfYAML = "zarf.yaml"

	// BlobsDir is the string for the blobs/sha256 dir in an OCI artifact
	BlobsDir = "blobs/sha256"

	// BundleYAML is the string for zarf.yaml
	BundleYAML = "uds-bundle.yaml"

	// BundlePrefix is the prefix for compiled uds bundles
	BundlePrefix = "uds-bundle-"

	// SBOMsTar is the sboms.tar file in a Zarf pkg
	SBOMsTar = "sboms.tar"

	// BundleSBOMTar is the name of the tarball containing the bundle's SBOM
	BundleSBOMTar = "bundle-sboms.tar"

	// BundleSBOM is the name of the untarred folder containing the bundle's SBOM
	BundleSBOM = "bundle-sboms"

	// BundleYAMLSignature is the name of the bundle's metadata signature file
	BundleYAMLSignature = "uds-bundle.yaml.sig"

	// PublicKeyFile is the name of the public key file
	PublicKeyFile = "public.key"
)

var (
	// CommonOptions tracks user-defined values that apply across commands.
	CommonOptions types.BundlerCommonOptions

	// CLIVersion track the version of the CLI
	CLIVersion = "unset"

	// CLIArch is the computer architecture of the device executing the CLI commands
	CLIArch string

	// SkipLogFile is a flag to skip logging to a file
	SkipLogFile bool
)

// GetArch returns the arch based on a priority list with options for overriding.
func GetArch(archs ...string) string {
	// List of architecture overrides.
	priority := append([]string{CLIArch}, archs...)

	// Find the first architecture that is specified.
	for _, arch := range priority {
		if arch != "" {
			return arch
		}
	}

	return runtime.GOARCH
}

var (
	// BundleAlwaysPull is a list of paths that will always be pulled from the remote repository.
	BundleAlwaysPull = []string{BundleYAML, BundleYAMLSignature}
)

// DefaultZarfInitOptions set these in the case of deploying a Zarf init pkg
// typically these are set as part of Zarf's Viper config, which we don't use in UDS
// could technically remove, but it doesn't hurt anything for now
var DefaultZarfInitOptions = zarfTypes.ZarfInitOptions{
	GitServer: zarfTypes.GitServerInfo{
		PushUsername: zarfConfig.ZarfGitPushUser,
	},
	RegistryInfo: zarfTypes.RegistryInfo{
		PushUsername: zarfConfig.ZarfRegistryPushUser,
	},
}

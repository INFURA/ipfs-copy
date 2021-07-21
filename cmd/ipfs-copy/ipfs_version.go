package main

import "github.com/coreos/go-semver/semver"

func hasShellStreamPinListSupport(version string) bool {
	sourceShellVersion := semver.New(version)
	// The `pin ls --stream` feature was added in https://github.com/ipfs/go-ipfs/blob/master/CHANGELOG.md#050-2020-04-28
	requiredEnumStreamVersion := semver.New("0.5.0")

	return requiredEnumStreamVersion.LessThan(*sourceShellVersion) || requiredEnumStreamVersion.Equal(*sourceShellVersion)
}

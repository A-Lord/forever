// Command gen_buffs_proto renders proto/buffs.proto from tools/database/buffmanifest.
//
// It imports nothing but the standard library and the manifest, so it runs while
// sim/core/proto and the generated buff files are stale: protoc needs buffs.proto
// before anything that reads the compiled protos can build. The spell data pass
// renders the same file and names it in -check when it is stale.
//
//	go run ./tools/gen_buffs_proto
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wowsims/forever/tools/database/buffmanifest"
)

const destPath = "proto/buffs.proto"

var out = flag.String("out", "", "output path; defaults to proto/buffs.proto under the repo root")

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "gen_buffs_proto: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	path := *out
	if path == "" {
		repoRoot, err := findRepoRoot()
		if err != nil {
			return err
		}
		path = filepath.Join(repoRoot, destPath)
	}

	rendered := buffmanifest.RenderProto(buffmanifest.Manifest)
	if err := os.WriteFile(path, rendered, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	fmt.Printf("gen_buffs_proto: wrote %s (%d fields, %d bytes)\n", path, len(buffmanifest.Manifest), len(rendered))
	return nil
}

// findRepoRoot walks up from the working directory looking for go.mod, so the tool works from
// anywhere (`go run ./tools/gen_buffs_proto` runs in the caller's directory).
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find go.mod above the working directory")
		}
		dir = parent
	}
}

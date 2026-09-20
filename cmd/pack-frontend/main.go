// Command pack-frontend packs a built frontend "dist" directory into the
// zstd-compressed tar archive the server embeds
// (web/public/defaultTheme/dist.tar.zst) and verifies that the result is
// self-consistent before writing it.
//
// Usage:
//
//	go run ./cmd/pack-frontend -dist <dist-dir> -out <dist.tar.zst> [-theme komari-theme.json]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/komari-monitor/komari/internal/frontenddist"
)

func main() {
	distDir := flag.String("dist", "", "path to the built frontend dist directory")
	outPath := flag.String("out", "", "output path for the packed archive (dist.tar.zst)")
	themePath := flag.String("theme", "", "optional path to komari-theme.json, copied next to -out")
	skipVerify := flag.Bool("skip-verify", false, "skip the frontend consistency check")
	flag.Parse()

	if *distDir == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "pack-frontend: -dist and -out are required")
		flag.PrintDefaults()
		os.Exit(2)
	}

	files, err := frontenddist.LoadDir(*distDir)
	if err != nil {
		fatal(err)
	}

	if !*skipVerify {
		if err := frontenddist.Verify(files); err != nil {
			fatal(fmt.Errorf("frontend consistency check failed: %w", err))
		}
	}

	archive, err := frontenddist.Encode(files)
	if err != nil {
		fatal(err)
	}

	decoded, err := frontenddist.Decode(archive)
	if err != nil {
		fatal(fmt.Errorf("packed archive failed to decode: %w", err))
	}
	if len(decoded) != len(files) {
		fatal(fmt.Errorf("packed archive contains %d files, expected %d", len(decoded), len(files)))
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*outPath, archive, 0o644); err != nil {
		fatal(err)
	}

	if *themePath != "" {
		content, err := os.ReadFile(*themePath)
		if err != nil {
			fatal(err)
		}
		dst := filepath.Join(filepath.Dir(*outPath), "komari-theme.json")
		if err := os.WriteFile(dst, content, 0o644); err != nil {
			fatal(err)
		}
	}

	fmt.Printf("pack-frontend: packed %d files into %s (%d bytes)\n", len(files), *outPath, len(archive))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "pack-frontend:", err)
	os.Exit(1)
}

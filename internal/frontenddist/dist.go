// Package frontenddist encodes, decodes and verifies the packed default-theme
// frontend that the server embeds from web/public/defaultTheme/dist.tar.zst.
//
// The archive is a zstd-compressed tar whose entries are relative to the built
// frontend "dist" directory (for example "index.html" and "assets/entry-x.js").
package frontenddist

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/klauspost/compress/zstd"
)

// IndexFile is the SPA entry document expected at the archive root.
const IndexFile = "index.html"

// Decode decompresses and unpacks a dist.tar.zst archive into a map keyed by
// the path relative to the dist root.
func Decode(archive []byte) (map[string][]byte, error) {
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("open zstd archive: %w", err)
	}
	defer decoder.Close()

	tarBytes, err := decoder.DecodeAll(archive, nil)
	if err != nil {
		return nil, fmt.Errorf("decode zstd archive: %w", err)
	}

	files := make(map[string][]byte)
	reader := tar.NewReader(bytes.NewReader(tarBytes))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar entry: %w", err)
		}

		name := path.Clean(strings.ReplaceAll(header.Name, "\\", "/"))
		if header.Typeflag == tar.TypeDir && name == "." {
			continue
		}
		if name == "." || name == ".." ||
			strings.HasPrefix(name, "../") ||
			strings.HasPrefix(name, "/") ||
			strings.ContainsRune(name, '\x00') {
			return nil, fmt.Errorf("invalid embedded tar path %q", name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg, tar.TypeRegA:
			content, err := io.ReadAll(reader)
			if err != nil {
				return nil, fmt.Errorf("read tar entry %q: %w", name, err)
			}
			if _, exists := files[name]; exists {
				return nil, fmt.Errorf("duplicate tar entry %q", name)
			}
			files[name] = content
		default:
			return nil, fmt.Errorf("unsupported tar entry %q type %d", name, header.Typeflag)
		}
	}

	if _, ok := files[IndexFile]; !ok {
		return nil, fmt.Errorf("tar archive does not contain %q", IndexFile)
	}
	return files, nil
}

// LoadDir reads a built frontend directory into the same map representation as
// Decode, keyed by the slash-separated path relative to dir.
func LoadDir(dir string) (map[string][]byte, error) {
	root := filepath.Clean(dir)
	files := make(map[string][]byte)
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(root, current)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(current)
		if err != nil {
			return fmt.Errorf("read %s: %w", current, err)
		}
		files[filepath.ToSlash(rel)] = content
		return nil
	})
	if err != nil {
		return nil, err
	}
	if _, ok := files[IndexFile]; !ok {
		return nil, fmt.Errorf("%s does not contain %q", root, IndexFile)
	}
	return files, nil
}

// Encode packs the map into a deterministic zstd-compressed tar archive.
func Encode(files map[string][]byte) ([]byte, error) {
	if _, ok := files[IndexFile]; !ok {
		return nil, fmt.Errorf("refusing to pack an archive without %q", IndexFile)
	}

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	for _, name := range names {
		content := files[name]
		header := &tar.Header{
			Name:     "./" + filepath.ToSlash(name),
			Mode:     0o644,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		}
		if err := tw.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("write tar header %q: %w", name, err)
		}
		if _, err := tw.Write(content); err != nil {
			return nil, fmt.Errorf("write tar entry %q: %w", name, err)
		}
	}
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("close tar: %w", err)
	}

	var zBuf bytes.Buffer
	zw, err := zstd.NewWriter(&zBuf,
		zstd.WithEncoderLevel(zstd.SpeedBestCompression),
		zstd.WithEncoderConcurrency(1),
	)
	if err != nil {
		return nil, fmt.Errorf("open zstd writer: %w", err)
	}
	if _, err := zw.Write(tarBuf.Bytes()); err != nil {
		return nil, fmt.Errorf("compress archive: %w", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close zstd writer: %w", err)
	}
	return zBuf.Bytes(), nil
}

var (
	htmlRefRe       = regexp.MustCompile(`(?i)\s(?:src|href)\s*=\s*["']([^"']+)["']`)
	jsFromRe        = regexp.MustCompile(`\bfrom\s*["']([^"']+)["']`)
	jsImportCallRe  = regexp.MustCompile(`\bimport\s*\(\s*["']([^"']+)["']\s*\)`)
	jsImportBareRe  = regexp.MustCompile(`\bimport\s*["']([^"']+)["']`)
	jsImportRegexps = []*regexp.Regexp{jsFromRe, jsImportCallRe, jsImportBareRe}
)

// referenceExtensions are the extensions that Verify resolves and requires to
// exist. HTML references may use any of them; the JS import graph only requires
// the module/stylesheet types.
var referenceExtensions = map[string]struct{}{
	".js": {}, ".mjs": {}, ".cjs": {}, ".css": {},
	".json": {}, ".webmanifest": {},
	".png": {}, ".jpg": {}, ".jpeg": {}, ".gif": {}, ".svg": {},
	".webp": {}, ".avif": {}, ".ico": {},
	".woff": {}, ".woff2": {}, ".ttf": {}, ".otf": {},
	".wasm": {}, ".map": {},
}

var moduleExtensions = map[string]struct{}{
	".js": {}, ".mjs": {}, ".cjs": {}, ".css": {},
}

// Verify checks that every asset referenced by the packed index.html - and by
// the ES module import graph of the packed chunks - actually exists inside the
// archive. A mismatch here means the embedded frontend is incomplete and would
// produce a blank page (or "Failed to fetch dynamically imported module") at
// runtime, so it should fail the build.
func Verify(files map[string][]byte) error {
	index, ok := files[IndexFile]
	if !ok {
		return fmt.Errorf("archive is missing %q", IndexFile)
	}

	seen := make(map[string]struct{})
	var missing []string
	record := func(from, specifier string, allowed map[string]struct{}) {
		key, ok := resolveReference(from, specifier, allowed)
		if !ok {
			return
		}
		if _, exists := files[key]; exists {
			return
		}
		message := fmt.Sprintf("%s (referenced by %s)", key, from)
		if _, duplicate := seen[message]; duplicate {
			return
		}
		seen[message] = struct{}{}
		missing = append(missing, message)
	}

	for _, match := range htmlRefRe.FindAllStringSubmatch(string(index), -1) {
		record(IndexFile, match[1], referenceExtensions)
	}

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if !strings.HasSuffix(strings.ToLower(name), ".js") {
			continue
		}
		content := files[name]
		for _, pattern := range jsImportRegexps {
			for _, match := range pattern.FindAllSubmatch(content, -1) {
				record(name, string(match[1]), moduleExtensions)
			}
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf(
			"frontend archive references %d file(s) that are not present:\n  %s",
			len(missing), strings.Join(missing, "\n  "),
		)
	}
	return nil
}

// resolveReference turns a specifier found in from into an archive key. It
// returns ok=false for external URLs and for specifiers that are not local
// asset files of an allowed type.
func resolveReference(from, specifier string, allowed map[string]struct{}) (string, bool) {
	specifier = strings.TrimSpace(specifier)
	if i := strings.IndexAny(specifier, "?#"); i >= 0 {
		specifier = specifier[:i]
	}
	if specifier == "" || strings.HasPrefix(specifier, "//") {
		return "", false
	}
	if strings.Contains(specifier, "://") ||
		strings.HasPrefix(specifier, "data:") ||
		strings.HasPrefix(specifier, "blob:") ||
		strings.HasPrefix(specifier, "javascript:") {
		return "", false
	}

	var key string
	switch {
	case strings.HasPrefix(specifier, "/"):
		key = strings.TrimPrefix(specifier, "/")
	case strings.HasPrefix(specifier, "./") || strings.HasPrefix(specifier, "../"):
		key = path.Clean(path.Join(path.Dir(from), specifier))
	default:
		return "", false
	}

	if key == "" || key == "." || key == ".." || strings.HasPrefix(key, "../") {
		return "", false
	}
	if _, ok := allowed[strings.ToLower(path.Ext(key))]; !ok {
		return "", false
	}
	return key, true
}

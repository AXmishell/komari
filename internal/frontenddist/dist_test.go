package frontenddist

import (
	"strings"
	"testing"
)

func sampleFiles() map[string][]byte {
	return map[string][]byte{
		IndexFile: []byte(`<!doctype html><html><head>` +
			`<link rel="stylesheet" href="/assets/index-abc.css">` +
			`<link rel="modulepreload" href="/assets/chunk-_layout-abc.js">` +
			`</head><body><div id="root"></div>` +
			`<script type="module" crossorigin src="/assets/entry-index-abc.js"></script>` +
			`</body></html>`),
		"assets/entry-index-abc.js": []byte(`import"./chunk-_layout-abc.js";` +
			`const s=document.createElement("link");` +
			`__vitePreload(()=>import("./chunk-themeSettings-def.js"));`),
		"assets/chunk-_layout-abc.js":       []byte(`from"./chunk-vendor-abc.js";`),
		"assets/chunk-vendor-abc.js":        []byte(`export const x=1;`),
		"assets/chunk-themeSettings-def.js": []byte(`export const y=2;`),
		"assets/index-abc.css":              []byte(`body{}`),
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	files := sampleFiles()
	archive, err := Encode(files)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	decoded, err := Decode(archive)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(decoded) != len(files) {
		t.Fatalf("decoded %d files, want %d", len(decoded), len(files))
	}
	for name, content := range files {
		got, ok := decoded[name]
		if !ok {
			t.Fatalf("decoded archive is missing %q", name)
		}
		if string(got) != string(content) {
			t.Fatalf("decoded %q mismatch:\n got %q\nwant %q", name, got, content)
		}
	}
}

func TestEncodeRejectsArchiveWithoutIndex(t *testing.T) {
	if _, err := Encode(map[string][]byte{"assets/x.js": []byte("x")}); err == nil {
		t.Fatal("Encode accepted an archive without index.html")
	}
}

func TestVerifyAcceptsConsistentArchive(t *testing.T) {
	if err := Verify(sampleFiles()); err != nil {
		t.Fatalf("Verify rejected a consistent archive: %v", err)
	}
}

func TestVerifyDetectsMissingReferencedChunk(t *testing.T) {
	files := sampleFiles()
	delete(files, "assets/chunk-_layout-abc.js")

	err := Verify(files)
	if err == nil {
		t.Fatal("Verify accepted an archive with a missing referenced chunk")
	}
	if !strings.Contains(err.Error(), "assets/chunk-_layout-abc.js") {
		t.Fatalf("Verify error does not name the missing chunk: %v", err)
	}
}

func TestVerifyDetectsMissingIndexReference(t *testing.T) {
	files := sampleFiles()
	delete(files, "assets/index-abc.css")

	err := Verify(files)
	if err == nil {
		t.Fatal("Verify accepted an archive with a missing stylesheet")
	}
	if !strings.Contains(err.Error(), "assets/index-abc.css") {
		t.Fatalf("Verify error does not name the missing stylesheet: %v", err)
	}
}

func TestVerifyIgnoresExternalReferences(t *testing.T) {
	files := map[string][]byte{
		IndexFile: []byte(`<html><head>` +
			`<link rel="icon" href="https://cdn.example.com/favicon.ico">` +
			`<link rel="manifest" href="/manifest.webmanifest">` +
			`</head><body></body></html>`),
		"manifest.webmanifest": []byte(`{}`),
	}
	if err := Verify(files); err != nil {
		t.Fatalf("Verify should ignore external URLs: %v", err)
	}
}

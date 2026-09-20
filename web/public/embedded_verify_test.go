package public

import (
	"testing"

	"github.com/komari-monitor/komari/internal/frontenddist"
)

// TestEmbeddedFrontendIsSelfConsistent guards the build: if the packed
// defaultTheme archive is missing any asset referenced by its index.html or by
// the ES module import graph of its chunks, the running server would serve
// index.html (text/html) for a .js request and blank the admin UI.
func TestEmbeddedFrontendIsSelfConsistent(t *testing.T) {
	if err := frontenddist.Verify(defaultDistFiles); err != nil {
		t.Fatalf("embedded default frontend is inconsistent: %v", err)
	}
}

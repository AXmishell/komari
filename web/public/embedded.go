package public

import "github.com/komari-monitor/komari/internal/frontenddist"

// defaultDistFiles holds the embedded default-theme frontend keyed by the path
// relative to its dist directory (for example "index.html", "assets/x.js").
var defaultDistFiles map[string][]byte

func loadEmbeddedDist() (map[string][]byte, error) {
	return frontenddist.Decode(embeddedDistArchive)
}

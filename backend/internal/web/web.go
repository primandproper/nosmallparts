// Package web serves the browser app the binary carries: the Vite build of
// frontend/, which `make web` copies into dist before the binary is built,
// since go:embed cannot reach outside this module.
package web

import (
	"bytes"
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
)

// Route is the pattern the app is mounted at: every GET no other route claims.
const Route = "/"

const (
	handlerName = "web_app_handler"
	indexFile   = "index.html"

	// assetsDir holds Vite's content-hashed output: a name there never changes
	// contents, so a browser can keep it forever.
	assetsDir      = "assets/"
	cacheImmutable = "public, max-age=31536000, immutable"
	// cacheRevalidate keeps index.html and the unhashed files fresh, so a
	// deploy reaches a returning browser on its next load.
	cacheRevalidate = "no-cache"

	notBuilt = "The web app was not built into this binary. Run `make web` before building it.\n"
)

//go:embed all:dist
var embedded embed.FS

// Handler serves the app's files, and index.html for any other path, so a
// client-side route such as /stage/ABCDEF loads directly.
type Handler struct {
	files  fs.FS
	logger logging.Logger
	index  []byte
}

// NewHandler builds the handler over the embedded build. A binary built without
// one still starts, answering with a note instead of the app.
func NewHandler(logger logging.Logger) (*Handler, error) {
	files, err := fs.Sub(embedded, "dist")
	if err != nil {
		return nil, err
	}

	return newHandler(files, observability.NewObserver(handlerName, logger, nil).Logger()), nil
}

func newHandler(files fs.FS, logger logging.Logger) *Handler {
	index, err := fs.ReadFile(files, indexFile)
	if err != nil {
		logger.Warn("serving no web app: this binary was built without one")
		index = nil
	}

	return &Handler{files: files, logger: logger, index: index}
}

// ServeHTTP answers with the file the path names when the build has one. A
// missing hashed asset is a 404, since answering a script tag with HTML only
// hides the mistake; anything else is a client-side route, answered with
// index.html.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")

	if name != "" && name != indexFile {
		if info, err := fs.Stat(h.files, name); err == nil && !info.IsDir() {
			if strings.HasPrefix(name, assetsDir) {
				w.Header().Set("Cache-Control", cacheImmutable)
			} else {
				w.Header().Set("Cache-Control", cacheRevalidate)
			}

			http.ServeFileFS(w, r, h.files, name)

			return
		}

		if strings.HasPrefix(name, assetsDir) {
			http.NotFound(w, r)

			return
		}
	}

	if h.index == nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte(notBuilt)); err != nil {
			h.logger.Error("writing the not-built note", err)
		}

		return
	}

	w.Header().Set("Cache-Control", cacheRevalidate)
	http.ServeContent(w, r, indexFile, time.Time{}, bytes.NewReader(h.index))
}

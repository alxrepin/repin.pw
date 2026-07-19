package get

import (
	"context"
	"errors"
	"net/http"
	"path"
	"strings"

	"github.com/gorilla/mux"

	"repin/internal/context/domain"
	"repin/internal/context/infrastructure/storage/minio"
)

type storage interface {
	Get(ctx context.Context, objectName string) (*minio.Object, error)
}

type Controller struct {
	storage storage
}

func NewController(storage storage) *Controller {
	return &Controller{storage: storage}
}

// Handle streams an object straight from the bucket so the storage endpoint is
// never exposed. It is a raw http.HandlerFunc rather than a typed controller
// because the body is bytes, not JSON.
func (c *Controller) Handle(w http.ResponseWriter, r *http.Request) {
	key := cleanKey(mux.Vars(r)["key"])
	if key == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	obj, err := c.storage.Get(r.Context(), key)
	if err != nil {
		if errors.Is(err, domain.ErrMediaNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}
	defer func() { _ = obj.Body.Close() }()

	if obj.ContentType != "" {
		w.Header().Set("Content-Type", obj.ContentType)
	}

	// Keys embed the Telegram file id, so a key always denotes the same bytes.
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	// ServeContent handles Range, If-Modified-Since and HEAD, which is what
	// makes video seeking work.
	http.ServeContent(w, r, path.Base(key), obj.ModTime, obj.Body)
}

// cleanKey rejects traversal outside the bucket prefix. path.Clean collapses
// any "..", and a leading slash would address the bucket root.
func cleanKey(raw string) string {
	if raw == "" || strings.Contains(raw, "..") {
		return ""
	}

	return strings.TrimPrefix(path.Clean("/"+raw), "/")
}

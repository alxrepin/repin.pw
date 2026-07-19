// Package media builds the public links under which stored objects are served.
package media

const routePrefix = "/api/v1/media/"

// URL turns a storage key into the link clients fetch. base is the public
// origin of the API; when it is empty the link stays root-relative, which keeps
// the API usable without knowing how it is exposed.
func URL(base, key string) string {
	if key == "" {
		return ""
	}

	return base + routePrefix + key
}

// URLPtr is the optional-field variant: no key means no link.
func URLPtr(base string, key *string) *string {
	if key == nil || *key == "" {
		return nil
	}

	url := URL(base, *key)

	return &url
}

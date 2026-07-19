// Package media builds the public links under which stored objects are served.
package media

// URL turns a storage key into the link clients fetch. base is the public
// origin serving the bucket; when it is empty the key is returned untouched,
// which keeps the API usable before that origin is configured.
func URL(base, key string) string {
	if key == "" {
		return ""
	}

	if base == "" {
		return key
	}

	return base + "/" + key
}

// URLPtr is the optional-field variant: no key means no link.
func URLPtr(base string, key *string) *string {
	if key == nil || *key == "" {
		return nil
	}

	url := URL(base, *key)

	return &url
}

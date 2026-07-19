package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type RequestPaginate struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func decodeRequest[T any](r *http.Request) (T, error) {
	var request T

	stringKeys := stringFieldKeys(reflect.TypeOf(request))

	merged := map[string]any{}

	if r.ContentLength > 0 {
		body := map[string]any{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return request, fmt.Errorf("%w: %s", ErrRequestDecode, err)
		}

		for k, v := range body {
			merged[k] = v
		}
	}

	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			merged[key] = coerce(values[0], stringKeys[key])
		}
	}

	for key, value := range mux.Vars(r) {
		merged[key] = coerce(value, stringKeys[key])
	}

	raw, err := json.Marshal(merged)
	if err != nil {
		return request, fmt.Errorf("%w: %s", ErrRequestDecode, err)
	}

	if err := json.Unmarshal(raw, &request); err != nil {
		return request, fmt.Errorf("%w: %s", ErrRequestDecode, err)
	}

	return request, nil
}

func coerce(s string, isString bool) any {
	if isString {
		return s
	}

	if n, err := strconv.Atoi(s); err == nil {
		return n
	}

	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}

	return s
}

func stringFieldKeys(t reflect.Type) map[string]bool {
	keys := map[string]bool{}
	collectStringKeys(t, keys)

	return keys
}

func collectStringKeys(t reflect.Type, keys map[string]bool) {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t == nil || t.Kind() != reflect.Struct {
		return
	}

	for i := range t.NumField() {
		field := t.Field(i)

		if field.Anonymous {
			collectStringKeys(field.Type, keys)
			continue
		}

		ft := field.Type
		if ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}

		if ft.Kind() != reflect.String {
			continue
		}

		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			name = field.Name
		}

		keys[name] = true
	}
}

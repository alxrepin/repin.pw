package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var durationType = reflect.TypeFor[time.Duration]()

func MustLoad[T any](cfg T) *T {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(fmt.Sprintf("config: loading .env: %s", err))
	}

	t := reflect.TypeOf(cfg)
	if t.Kind() == reflect.Pointer {
		panic("config: MustLoad expects a non-pointer config type")
	}

	out := reflect.New(t).Interface().(*T)
	hydrate(reflect.ValueOf(out).Elem())

	return out
}

func hydrate(v reflect.Value) {
	t := v.Type()

	for i := range t.NumField() {
		field := v.Field(i)
		meta := t.Field(i)

		if field.Kind() == reflect.Struct && field.Type() != durationType {
			hydrate(field)
			continue
		}

		name := meta.Tag.Get("env")
		if name == "" {
			continue
		}

		def := meta.Tag.Get("envDefault")

		raw, ok := os.LookupEnv(name)
		if !ok || raw == "" {
			raw = def
		}

		if field.Kind() == reflect.Pointer {
			if raw == "" {
				continue
			}

			ptr := reflect.New(field.Type().Elem())
			set(ptr.Elem(), name, raw)
			field.Set(ptr)

			continue
		}

		if raw == "" {
			panic(fmt.Sprintf("config: required env %q is not set", name))
		}

		set(field, name, raw)
	}
}

func set(field reflect.Value, name, raw string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)

	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		must(name, err)
		field.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == durationType {
			d, err := time.ParseDuration(raw)
			must(name, err)
			field.SetInt(int64(d))

			return
		}

		n, err := strconv.ParseInt(raw, 10, 64)
		must(name, err)
		field.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(raw, 10, 64)
		must(name, err)
		field.SetUint(n)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(raw, 64)
		must(name, err)
		field.SetFloat(f)

	case reflect.Slice:
		if field.Type().Elem().Kind() != reflect.String {
			panic(fmt.Sprintf("config: %q: only []string slices are supported", name))
		}

		field.Set(reflect.ValueOf(splitList(raw)))

	default:
		panic(fmt.Sprintf("config: %q: unsupported field kind %s", name, field.Kind()))
	}
}

func splitList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}

	return out
}

func must(name string, err error) {
	if err != nil {
		panic(fmt.Sprintf("config: %q: %s", name, err))
	}
}

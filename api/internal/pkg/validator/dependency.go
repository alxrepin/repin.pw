package validator

import (
	"fmt"
	"reflect"
)

func ValidateStructDependencies(s any) error {
	v := reflect.ValueOf(s)

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return fmt.Errorf("nil pointer to %s", v.Type().Elem().Name())
		}

		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", v.Kind())
	}

	return validateFields(v, v.Type().Name())
}

func validateFields(v reflect.Value, prefix string) error {
	t := v.Type()

	for i := range v.NumField() {
		field := v.Field(i)
		name := fmt.Sprintf("%s.%s", prefix, t.Field(i).Name)

		switch field.Kind() {
		case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
			if field.IsNil() {
				return fmt.Errorf("%s is nil", name)
			}

		case reflect.Struct:
			if err := validateFields(field, name); err != nil {
				return err
			}

		default:
			// value kinds (numbers, strings, bools) carry no dependency
		}
	}

	return nil
}

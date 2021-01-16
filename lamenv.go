package lamenv

import (
	"os"
	"reflect"
	"strconv"
	"strings"
)

func Unmarshal(object interface{}, parts []string) error {
	return decode(reflect.ValueOf(object), parts)
}

func decode(conf reflect.Value, parts []string) error {
	v := conf
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			// if the pointer is not initialized, then accessing to its element will return `reflect.invalid`
			// So we have to create a new instance of the pointer first
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		decodeString(v, parts)
	case reflect.Bool:
		if err := decodeBool(v, parts); err != nil {
			return err
		}
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		if err := decodeInt(v, parts); err != nil {
			return err
		}
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		if err := decodeUInt(v, parts); err != nil {
			return err
		}
	case reflect.Float32,
		reflect.Float64:
		if err := decodeFloat(v, parts); err != nil {
			return err
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			attr := v.Field(i)
			attrType := v.Type().Field(i)
			if err := decode(attr, append(parts, attrType.Name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeString(v reflect.Value, parts []string) {
	if s, ok := lookupEnv(parts); ok {
		v.SetString(s)
	}
}

func decodeBool(v reflect.Value, parts []string) error {
	if s, ok := lookupEnv(parts); ok {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(b)
	}
	return nil
}

func decodeInt(v reflect.Value, parts []string) error {
	if s, ok := lookupEnv(parts); ok {
		i, err := strconv.ParseInt(s, 10, 0)
		if err != nil {
			return err
		}
		v.SetInt(i)
	}
	return nil
}

func decodeUInt(v reflect.Value, parts []string) error {
	if s, ok := lookupEnv(parts); ok {
		i, err := strconv.ParseUint(s, 10, 0)
		if err != nil {
			return err
		}
		v.SetUint(i)
	}
	return nil
}

func decodeFloat(v reflect.Value, parts []string) error {
	if s, ok := lookupEnv(parts); ok {
		i, err := strconv.ParseFloat(s, 10)
		if err != nil {
			return err
		}
		v.SetFloat(i)
	}
	return nil
}

func lookupEnv(parts []string) (string, bool) {
	return os.LookupEnv(buildEnvVariable(parts))
}

func buildEnvVariable(parts []string) string {
	newParts := make([]string, len(parts))
	for i, s := range parts {
		newParts[i] = strings.ToUpper(s)
	}
	return strings.Join(newParts, "_")
}

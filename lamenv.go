package lamenv

import (
	"os"
	"reflect"
	"strconv"
	"strings"
)

func Unmarshal(object interface{}, parts []string) error {
	return New().Unmarshall(object, parts)
}

type Lamenv struct {
	// TagSupports is a list of tag like "yaml", "json"
	// that the code will look at it to know the name of the field
	TagSupports []string
}

func New() *Lamenv {
	return &Lamenv{
		TagSupports: []string{
			"yaml", "json", "mapstructure",
		},
	}
}

func (l *Lamenv) Unmarshall(object interface{}, parts []string) error {
	return l.decode(reflect.ValueOf(object), parts)
}

func (l *Lamenv) decode(conf reflect.Value, parts []string) error {
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
		l.decodeString(v, parts)
	case reflect.Bool:
		if err := l.decodeBool(v, parts); err != nil {
			return err
		}
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		if err := l.decodeInt(v, parts); err != nil {
			return err
		}
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		if err := l.decodeUInt(v, parts); err != nil {
			return err
		}
	case reflect.Float32,
		reflect.Float64:
		if err := l.decodeFloat(v, parts); err != nil {
			return err
		}
	case reflect.Struct:
		if err := l.decodeStruct(v, parts); err != nil {
			return err
		}
	}
	return nil
}

func (l *Lamenv) decodeString(v reflect.Value, parts []string) {
	if s, ok := lookupEnv(parts); ok {
		v.SetString(s)
	}
}

func (l *Lamenv) decodeBool(v reflect.Value, parts []string) error {
	if s, ok := lookupEnv(parts); ok {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(b)
	}
	return nil
}

func (l *Lamenv) decodeInt(v reflect.Value, parts []string) error {
	if s, ok := lookupEnv(parts); ok {
		i, err := strconv.ParseInt(s, 10, 0)
		if err != nil {
			return err
		}
		v.SetInt(i)
	}
	return nil
}

func (l *Lamenv) decodeUInt(v reflect.Value, parts []string) error {
	if s, ok := lookupEnv(parts); ok {
		i, err := strconv.ParseUint(s, 10, 0)
		if err != nil {
			return err
		}
		v.SetUint(i)
	}
	return nil
}

func (l *Lamenv) decodeFloat(v reflect.Value, parts []string) error {
	if s, ok := lookupEnv(parts); ok {
		i, err := strconv.ParseFloat(s, 10)
		if err != nil {
			return err
		}
		v.SetFloat(i)
	}
	return nil
}

func (l *Lamenv) decodeStruct(v reflect.Value, parts []string) error {
	for i := 0; i < v.NumField(); i++ {
		attr := v.Field(i)
		attrType := v.Type().Field(i)
		attrName, ok := l.lookupTag(attrType.Tag)
		if ok {
			if attrName == ",squash" {
				if err := l.decode(attr, parts); err != nil {
					return err
				}
				continue
			}
			if strings.Contains(attrName, "omitempty") {
				// it means there is no variables
				// related to the current field, so we can ignore it
				if !exists(append(parts, attrName)) {
					continue
				}
			}
		} else {
			attrName = attrType.Name
		}
		if err := l.decode(attr, append(parts, attrName)); err != nil {
			return err
		}
	}
	return nil
}

func (l *Lamenv) lookupTag(tag reflect.StructTag) (string, bool) {
	for _, tagSupport := range l.TagSupports {
		if s, ok := tag.Lookup(tagSupport); ok {
			return s, ok
		}
	}
	return "", false
}

func lookupEnv(parts []string) (string, bool) {
	return os.LookupEnv(buildEnvVariable(parts))
}

func exists(prefixes []string) bool {
	variable := buildEnvVariable(prefixes)
	for _, e := range os.Environ() {
		envSplit := strings.Split(e, "=")
		if len(envSplit) != 2 {
			continue
		}
		if strings.Contains(envSplit[0], variable) {
			return true
		}
	}
	return false
}

func buildEnvVariable(parts []string) string {
	newParts := make([]string, len(parts))
	for i, s := range parts {
		newParts[i] = strings.ToUpper(s)
	}
	return strings.Join(newParts, "_")
}

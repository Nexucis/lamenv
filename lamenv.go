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
	case reflect.Slice:
		if err := l.decodeSlice(v, parts); err != nil {
			return err
		}
	case reflect.Struct:
		if err := l.decodeStruct(v, parts); err != nil {
			return err
		}
	default:
		if input, ok := lookupEnv(parts); ok {
			return l.decodeNative(v, input)
		}
	}
	return nil
}

func (l *Lamenv) decodeNative(v reflect.Value, input string) error {
	switch v.Kind() {
	case reflect.String:
		l.decodeString(v, input)
	case reflect.Bool:
		if err := l.decodeBool(v, input); err != nil {
			return err
		}
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		if err := l.decodeInt(v, input); err != nil {
			return err
		}
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		if err := l.decodeUInt(v, input); err != nil {
			return err
		}
	case reflect.Float32,
		reflect.Float64:
		if err := l.decodeFloat(v, input); err != nil {
			return err
		}
	}
	return nil
}

func (l *Lamenv) decodeString(v reflect.Value, input string) {
	v.SetString(input)
}

func (l *Lamenv) decodeBool(v reflect.Value, input string) error {
	b, err := strconv.ParseBool(strings.TrimSpace(input))
	if err != nil {
		return err
	}
	v.SetBool(b)
	return nil
}

func (l *Lamenv) decodeInt(v reflect.Value, input string) error {
	i, err := strconv.ParseInt(strings.TrimSpace(input), 10, 0)
	if err != nil {
		return err
	}
	v.SetInt(i)
	return nil
}

func (l *Lamenv) decodeUInt(v reflect.Value, input string) error {
	i, err := strconv.ParseUint(strings.TrimSpace(input), 10, 0)
	if err != nil {
		return err
	}
	v.SetUint(i)
	return nil
}

func (l *Lamenv) decodeFloat(v reflect.Value, input string) error {
	i, err := strconv.ParseFloat(strings.TrimSpace(input), 10)
	if err != nil {
		return err
	}
	v.SetFloat(i)
	return nil
}

// decodeSlice will support two different syntax for the slice of native type (and only one for the struct)
// First one would be to have a single variable containing the whole slice. Each items are separated by a comma.
// The second one would be:
//                <PREFIX>_<SLICE_INDEX>(_<SUFFIX>)?
// Note: the second syntax would be the only one supported for the slice of struct
func (l *Lamenv) decodeSlice(v reflect.Value, parts []string) error {
	sliceType := v.Type().Elem()
	if isNative(sliceType) && exists(parts) {
		e, _ := lookupEnv(parts)
		for _, s := range strings.Split(e, ",") {
			tmp := reflect.New(sliceType).Elem()
			if err := l.decodeNative(tmp, s); err != nil {
				return err
			}
			v.Set(reflect.Append(v, tmp))
		}
	} else {
		// Second syntax. While we are able to find an environment variable that is starting by <PREFIX>_<SLICE_INDEX>
		//  then it will create a new item in a slice and will use the next recursive loop to set it.
		i := 0
		for ok := contains(append(parts, strconv.Itoa(i))); ok; ok = contains(append(parts, strconv.Itoa(i))) {
			// create a new item and pass it to the method decode to be able to "decode" its value
			tmp := reflect.New(sliceType).Elem()
			if err := l.decode(tmp, append(parts, strconv.Itoa(i))); err != nil {
				return err
			}
			v.Set(reflect.Append(v, tmp))
			i++
		}
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
				if !contains(append(parts, attrName)) {
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

func contains(parts []string) bool {
	variable := buildEnvVariable(parts)
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

func exists(parts []string) bool {
	_, ok := lookupEnv(parts)
	return ok
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

func isNative(t reflect.Type) bool {
	kind := t.Kind()
	if kind == reflect.Ptr {
		kind = t.Elem().Kind()
	}
	return kind != reflect.Slice && kind != reflect.Struct && kind != reflect.Map
}

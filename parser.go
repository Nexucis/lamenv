package lamenv

import "reflect"

type nodeKind uint

const (
	node nodeKind = iota
	root
	leaf
)

// ring will represent the struct hold by reflect.Type
type ring struct {
	kind     nodeKind
	value    string
	children []*ring
}

func newRing(t reflect.Type, tag []string) *ring {
	root := &ring{
		kind: root,
	}
	root.buildRing(t, tag)
	return root
}

func (r *ring) buildRing(t reflect.Type, tag []string) {
	switch t.Kind() {
	case reflect.Ptr:
		r.buildRing(t.Elem(), tag)
	case reflect.Slice,
		reflect.Array:
		r.value = r.value + "_0"
		r.buildRing(t.Elem(), tag)
	case reflect.Struct:
		// to avoid looking at the unexported type, we will create a value to then have access to the method reflect.CanSet()
		v := reflect.Indirect(reflect.New(t))
		for i := 0; i < t.NumField(); i++ {
			if !v.Field(i).CanSet() {
				// the field is not exported, so no need to look at it as we won't be able to set it in a later stage
				continue
			}
			attrField := t.Field(i)
			attrName, ok := lookupTag(attrField.Tag, tag)
			if ok {
				if attrName == "-" {
					continue
				}
				if attrName == ",squash" || attrName == ",inline" {
					// in this case it just means the next node won't provide any additional value
					attrName = ""
				}
			} else {
				attrName = attrField.Name
			}
			node := &ring{
				kind:  node,
				value: attrName,
			}
			node.buildRing(attrField.Type, tag)
			r.children = append(r.children, node)
		}
	default:
		r.kind = leaf
	}
}

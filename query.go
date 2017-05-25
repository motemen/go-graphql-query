package graphqlquery

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
)

type Builder struct {
	Q      interface{}
	V      interface{}
	params map[interface{}]map[string]interface{}
	binds  map[interface{}]interface{}
}

func New(query interface{}, variables interface{}) *Builder {
	b := &Builder{Q: query, V: variables}
	b.scanParams()
	return b
}

func (b *Builder) Bind(ref interface{}, value interface{}) *Builder {
	if b.binds == nil {
		b.binds = map[interface{}]interface{}{}
	}
	b.binds[ref] = b.varName(value)
	return b
}

type varname struct {
	name string
}

func (v varname) GoString() string {
	return "$" + v.name
}

func (b *Builder) varName(value interface{}) *varname {
	rv, ok := reflectStruct(reflect.ValueOf(b.V))
	if !ok {
		panic("TODO")
	}

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		ft := rt.Field(i)
		fv := rv.Field(i)
		if fv.Addr().Interface() == value {
			return &varname{b.toName(ft.Name)}
		}
	}

	return nil
}

func (b *Builder) scanParams() error {
	rv, ok := reflectStruct(reflect.ValueOf(b.Q))
	if !ok {
		return fmt.Errorf("must be a struct or a pointer to struct %+v", b.Q)
	}

	b.scanParamsStruct(rv, reflect.Value{}, []string{})
	for k, v := range b.params {
		log.Printf("%+v %+v", k, v)
	}
	return nil
}

func (b *Builder) queryParams() string {
	params := []string{}
	for _, names := range b.params {
		for name, p := range names {
			param := fmt.Sprintf("%s: %s", "$"+b.toName(name), b.typeName(p))
			params = append(params, param)
		}
	}
	if len(params) == 0 {
		return ""
	}
	return "(" + strings.Join(params, ", ") + ")"
}

func (b *Builder) typeName(v interface{}) string {
	rt := reflect.TypeOf(v)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	switch rt.Kind() {
	case reflect.Array:
	case reflect.Bool:
		return "Boolean" // ?
	case reflect.Chan:
	case reflect.Complex128:
	case reflect.Complex64:
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
	case reflect.Func:
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		return "Int"
	case reflect.Interface:
	case reflect.Invalid:
	case reflect.Map:
	case reflect.Ptr:
	case reflect.Slice:
	case reflect.String:
		return "String"
	case reflect.Struct:
	case reflect.Uint:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
	case reflect.Uint8:
	case reflect.Uintptr:
	case reflect.UnsafePointer:
	}

	return rt.String()
}

func (b *Builder) scanParamsStruct(rv, parent reflect.Value, path []string) {
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		ft := rt.Field(i)
		fv, ok := reflectStruct(rv.Field(i))
		if !ok {
			continue
		}

		if isGraphQLParamField(ft) {
			for i := 0; i < fv.NumField(); i++ {
				b.addParam(
					rv.Addr().Interface(),
					ft.Type.Field(i).Name,
					fv.Field(i).Addr().Interface(),
				)
			}

			continue
		}

		newPath := make([]string, len(path)+1)
		copy(newPath, path)
		newPath[len(newPath)-1] = ft.Name

		b.scanParamsStruct(fv, rv, newPath)
	}
}

func (b *Builder) addParam(node interface{}, name string, param interface{}) {
	if b.params == nil {
		b.params = map[interface{}]map[string]interface{}{}
	}
	if b.params[node] == nil {
		b.params[node] = map[string]interface{}{}
	}

	b.params[node][name] = param
}

func isGraphQLParamField(t reflect.StructField) bool {
	if t.Name == "GraphQLParams" {
		return true
	}

	return false
}

func (b Builder) String() (string, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "query %s {\n", b.queryParams())
	err := b.toString(&buf, b.Q, 0)
	fmt.Fprintf(&buf, "}")
	return buf.String(), err
}

func (b Builder) toString(w io.Writer, v interface{}, depth int) error {
	rv, ok := reflectStruct(reflect.ValueOf(v))
	if !ok {
		return fmt.Errorf("invalid value: %+v", v)
	}

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		ft := rt.Field(i)
		if isGraphQLParamField(ft) {
			continue
		}

		fv := rv.Field(i)
		_, isStruct := reflectStruct(fv)
		if isStruct {
			params := ""
			if fv.CanAddr() && b.params[fv.Addr().Interface()] != nil {
				pp := []string{}
				for name, param := range b.params[fv.Addr().Interface()] {
					pp = append(pp, fmt.Sprintf("%s: %#v", b.toName(name), b.boundValue(param)))
				}
				params = "(" + strings.Join(pp, ", ") + ")"
			}
			fmt.Fprintf(w, "%s%s%s {\n", strings.Repeat(" ", depth*2), b.toName(ft.Name), params)
			b.toString(w, fv.Addr().Interface(), depth+1)
			fmt.Fprintf(w, "%s}\n", strings.Repeat(" ", depth*2))
		} else {
			fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", depth*2), b.toName(ft.Name))
		}
	}

	return nil
}

func (b Builder) boundValue(p interface{}) interface{} {
	if v, ok := b.binds[p]; ok {
		return v
	}
	return reflect.Indirect(reflect.ValueOf(p)).Interface()
}

func (b Builder) toName(name string) string {
	return strings.ToLower(name[0:1]) + name[1:]
}

func reflectStruct(rv reflect.Value) (reflect.Value, bool) {
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	return rv, rv.Kind() == reflect.Struct
}

type Var struct {
	V interface{}
}

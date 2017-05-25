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
}

func New(query interface{}, variables interface{}) *Builder {
	b := &Builder{Q: query, V: variables}
	b.scanParams()
	return b
}

func (b *Builder) Bind(ref interface{}, value interface{}) *Builder {
	return b
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
	err := b.toString(&buf, b.Q, 0)
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
			if fv.CanAddr() {
				pp := []string{}
				for name, param := range b.params[fv.Addr().Interface()] {
					pp = append(pp, fmt.Sprintf("%s: %v", b.toName(name), b.boundValue(param)))
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

func (b Builder) boundValue(interface{}) interface{} {
	return nil
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

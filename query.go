package graphqlquery

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type Builder struct {
	Q    interface{}
	args map[interface{}]map[string]argSpec
}

type argSpec struct {
	field reflect.StructField
	value reflect.Value
}

func (a argSpec) argValue() interface{} {
	value := getTag(a.field, 0)
	if value != "" {
		return value
	}

	return a.value.Interface()
}

func (a argSpec) variableName() string {
	name := getTag(a.field, 0)
	if strings.HasPrefix(name, "$") {
		return name
	}

	return ""
}

func getTag(f reflect.StructField, n int) string {
	tags := strings.Split(f.Tag.Get("graphql"), ",")
	if len(tags) > n {
		return tags[n]
	}
	return ""
}

func getTagNamed(f reflect.StructField, name string) string {
	tags := strings.Split(f.Tag.Get("graphql"), ",")
	for _, tag := range tags {
		if strings.HasPrefix(tag, name+"=") {
			return tag[len(name+"="):]
		}
	}
	return ""
}

func New(query interface{}) *Builder {
	b := &Builder{Q: query}
	b.scan()
	return b
}

func (b *Builder) scan() error {
	rv, ok := reflectStruct(reflect.ValueOf(b.Q))
	if !ok {
		return fmt.Errorf("must be a struct or a pointer to struct %+v", b.Q)
	}

	b.scanStruct(rv, reflect.Value{}, []string{})

	return nil
}

func (b *Builder) queryArguments() string {
	args := []string{}
	for _, names := range b.args {
		for _, spec := range names {
			varName := spec.variableName()
			if varName == "" {
				continue
			}
			param := fmt.Sprintf("%s: %s", varName, b.typeName(spec.field.Type))
			if getTag(spec.field, 1) == "notnull" {
				param += "!"
			}
			args = append(args, param)
		}
	}
	if len(args) == 0 {
		return ""
	}
	return "(" + strings.Join(args, ", ") + ")"
}

func (b *Builder) typeName(rt reflect.Type) string {
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	switch rt.Kind() {
	case reflect.Array, reflect.Slice:
		return "[" + b.typeName(rt.Elem()) + "]"
	case reflect.Bool:
		return "Boolean"
	case reflect.Float32, reflect.Float64:
		return "Float"
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		return "Int"
	case reflect.Ptr:
		return b.typeName(rt.Elem())
	case reflect.String:
		return "String"
	}

	return "" // TODO
}

func (b *Builder) scanStruct(rv, parent reflect.Value, path []string) {
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		ft := rt.Field(i)
		fv, ok := reflectStruct(rv.Field(i))
		if !ok {
			continue
		}

		if isArgumentsField(ft) {
			for i := 0; i < fv.NumField(); i++ {
				b.addArg(
					rv.Addr().Interface(),
					ft.Type.Field(i),
					fv.Field(i),
				)
			}

			continue
		}

		newPath := make([]string, len(path)+1)
		copy(newPath, path)
		newPath[len(newPath)-1] = ft.Name

		b.scanStruct(fv, rv, newPath)
	}
}

func (b *Builder) addArg(node interface{}, field reflect.StructField, value reflect.Value) {
	if b.args == nil {
		b.args = map[interface{}]map[string]argSpec{}
	}
	if b.args[node] == nil {
		b.args[node] = map[string]argSpec{}
	}

	b.args[node][field.Name] = argSpec{
		field: field,
		value: value,
	}
}

func isArgumentsField(t reflect.StructField) bool {
	if t.Name == "GraphQLArguments" {
		return true
	}

	return false
}

func (b Builder) String() (string, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "query%s {\n", b.queryArguments())
	err := b.toString(&buf, b.Q, 1)
	fmt.Fprintf(&buf, "}")
	return buf.String(), err
}

func (b Builder) writeStructField(w io.Writer, depth int, field reflect.StructField, value reflect.Value) error {
	directive := getTag(field, 0)
	if strings.HasPrefix(directive, "@") {
		directive = " " + directive
	} else {
		directive = ""
	}

	args := b.argsStringForField(value)
	fmt.Fprintf(w, "%s%s%s%s {\n", strings.Repeat(" ", depth*2), b.toName(field.Name), args, directive)
	var i interface{}
	if value.CanAddr() {
		i = value.Addr().Interface()
	} else {
		i = value.Interface()
	}
	err := b.toString(w, i, depth+1)
	fmt.Fprintf(w, "%s}\n", strings.Repeat(" ", depth*2))
	return err
}

func (b Builder) writeSimpleField(w io.Writer, depth int, field reflect.StructField) {
	fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", depth*2), b.toName(field.Name))
}

func (b Builder) toString(w io.Writer, v interface{}, depth int) error {
	rv, ok := reflectStruct(reflect.ValueOf(v))
	if !ok {
		return fmt.Errorf("invalid value: %+v", v)
	}

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		if isArgumentsField(field) {
			continue
		}

		value := rv.Field(i)
		switch {
		case isKindOf(value.Type(), reflect.Struct):
			b.writeStructField(w, depth, field, value)
			continue

		case isKindOf(value.Type(), reflect.Slice):
			if et := value.Type().Elem(); isKindOf(et, reflect.Struct) {
				b.writeStructField(w, depth, field, reflect.New(et))
				continue
			}
		}

		b.writeSimpleField(w, depth, field)
	}

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

func isKindOf(rt reflect.Type, kind reflect.Kind) bool {
	k := rt.Kind()
	if k == kind {
		return true
	}
	if k == reflect.Ptr {
		return rt.Elem().Kind() == kind
	}
	return false
}

type variable string

func (v variable) GoString() string {
	return "$" + string(v)
}

func (b Builder) argsStringForField(fv reflect.Value) string {
	args := ""
	if fv.CanAddr() && b.args[fv.Addr().Interface()] != nil {
		aa := []string{}
		for name, arg := range b.args[fv.Addr().Interface()] {
			// FIXME: %v is not correct, must use JSON
			aa = append(aa, fmt.Sprintf("%s: %v", b.toName(name), arg.argValue()))
		}
		args = "(" + strings.Join(aa, ", ") + ")"
	}
	return args
}

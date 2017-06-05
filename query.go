// Package graphqlquery generates GraphQL queries from the result structs.
package graphqlquery

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

// Build makes GraphQL query suitable for q, which is also
// a result JSON object for the GraphQL result JSON.
// See the example.
func Build(q interface{}) ([]byte, error) {
	b := &builder{
		query: q,
	}

	err := b.scan()
	if err != nil {
		return nil, err
	}

	return b.build()
}

type builder struct {
	query interface{}
	args  map[interface{}]map[string]argSpec
}

type argSpec struct {
	field reflect.StructField
	value reflect.Value
	path  []string
}

func (a argSpec) argValue() interface{} {
	value := getTag(a.field, 0)
	if value != "" {
		return value
	}

	return a.value.Interface()
}

func (a argSpec) variableName() string {
	return getTagWithPrefix(a.field, "$")
}

func getTag(field reflect.StructField, n int) string {
	tags := parseTags(field.Tag.Get("graphql"))
	if len(tags) > n {
		return tags[n]
	}
	return ""
}

func getTagWithPrefix(field reflect.StructField, prefix string) string {
	tags := parseTags(field.Tag.Get("graphql"))
	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {
			return tag
		}
	}

	return ""
}

func getTagNamed(f reflect.StructField, name string) string {
	tags := parseTags(f.Tag.Get("graphql"))
	for _, tag := range tags {
		if strings.HasPrefix(tag, name+"=") {
			return tag[len(name+"="):]
		}
	}
	return ""
}

func parseTags(s string) []string {
	tags := strings.Split(s, ",")
TAGS:
	for i := 0; i < len(tags); i++ {
		if tags[i] == "" {
			continue
		}

		if tags[i][0] != '(' {
			continue
		}

		for j := i; j < len(tags); j++ {
			if tags[j] == "" {
				continue
			}

			if tags[j][len(tags[j])-1] != ')' {
				continue
			}
			if j == i {
				continue TAGS
			} else {
				for k := i + 1; k <= j; k++ {
					tags[i] += "," + tags[k]
				}
				tags = append(tags[0:i+1], tags[j+1:]...)
			}
		}
	}

	return tags
}

func (b *builder) scan() error {
	rv, ok := reflectStruct(reflect.ValueOf(b.query))
	if !ok {
		return fmt.Errorf("must be a struct or a pointer to struct %+v", b.query)
	}

	b.scanStruct(rv, reflect.Value{}, []string{})

	return nil
}

func (b *builder) queryArguments() (string, error) {
	args := []string{}
	for _, names := range b.args {
		for name, spec := range names {
			varName := spec.variableName()
			if varName == "" {
				continue
			}
			typeName, err := b.typeName(spec.field.Type)
			if err != nil {
				path := strings.Join(spec.path, ".")
				if path == "" {
					path = "(root)"
				}
				return "", fmt.Errorf("argument %q at %s: %s", name, path, err)
			}
			param := fmt.Sprintf("%s: %s", varName, typeName)
			if getTag(spec.field, 1) == "notnull" {
				param += "!"
			}
			args = append(args, param)
		}
	}
	if len(args) == 0 {
		return "", nil
	}
	sort.Strings(args)
	return "(" + strings.Join(args, ", ") + ")", nil
}

func (b *builder) typeName(rt reflect.Type) (string, error) {
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if unicode.IsUpper(rune(rt.Name()[0])) {
		return rt.Name(), nil
	}

	switch rt.Kind() {
	case reflect.Array, reflect.Slice:
		name, err := b.typeName(rt.Elem())
		return "[" + name + "]", err
	case reflect.Bool:
		return "Boolean", nil
	case reflect.Float32, reflect.Float64:
		return "Float", nil
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		return "Int", nil
	case reflect.Ptr:
		return b.typeName(rt.Elem())
	case reflect.String:
		return "String", nil
	}

	return "", fmt.Errorf("could not find type name for %s", rt.Name())
}

func (b *builder) scanStruct(rv, parent reflect.Value, path []string) {
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
					path,
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

func (b *builder) addArg(node interface{}, field reflect.StructField, value reflect.Value, path []string) {
	if b.args == nil {
		b.args = map[interface{}]map[string]argSpec{}
	}
	if b.args[node] == nil {
		b.args[node] = map[string]argSpec{}
	}

	b.args[node][field.Name] = argSpec{
		field: field,
		value: value,
		path:  path,
	}
}

func isArgumentsField(t reflect.StructField) bool {
	if t.Name == "GraphQLArguments" {
		return true
	}

	return false
}

func (b builder) build() ([]byte, error) {
	var buf bytes.Buffer
	args, err := b.queryArguments()
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(&buf, "query%s {\n", args)
	err = b.toString(&buf, b.query, 1)
	fmt.Fprintf(&buf, "}")
	return buf.Bytes(), err
}

func (b builder) writeStructField(w io.Writer, depth int, field reflect.StructField, value reflect.Value) error {
	var (
		name      = b.nameForField(field)
		args      = b.argsStringForField(field, value)
		directive = b.directiveStringForField(field, value)
		fragment  = getTagWithPrefix(field, "...")
	)

	if directive != "" {
		directive = " " + directive
	}

	if fragment != "" {
		if field.Anonymous {
			fmt.Fprintf(w, "%s%s%s {\n", strings.Repeat(" ", depth*2), fragment, directive)
		} else {
			fmt.Fprintf(w, "%s%s%s {\n", strings.Repeat(" ", depth*2), fragment, directive)
			copyField := field
			copyField.Tag = ""
			err := b.writeStructField(w, depth+1, copyField, value)
			fmt.Fprintf(w, "%s}\n", strings.Repeat(" ", depth*2))
			return err
		}
	} else {
		fmt.Fprintf(w, "%s%s%s%s {\n", strings.Repeat(" ", depth*2), name, args, directive)
	}

	var i interface{}
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			i = reflect.New(value.Type().Elem()).Interface()
		} else {
			i = value.Interface()
		}
	} else if value.CanAddr() {
		i = value.Addr().Interface()
	} else {
		i = value.Interface()
	}

	err := b.toString(w, i, depth+1)

	fmt.Fprintf(w, "%s}\n", strings.Repeat(" ", depth*2))

	return err
}

func (b builder) nameForField(field reflect.StructField) string {
	name := field.Tag.Get("json")
	if p := strings.Index(name, ","); p != -1 {
		name = name[0:p]
	}

	if name == "" {
		name = b.toName(field.Name)
	}

	aliasOf := getTagNamed(field, "aliasof")
	if aliasOf != "" {
		name = name + ": " + aliasOf
	}

	return name
}

func (b builder) directiveStringForField(field reflect.StructField, value reflect.Value) string {
	return getTagWithPrefix(field, "@")
}

func (b builder) writeSimpleField(w io.Writer, depth int, field reflect.StructField) {
	var (
		name      = b.nameForField(field)
		args      = b.argsStringForField(field, reflect.Value{})
		directive = b.directiveStringForField(field, reflect.Value{})
		fragment  = getTagWithPrefix(field, "...")
	)
	if fragment != "" {
		fmt.Fprintf(w, "%s%s {\n", strings.Repeat(" ", depth*2), fragment)
		fmt.Fprintf(w, "%s%s%s%s\n", strings.Repeat(" ", (depth+1)*2), name, args, directive)
		fmt.Fprintf(w, "%s}\n", strings.Repeat(" ", depth*2))
	} else {
		fmt.Fprintf(w, "%s%s%s%s\n", strings.Repeat(" ", depth*2), name, args, directive)
	}
}

func (b builder) toString(w io.Writer, v interface{}, depth int) error {
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
			err := b.writeStructField(w, depth, field, value)
			if err != nil {
				return err
			}
			continue

		case isKindOf(value.Type(), reflect.Slice):
			if et := value.Type().Elem(); isKindOf(et, reflect.Struct) {
				err := b.writeStructField(w, depth, field, reflect.New(et))
				if err != nil {
					return err
				}
				continue
			}
		}

		b.writeSimpleField(w, depth, field)
	}

	return nil
}

func (b builder) toName(name string) string {
	for i, r := range name {
		if i == 0 {
			continue
		}
		if i == 1 && !unicode.IsUpper(r) {
			return strings.ToLower(name[0:i]) + name[i:]
		}
		if !unicode.IsUpper(r) {
			return strings.ToLower(name[0:i-1]) + name[i-1:]
		}
	}
	return strings.ToLower(name)
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

func (b builder) argsStringForField(field reflect.StructField, fv reflect.Value) string {
	args := ""
	if fv.CanAddr() && b.args[fv.Addr().Interface()] != nil {
		aa := []string{}
		for name, arg := range b.args[fv.Addr().Interface()] {
			// FIXME: %v is not correct, must use JSON
			aa = append(aa, fmt.Sprintf("%s: %v", b.toName(name), arg.argValue()))
		}
		sort.Strings(aa)
		args = "(" + strings.Join(aa, ", ") + ")"
	} else if tag := getTagWithPrefix(field, "("); tag != "" {
		args = tag
	}
	return args
}

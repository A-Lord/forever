package database

// The store's rows are mirrored in spelldata_store.go instead of imported from sim/core/spelldata.
// The cost of that is a mirror which can drift: a field added to the store would simply never be
// written, and the row would read as a zero the client never stated. This asserts the two sides
// name the same fields, from the store's own source rather than from a copy of it.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const storePackageDir = "../../sim/core/spelldata"

var mirroredTypes = []struct {
	storeType string
	mirror    any

	// Store fields no row states yet. A field listed here and then emitted has to leave the list,
	// which is what keeps the list from becoming a second place where drift hides.
	notYetEmitted []string

	// Mirror fields the store has no field for, because they are the generator's own bookkeeping.
	generatorOnly []string
}{
	{
		storeType: "Spell",
		mirror:    storeSpell{},
	},
	{
		storeType:     "Effect",
		mirror:        storeEffect{},
		generatorOnly: []string{"HandLinked"},
	},
	{
		storeType: "Power",
		mirror:    storePower{},
	},
}

func TestSpellStoreMirrorNamesEveryField(t *testing.T) {
	structs, err := storeStructFields()
	if err != nil {
		t.Fatalf("reading %s: %v", storePackageDir, err)
	}

	for _, m := range mirroredTypes {
		fields, ok := structs[m.storeType]
		if !ok {
			t.Errorf("%s declares no type %s - the store moved and the mirror cannot be checked",
				storePackageDir, m.storeType)
			continue
		}

		mirror := mirrorFields(m.mirror)

		var missing []string
		for _, field := range fields {
			if !mirror[field] && !containsName(m.notYetEmitted, field) {
				missing = append(missing, field)
			}
		}
		if len(missing) > 0 {
			t.Errorf("%s.%s: the generator's mirror has no field for %s - add it to %T and emit it, "+
				"or list it in notYetEmitted", storePackageDir, m.storeType, strings.Join(missing, ", "), m.mirror)
		}

		var stale []string
		for field := range mirror {
			if !containsName(fields, field) && !containsName(m.generatorOnly, field) {
				stale = append(stale, field)
			}
		}
		sort.Strings(stale)
		if len(stale) > 0 {
			t.Errorf("%T: %s is not a field of %s.%s - the store dropped or renamed it",
				m.mirror, strings.Join(stale, ", "), storePackageDir, m.storeType)
		}

		for _, field := range m.notYetEmitted {
			if !containsName(fields, field) {
				t.Errorf("%s.%s has no field %s, so listing it in notYetEmitted says nothing",
					storePackageDir, m.storeType, field)
			}
			if mirror[field] {
				t.Errorf("%T carries %s, so it is emitted and belongs out of notYetEmitted", m.mirror, field)
			}
		}
	}
}

// The exported field names of every struct the store package declares by hand. The generated files
// are skipped: they hold data, and spells_auto_gen.go is megabytes of it.
func storeStructFields() (map[string][]string, error) {
	entries, err := os.ReadDir(storePackageDir)
	if err != nil {
		return nil, err
	}

	structs := map[string][]string{}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") ||
			strings.HasSuffix(name, "_auto_gen.go") {
			continue
		}

		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(storePackageDir, name), nil, 0)
		if err != nil {
			return nil, err
		}

		ast.Inspect(file, func(n ast.Node) bool {
			spec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			structType, ok := spec.Type.(*ast.StructType)
			if !ok {
				return true
			}
			var fields []string
			for _, field := range structType.Fields.List {
				for _, ident := range field.Names {
					if ident.IsExported() {
						fields = append(fields, ident.Name)
					}
				}
			}
			structs[spec.Name.Name] = fields
			return false
		})
	}
	return structs, nil
}

func mirrorFields(v any) map[string]bool {
	fields := map[string]bool{}
	typ := reflect.TypeOf(v)
	for i := 0; i < typ.NumField(); i++ {
		if field := typ.Field(i); field.IsExported() {
			fields[field.Name] = true
		}
	}
	return fields
}

func containsName(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

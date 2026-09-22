package database

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/dbcenums"
)

// Where the aura and effect names live. They are Go constants rather than a client table, so the only
// way to read them back is to parse the source - a Go program cannot ask for a constant's name.
const dbcEnumsDir = "sim/core/dbcenums"

// The two enums a rank effect names. Their constants live in sim/core/dbcenums and are mirrored into
// sim/common/shared because that mirror is what the generated family tables read; the follow-up that
// retires those tables for the store retires the mirror with them.
var rankEnumTypes = []struct {
	dbcType    string
	sharedType string
	fieldDoc   string
}{
	{"SpellEffectType", "SpellDataEffectKind", "What the effect does - E_SCHOOL_DAMAGE, E_APPLY_AURA."},
	{"EffectAuraType", "SpellDataAura", "Which aura it applies, when the effect is E_APPLY_AURA."},
}

// Parsed rather than hand-copied, so these names cannot drift from the ones the extractor reads.
func parseDBCEnums() (map[string]map[int32]string, error) {
	entries, err := os.ReadDir(dbcEnumsDir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", dbcEnumsDir, err)
	}

	wanted := map[string]bool{}
	for _, t := range rankEnumTypes {
		wanted[t.dbcType] = true
	}

	fset := token.NewFileSet()
	out := map[string]map[int32]string{}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dbcEnumsDir, entry.Name())
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		if err := collectEnumNames(file, wanted, out); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	}
	return out, nil
}

func collectEnumNames(file *ast.File, wanted map[string]bool, out map[string]map[int32]string) error {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || vs.Type == nil || len(vs.Names) != 1 || len(vs.Values) != 1 {
				continue
			}
			ident, ok := vs.Type.(*ast.Ident)
			if !ok || !wanted[ident.Name] {
				continue
			}
			value, ok := constInt(vs.Values[0])
			if !ok {
				continue
			}
			name := vs.Names[0].Name
			if out[ident.Name] == nil {
				out[ident.Name] = map[int32]string{}
			}
			// Every constant of these types carries an explicit value and no two share one, so a
			// collision means the package changed shape and the mapping can no longer be trusted.
			if prev, dup := out[ident.Name][value]; dup {
				return fmt.Errorf("%s has two names for %d: %s and %s", ident.Name, value, prev, name)
			}
			out[ident.Name][value] = name
		}
	}
	return nil
}

func constInt(expr ast.Expr) (int32, bool) {
	if unary, ok := expr.(*ast.UnaryExpr); ok && unary.Op == token.SUB {
		v, ok := constInt(unary.X)
		return -v, ok
	}
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.INT {
		return 0, false
	}
	v, err := strconv.ParseInt(lit.Value, 0, 32)
	if err != nil {
		return 0, false
	}
	return int32(v), true
}

// The names the generated tables reference, for the values they actually use. A value the client data
// carries but dbc has not named keeps its number, so a new aura shows up as a bare int rather than
// failing the build.
type rankEnumNamer struct {
	byType map[string]map[int32]string
	used   map[string]map[int32]bool
}

func newRankEnumNamer() (*rankEnumNamer, error) {
	byType, err := parseDBCEnums()
	if err != nil {
		return nil, err
	}
	for _, t := range rankEnumTypes {
		if len(byType[t.dbcType]) == 0 {
			return nil, fmt.Errorf("no %s constants found in %s", t.dbcType, dbcEnumsDir)
		}
	}
	return &rankEnumNamer{byType: byType, used: map[string]map[int32]bool{}}, nil
}

// The expression to write for a value, recording it so the constant gets emitted.
func (n *rankEnumNamer) name(dbcType string, value int32) string {
	if _, ok := n.byType[dbcType][value]; !ok {
		return fmt.Sprintf("%d", value)
	}
	if n.used[dbcType] == nil {
		n.used[dbcType] = map[int32]bool{}
	}
	n.used[dbcType][value] = true
	return "shared." + n.byType[dbcType][value]
}

func (n *rankEnumNamer) Effect(value dbcenums.SpellEffectType) string {
	return n.name("SpellEffectType", int32(value))
}

func (n *rankEnumNamer) Aura(value dbcenums.EffectAuraType) string {
	return n.name("EffectAuraType", int32(value))
}

// The store's rows name the same two enums, through the package that declares them. A value the
// client data carries and dbcenums has not named keeps its number, marked so a reader is not left
// looking for a constant.
func (n *rankEnumNamer) storeEffect(value int32) string {
	return n.storeName("SpellEffectType", value)
}

func (n *rankEnumNamer) storeAura(value int32) string {
	return n.storeName("EffectAuraType", value)
}

func (n *rankEnumNamer) storeName(dbcType string, value int32) string {
	name, named := n.byType[dbcType][value]
	if !named {
		return fmt.Sprintf("%d /* unnamed */", value)
	}
	return "dbcenums." + name
}

func (n *rankEnumNamer) render() ([]byte, error) {
	var b strings.Builder
	b.WriteString("// Code generated by tools/database/gen_spelldata. DO NOT EDIT.\n")
	b.WriteString("//\n")
	b.WriteString("// The aura and effect names the generated spell data tables reference, mirrored from\n")
	b.WriteString("// tools/database/dbc/enums.go. Only the values the tables use are emitted, and both sides come\n")
	b.WriteString("// from the same parse, so a name here cannot drift from the one the extractor reads.\n\n")
	b.WriteString("package shared\n")

	for _, t := range rankEnumTypes {
		values := make([]int32, 0, len(n.used[t.dbcType]))
		for v := range n.used[t.dbcType] {
			values = append(values, v)
		}
		sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
		if len(values) == 0 {
			continue
		}

		fmt.Fprintf(&b, "\n// %s\n", t.fieldDoc)
		fmt.Fprintf(&b, "const (\n")
		for _, v := range values {
			fmt.Fprintf(&b, "\t%s %s = %d\n", n.byType[t.dbcType][v], t.sharedType, v)
		}
		fmt.Fprintf(&b, ")\n")
	}
	// Refuse to write rather than write something unformatted: the class files go through the same
	// gate, and an unformatted emit would show up as a spurious diff on the next run.
	return format.Source([]byte(b.String()))
}

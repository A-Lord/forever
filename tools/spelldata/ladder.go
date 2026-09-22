package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
)

// Where a spell id sits in a class file's ladders: the package, the field name the generator gave the
// family, and the call that reaches this rank of it.
type ladderRef struct {
	pkg   string
	field string
	call  string
}

func (r ladderRef) String() string {
	return fmt.Sprintf("%s spellData.%s.%s", r.pkg, r.field, r.call)
}

// One generated ladder, as the class file states it: the ranks in rank order, each with the call that
// reaches it. A talent states one rank line, since its ranks are one spell's curve.
type ladderFamily struct {
	pkg         string
	field       string
	talentRanks int32
	ranks       []familyRank
}

type familyRank struct {
	id       int32
	accessor string
}

func (f *ladderFamily) key() string {
	return f.pkg + "/" + f.field
}

var (
	loadLadders sync.Once
	ladderIndex = map[int32][]ladderRef{}
	familyIndex = map[string]*ladderFamily{}
)

// The ladder calls that reach this id, in package and field order. A class whose generated file is
// still a shared.SpellDataTable states no ladder and answers nothing, and so does a file that does
// not parse - another session editing a class package must not stop the printer.
func ladderRefs(id int32) []string {
	loadLadders.Do(scanLadders)

	refs := ladderIndex[id]
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		out = append(out, ref.String())
	}
	return out
}

// Every ladder the generated class files state, keyed `<package>/<field>`.
func ladderFamilies() map[string]*ladderFamily {
	loadLadders.Do(scanLadders)
	return familyIndex
}

// Every `spelldata.Ranked(...)` and `spelldata.Talent(...)` in the generated class files, read as
// source rather than through an import: tools/spelldata must not depend on a class package.
func scanLadders() {
	root, err := moduleRoot()
	if err != nil {
		return
	}

	files, err := filepath.Glob(filepath.Join(root, "sim", "*", "spell_data_auto_gen.go"))
	if err != nil {
		return
	}
	sort.Strings(files)

	for _, path := range files {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			continue
		}
		collectLadders(file)
	}
}

func collectLadders(file *ast.File) {
	pkg := file.Name.Name

	ast.Inspect(file, func(node ast.Node) bool {
		kv, ok := node.(*ast.KeyValueExpr)
		if !ok {
			return true
		}
		field, ok := kv.Key.(*ast.Ident)
		if !ok {
			return true
		}
		ctor, args := ladderCall(kv.Value)
		if ctor == "" {
			return true
		}

		family := newLadderFamily(pkg, field.Name, ctor, args)
		if family == nil {
			return true
		}
		familyIndex[family.key()] = family
		for _, entry := range family.refs() {
			ladderIndex[entry.id] = append(ladderIndex[entry.id], entry.ref)
		}
		return true
	})
}

// The constructor name and its integer arguments, for a value that is a call on the spelldata
// package. Anything else answers an empty name.
func ladderCall(value ast.Expr) (string, []int32) {
	call, ok := value.(*ast.CallExpr)
	if !ok {
		return "", nil
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", nil
	}
	if pkg, ok := sel.X.(*ast.Ident); !ok || pkg.Name != "spelldata" {
		return "", nil
	}

	args := make([]int32, 0, len(call.Args))
	for _, arg := range call.Args {
		lit, ok := arg.(*ast.BasicLit)
		if !ok || lit.Kind != token.INT {
			return "", nil
		}
		value, err := strconv.ParseInt(lit.Value, 0, 32)
		if err != nil {
			return "", nil
		}
		args = append(args, int32(value))
	}
	return sel.Sel.Name, args
}

type ladderEntry struct {
	id  int32
	ref ladderRef
}

// The ladder a constructor states. Ranked states one spell per rank, so the ranks are its arguments in
// order; Talent states one spell whose ranks are a curve, so the ladder is that one id reached by rank
// number. Any other call, or one whose arguments are not the integers the generator writes, is none.
func newLadderFamily(pkg, field, ctor string, args []int32) *ladderFamily {
	switch ctor {
	case "Talent":
		if len(args) != 2 || args[1] <= 0 {
			return nil
		}
		return &ladderFamily{pkg: pkg, field: field, talentRanks: args[1], ranks: []familyRank{
			{id: args[0], accessor: fmt.Sprintf("Rank(n), n up to %d", args[1])},
		}}

	case "Ranked":
		if len(args) == 0 {
			return nil
		}
		family := &ladderFamily{pkg: pkg, field: field}
		for i, id := range args {
			accessor := fmt.Sprintf("Rank(%d)", i+1)
			if i == len(args)-1 {
				accessor = "Highest()"
			}
			family.ranks = append(family.ranks, familyRank{id: id, accessor: accessor})
		}
		return family
	}
	return nil
}

// How a class file names each of the ladder's ids. A rank below the top is named by its own id rather
// than by position, which is what a reader holding that id is looking for.
func (f *ladderFamily) refs() []ladderEntry {
	out := make([]ladderEntry, 0, len(f.ranks))
	for i, rank := range f.ranks {
		call := rank.accessor
		if f.talentRanks == 0 && i < len(f.ranks)-1 {
			call = fmt.Sprintf("ByID(%d)", rank.id)
		}
		out = append(out, ladderEntry{id: rank.id, ref: ladderRef{pkg: f.pkg, field: f.field, call: call}})
	}
	return out
}

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// What -expr answers: the ladder pick alone, an effect of it, or a value read off one.
const (
	kindSpell  = "spell"
	kindEffect = "effect"
	kindValue  = "value"
)

// The identifiers a class file writes into an accessor call, with the value sim/core states for each.
// Anything else in an argument is refused rather than guessed at.
var namedValues = map[string]int64{"core.CharacterLevel": core.CharacterLevel}

// One `.Name(args)` of a chain, or a bare `.Name` where the caller wrote no call.
type segment struct {
	name string
	args []argument
	call bool
}

// An argument as the caller wrote it and as the evaluator reads it: `core.CharacterLevel` is passed as
// 60 and prints as 60 in the trail, so the trail is the chain with every name resolved.
type argument struct {
	source  string
	literal string
	value   any
}

func (s segment) text() string {
	if !s.call {
		return s.name
	}
	parts := make([]string, 0, len(s.args))
	for _, arg := range s.args {
		parts = append(parts, arg.literal)
	}
	return s.name + "(" + strings.Join(parts, ", ") + ")"
}

// A chain read to the end: the row the card states, the effect the chain went through, and the value
// the last accessor answered, with that accessor's own doc comment.
type exprResult struct {
	kind  string
	trail string

	spell      *spelldata.Spell
	readEffect int

	// The talent rank the spell is, and how many the talent has; 0 on a row that is not one.
	rank  int32
	ranks int32

	value     string
	doc       string
	accessors []string
}

// A ladder pick followed by the store's own accessors, evaluated by name over the exported method
// set: an accessor added to sim/core/spelldata is readable here without a list to keep in step.
func evalExpr(index map[string]*ladderFamily, expr, pkg string) (result *exprResult, err error) {
	// The store panics where a caller asks for a rank or an effect it does not carry, on purpose. The
	// panic is this reader's answer, as an error.
	defer func() {
		if r := recover(); r != nil {
			result, err = nil, fmt.Errorf("%v", r)
		}
	}()

	root, segments, err := parseChain(expr)
	if err != nil {
		return nil, err
	}

	head, used := headExpr(root, segments)
	picked, err := resolveExpr(index, head, pkg)
	if err != nil {
		return nil, err
	}
	s := picked.spell

	res := &exprResult{kind: kindSpell, trail: head, spell: s, value: strconv.Itoa(int(s.ID)),
		rank: picked.rank, ranks: picked.family.talentRanks, accessors: []string{}}
	current := reflect.ValueOf(s)

	for _, seg := range segments[used:] {
		answer, err := callSegment(current, seg)
		if err != nil {
			return nil, err
		}
		res.trail += "." + seg.text()
		res.doc = methodDoc(baseTypeName(current.Type()), seg.name)
		current = answer

		switch value := answer.Interface().(type) {
		case *spelldata.Spell:
			if value == spelldata.Nil {
				return nil, fmt.Errorf("%s answers a spell the store does not carry", res.trail)
			}
			if value != res.spell {
				res.rank, res.ranks = 0, 0
			}
			res.kind, res.spell, res.readEffect = kindSpell, value, 0
		case *spelldata.Effect:
			res.kind, res.readEffect = kindEffect, effectPosition(res.spell, value)
		default:
			res.kind = kindValue
		}
	}

	switch res.kind {
	case kindSpell:
		res.value = strconv.Itoa(int(res.spell.ID))
	case kindEffect:
		effect := current.Interface().(*spelldata.Effect)
		res.accessors = effectAccessors(effect)
		res.value = "no effect"
		if res.readEffect > 0 {
			res.value = fmt.Sprintf("effect %d", res.readEffect)
		}
	default:
		value, err := formatValue(current)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", res.trail, err)
		}
		res.value = value
	}
	return res, nil
}

// The chain as a name and the accessors written on it. Anything that is not a name followed by
// selectors and calls - an operator, a conversion, an index - is refused whole.
func parseChain(expr string) (string, []segment, error) {
	node, err := parser.ParseExpr(strings.TrimSpace(expr))
	if err != nil {
		return "", nil, fmt.Errorf("%q is not a chain of accessor calls", expr)
	}
	return flattenChain(node)
}

func flattenChain(node ast.Expr) (string, []segment, error) {
	switch n := node.(type) {
	case *ast.Ident:
		return n.Name, nil, nil

	case *ast.SelectorExpr:
		root, segments, err := flattenChain(n.X)
		if err != nil {
			return "", nil, err
		}
		return root, append(segments, segment{name: n.Sel.Name}), nil

	case *ast.CallExpr:
		sel, ok := n.Fun.(*ast.SelectorExpr)
		if !ok {
			return "", nil, fmt.Errorf("%s is not an accessor call", nodeText(n.Fun))
		}
		root, segments, err := flattenChain(sel.X)
		if err != nil {
			return "", nil, err
		}
		args, err := readArgs(n.Args)
		if err != nil {
			return "", nil, err
		}
		return root, append(segments, segment{name: sel.Sel.Name, args: args, call: true}), nil
	}
	return "", nil, fmt.Errorf("%s is not a chain of accessor calls: write spellData.<Family>.Highest() and the accessors on it", nodeText(node))
}

// Every argument as a number, a string or one of the names sim/core states. An argument that is
// anything else - a variable, an arithmetic expression - is refused: -expr reads, it does not run the
// package around it.
func readArgs(args []ast.Expr) ([]argument, error) {
	out := make([]argument, 0, len(args))
	for _, arg := range args {
		read, err := readArg(arg)
		if err != nil {
			return nil, err
		}
		out = append(out, read)
	}
	return out, nil
}

func readArg(arg ast.Expr) (argument, error) {
	source := nodeText(arg)

	if unary, ok := arg.(*ast.UnaryExpr); ok && unary.Op == token.SUB {
		read, err := readArg(unary.X)
		if err != nil {
			return argument{}, err
		}
		switch value := read.value.(type) {
		case int64:
			return argument{source: source, literal: "-" + read.literal, value: -value}, nil
		case float64:
			return argument{source: source, literal: "-" + read.literal, value: -value}, nil
		}
		return argument{}, fmt.Errorf("%s is not a number to negate", source)
	}

	if lit, ok := arg.(*ast.BasicLit); ok {
		switch lit.Kind {
		case token.INT:
			value, err := strconv.ParseInt(lit.Value, 0, 64)
			if err != nil {
				return argument{}, fmt.Errorf("%s is not an integer this reads", source)
			}
			return argument{source: source, literal: lit.Value, value: value}, nil
		case token.FLOAT:
			value, err := strconv.ParseFloat(lit.Value, 64)
			if err != nil {
				return argument{}, fmt.Errorf("%s is not a number this reads", source)
			}
			return argument{source: source, literal: lit.Value, value: value}, nil
		case token.STRING:
			value, err := strconv.Unquote(lit.Value)
			if err != nil {
				return argument{}, fmt.Errorf("%s is not a string this reads", source)
			}
			return argument{source: source, literal: lit.Value, value: value}, nil
		}
	}

	if value, ok := namedValues[source]; ok {
		return argument{source: source, literal: strconv.FormatInt(value, 10), value: value}, nil
	}

	names := make([]string, 0, len(namedValues))
	for name := range namedValues {
		names = append(names, name)
	}
	return argument{}, fmt.Errorf("%s is not a literal: an argument is a number, a string or %s",
		source, strings.Join(names, ", "))
}

// The ladder call at the head of the chain, in the shape resolveExpr takes, and how many segments it
// used. A chain that states no pick is handed over as it stands, so resolveExpr says what is missing.
func headExpr(root string, segments []segment) (string, int) {
	if root == "spellData" {
		switch len(segments) {
		case 0:
			return root, 0
		case 1:
			return "spellData." + segments[0].text(), 1
		}
		return "spellData." + segments[0].name + "." + segments[1].text(), 2
	}
	if len(segments) == 0 {
		return root, 0
	}
	return "spellData." + root + "." + segments[0].text(), 1
}

func callSegment(recv reflect.Value, seg segment) (reflect.Value, error) {
	owner := recv.Type().String()
	if !seg.call {
		return reflect.Value{}, fmt.Errorf("%q is not a method of %s: write the accessor as a call", seg.name, owner)
	}

	method := recv.MethodByName(seg.name)
	if !method.IsValid() {
		return reflect.Value{}, fmt.Errorf("%q is not a method of %s", seg.name, owner)
	}
	signature := method.Type()
	label := fmt.Sprintf("%s.%s", owner, seg.name)

	if signature.IsVariadic() {
		return reflect.Value{}, fmt.Errorf("%s takes a list of arguments, which -expr does not write", label)
	}
	if signature.NumIn() != len(seg.args) {
		return reflect.Value{}, fmt.Errorf("%s takes %s, not %d", label, arguments(signature.NumIn()), len(seg.args))
	}
	if signature.NumOut() != 1 {
		return reflect.Value{}, fmt.Errorf("%s answers %d values, and -expr reads one", label, signature.NumOut())
	}

	in := make([]reflect.Value, 0, len(seg.args))
	for i, arg := range seg.args {
		value, err := convertArg(arg, signature.In(i))
		if err != nil {
			return reflect.Value{}, fmt.Errorf("%s: %w", label, err)
		}
		in = append(in, value)
	}
	return method.Call(in)[0], nil
}

// A literal as the parameter's own type. An integer widens into a float, a fractional number does not
// narrow into an integer: truncating it silently is the reading a caller would not notice.
func convertArg(arg argument, want reflect.Type) (reflect.Value, error) {
	switch value := arg.value.(type) {
	case int64:
		if isInteger(want) || isFloat(want) {
			return reflect.ValueOf(value).Convert(want), nil
		}
	case float64:
		if isFloat(want) {
			return reflect.ValueOf(value).Convert(want), nil
		}
	case string:
		if want.Kind() == reflect.String {
			return reflect.ValueOf(value).Convert(want), nil
		}
	}
	return reflect.Value{}, fmt.Errorf("%s is not the %s it takes", arg.source, want)
}

func arguments(n int) string {
	if n == 1 {
		return "1 argument"
	}
	return fmt.Sprintf("%d arguments", n)
}

func isInteger(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

func isFloat(t reflect.Type) bool {
	return t.Kind() == reflect.Float32 || t.Kind() == reflect.Float64
}

// The value the last accessor answered, in the units the text form states elsewhere: a time as the
// duration it is, a number read back as the float32 the client's columns are.
func formatValue(v reflect.Value) (string, error) {
	switch value := v.Interface().(type) {
	case time.Duration:
		return value.String(), nil
	case float64:
		return number(value), nil
	case float32:
		return number(float64(value)), nil
	case bool:
		return strconv.FormatBool(value), nil
	case string:
		return value, nil
	}

	switch {
	case isInteger(v.Type()):
		return strconv.FormatInt(v.Int(), 10), nil
	case isFloat(v.Type()):
		return number(v.Float()), nil
	}
	return "", fmt.Errorf("a %s is not a value to read", v.Type())
}

func (r *exprResult) title() string {
	return rankTitle(r.spell, r.rank, r.ranks)
}

// Where an effect sits in the row the chain came through, counted the way EffectN counts. 0 for an
// effect the row does not carry, which is what a finder answers where the row has none.
func effectPosition(s *spelldata.Spell, e *spelldata.Effect) int {
	for i := range s.Effects {
		if &s.Effects[i] == e {
			return i + 1
		}
	}
	return 0
}

// Every Effect accessor that reads something off this effect, with what it answers there: the
// zero-argument ones and the ones that take a caster level. One that answers nothing - a zero amount,
// no period - is left out, so the list is the readings this effect actually states.
func effectAccessors(e *spelldata.Effect) []string {
	v := reflect.ValueOf(e)

	out := []string{}
	for i := range v.NumMethod() {
		name := v.Type().Method(i).Name
		method := v.Method(i)
		signature := method.Type()
		if signature.IsVariadic() || signature.NumOut() != 1 {
			continue
		}

		var in []reflect.Value
		label := name + "()"
		switch {
		case signature.NumIn() == 0:
		case signature.NumIn() == 1 && isInteger(signature.In(0)):
			in = []reflect.Value{reflect.ValueOf(int64(core.CharacterLevel)).Convert(signature.In(0))}
			label = fmt.Sprintf("%s(%d)", name, core.CharacterLevel)
		default:
			continue
		}

		value, err := formatValue(method.Call(in)[0])
		if err != nil || value == "" || value == "0" || value == "0s" || value == "false" {
			continue
		}
		out = append(out, label+" = "+value)
	}
	return out
}

var (
	loadDocs   sync.Once
	methodDocs = map[string]string{}
)

// The doc comment sim/core/spelldata writes on an accessor, keyed `<Type>.<Method>`, so the reader
// states what the store itself says about the value. Empty for an accessor with no comment.
func methodDoc(owner, method string) string {
	loadDocs.Do(scanMethodDocs)
	return methodDocs[owner+"."+method]
}

func scanMethodDocs() {
	root, err := moduleRoot()
	if err != nil {
		return
	}
	files, err := filepath.Glob(filepath.Join(root, "sim", "core", "spelldata", "*.go"))
	if err != nil {
		return
	}

	for _, path := range files {
		name := filepath.Base(path)
		// The generated table is one long literal and states no accessor, and a test file's helpers are
		// not what a caller reads.
		if name == "spells_auto_gen.go" || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Doc == nil || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}
			if owner := receiverTypeName(fn.Recv.List[0].Type); owner != "" {
				methodDocs[owner+"."+fn.Name.Name] = strings.TrimSpace(fn.Doc.Text())
			}
		}
	}
}

func receiverTypeName(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

// The type name the doc comments are keyed by: `*spelldata.Effect` is written on `Effect`.
func baseTypeName(t reflect.Type) string {
	name := strings.TrimPrefix(t.String(), "*")
	if _, after, ok := strings.Cut(name, "."); ok {
		return after
	}
	return name
}

func nodeText(node ast.Expr) string {
	var out strings.Builder
	if err := printer.Fprint(&out, token.NewFileSet(), node); err != nil {
		return "the expression"
	}
	return out.String()
}

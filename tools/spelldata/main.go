// Command spelldata prints one client spell row the way the store carries it, so a bare id in
// hand-written code can be read without grepping the generated table.
//
// Run from the repository root:
//
//	go run ./tools/spelldata 11574           # the row, humanised and with the client's own literal
//	go run ./tools/spelldata Whirlwind       # every row with that name, then the highest rank
//	go run ./tools/spelldata Rend -all       # every match in full
//	go run ./tools/spelldata 11574 -json     # the same as JSON, which tools/vscode-spelldata reads
//	go run ./tools/spelldata -family warrior/Execute   # the ladder's ranks, then its highest in full
//	go run ./tools/spelldata -expr 'spellData.Execute.Rank(3)' -package warrior   # the row that reaches
//	go run ./tools/spelldata -expr 'spellData.Execute.Highest().EffectN(1).Average(core.CharacterLevel)' -package warrior   # the value that reads
//	go run ./tools/spelldata -config 'spelldata.SpellConfig(&warrior.Unit, executeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))' -package warrior   # the config the resolver builds
//	go run ./tools/spelldata -hover sim/warrior/execute.go 12:40   # the markdown an editor hover shows at line:column, 1-based
//	go run ./tools/spelldata -lsp            # a language server on stdio answering those hovers
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/spelldata"
)

// An editor's language client may add its own transport flag, such as --stdio.
func isLSPInvocation(args []string) bool {
	if len(args) == 0 || strings.TrimLeft(args[0], "-") != "lsp" {
		return false
	}
	for _, arg := range args[1:] {
		if strings.TrimLeft(arg, "-") != "stdio" {
			return false
		}
	}
	return true
}

func main() {
	if isLSPInvocation(os.Args[1:]) {
		shutdown, err := serveLSP(os.Stdin, os.Stdout)
		if err != nil {
			fmt.Fprintln(os.Stderr, "spelldata:", err)
		}
		if !shutdown || err != nil {
			os.Exit(1)
		}
		return
	}
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "spelldata:", err)
		os.Exit(1)
	}
}

type options struct {
	query  string
	family string
	expr   string
	pkg    string
	hover  string
	config string
	json   bool
	all    bool
}

// The flags are taken in any position: `11574 -json` is how a caller writes it, and the stdlib flag
// package stops reading flags at the first positional argument. The ones that take a value read it as
// the next argument or after an `=`.
func parseArgs(args []string) (options, error) {
	var opts options
	valued := map[string]*string{"family": &opts.family, "expr": &opts.expr, "package": &opts.pkg, "hover": &opts.hover, "config": &opts.config}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-json", "--json":
			opts.json = true
			continue
		case "-all", "--all":
			opts.all = true
			continue
		}

		if name, value, ok := strings.Cut(strings.TrimLeft(arg, "-"), "="); ok && strings.HasPrefix(arg, "-") {
			into, known := valued[name]
			if !known {
				return opts, fmt.Errorf("unknown argument %q", arg)
			}
			*into = value
			continue
		}
		if into, known := valued[strings.TrimLeft(arg, "-")]; known && strings.HasPrefix(arg, "-") {
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s takes a value", arg)
			}
			i++
			*into = args[i]
			continue
		}

		if strings.HasPrefix(arg, "-") {
			return opts, fmt.Errorf("unknown argument %q", arg)
		}
		if opts.query != "" {
			return opts, fmt.Errorf("name one spell, not both %q and %q", opts.query, arg)
		}
		opts.query = arg
	}

	asked := 0
	for _, mode := range []string{opts.family, opts.expr, opts.hover, opts.config} {
		if mode != "" {
			asked++
		}
	}
	if opts.hover != "" && opts.query == "" {
		return opts, fmt.Errorf("-hover takes a file and a position: -hover <file> <line>:<column>")
	}
	if asked > 1 || asked == 1 && opts.query != "" && opts.hover == "" {
		return opts, fmt.Errorf("ask for one thing: a spell, -family, -expr, -config or -hover")
	}
	if opts.query == "" && asked == 0 {
		return opts, fmt.Errorf("usage: go run ./tools/spelldata <id | name> [-all] [-json] | -family <class>/<Family> | -expr <call> [-package <class>] | -config <SpellConfig call> [-package <class>] | -hover <file> <line>:<column> | -lsp")
	}
	return opts, nil
}

func run(args []string, out io.Writer) error {
	opts, err := parseArgs(args)
	if err != nil {
		return err
	}

	if opts.family != "" {
		return runFamily(out, opts)
	}
	if opts.expr != "" {
		return runExpr(out, opts)
	}
	if opts.hover != "" {
		return runHover(out, opts)
	}
	if opts.config != "" {
		return runConfig(out, opts)
	}

	if id, err := strconv.ParseInt(opts.query, 10, 32); err == nil {
		s := spelldata.Find(int32(id))
		if s == spelldata.Nil {
			return fmt.Errorf("spell %d is not in the store", id)
		}
		if opts.json {
			return writeJSON(out, asJSON(s))
		}
		writeText(out, s, 0)
		return nil
	}

	matches := spelldata.ByName(opts.query)
	if len(matches) == 0 {
		return fmt.Errorf("no spell is named %q", opts.query)
	}

	picked := matches
	if !opts.all {
		picked = []*spelldata.Spell{highestRank(matches)}
	}

	if opts.json {
		rows := make([]spellJSON, 0, len(picked))
		for _, s := range picked {
			rows = append(rows, asJSON(s))
		}
		return writeJSON(out, rows)
	}

	if len(matches) > 1 {
		for _, s := range matches {
			fmt.Fprintf(out, "%-8d %s  %s\n", s.ID, s.Name, s.Rank)
		}
		fmt.Fprintln(out)
	}
	for i, s := range picked {
		if i > 0 {
			fmt.Fprintln(out)
		}
		writeText(out, s, 0)
	}
	return nil
}

// A ladder as a whole: its rank index, then the highest rank in full, which is the rank most callers
// are after and the one a bare Highest() reaches.
func runFamily(out io.Writer, opts options) error {
	family, err := findFamily(ladderFamilies(), opts.family, opts.pkg)
	if err != nil {
		return err
	}

	highest, err := familyHighest(family)
	if err != nil {
		return err
	}

	if opts.json {
		top := asJSON(highest.spell)
		top.Title = highest.title()
		return writeJSON(out, familyJSON{
			Family:  family.key(),
			Ranks:   familyRows(family),
			Highest: top,
		})
	}

	writeFamilyText(out, family)
	writeTitledText(out, highest.title(), highest.spell, 0)
	return nil
}

func runExpr(out io.Writer, opts options) error {
	result, err := evalExpr(ladderFamilies(), opts.expr, opts.pkg)
	if err != nil {
		return err
	}

	if opts.json {
		row := asJSON(result.spell)
		row.Title = result.title()
		return writeJSON(out, exprJSON{
			spellJSON:  row,
			Expr:       opts.expr,
			Resolved:   result.spell.ID,
			Kind:       result.kind,
			Trail:      result.trail,
			Value:      result.value,
			Doc:        result.doc,
			ReadEffect: result.readEffect,
			Accessors:  result.accessors,
		})
	}

	writeExprText(out, result, opts.expr)
	return nil
}

func runConfig(out io.Writer, opts options) error {
	trace := &tracer{}
	var declarations map[string]declaration
	if root, err := moduleRoot(); err == nil && opts.pkg != "" {
		declarations = defaultWorkspace.declarations(filepath.Join(root, "sim", opts.pkg), "", "", trace)
	}
	result, err := evalSpellConfig(opts.config, declarations, opts.pkg, trace)
	if err != nil {
		return err
	}
	var text strings.Builder
	writeConfigText(&text, result)
	_, err = io.WriteString(out, text.String())
	return err
}

func runHover(out io.Writer, opts options) error {
	lineText, colText, ok := strings.Cut(opts.query, ":")
	line, lineErr := strconv.Atoi(lineText)
	col, colErr := strconv.Atoi(colText)
	if !ok || lineErr != nil || colErr != nil || line < 1 || col < 1 {
		return fmt.Errorf("%q is not a position: write <line>:<column>, both counted from 1", opts.query)
	}
	text, err := os.ReadFile(opts.hover)
	if err != nil {
		return err
	}

	markdown, trace, found := Hover(string(text), line-1, col-1, pathURI(opts.hover))
	for _, entry := range trace {
		fmt.Fprintln(os.Stderr, entry)
	}
	if !found {
		return fmt.Errorf("no hover at %s:%s", opts.hover, opts.query)
	}
	_, err = io.WriteString(out, markdown)
	return err
}

// What the chain answered, then the row it was read off. A pick states the call as the caller wrote it,
// since nothing in it was substituted; a longer chain states the trail, which is that call with every
// name resolved to the number it stands for.
func writeExprText(out io.Writer, result *exprResult, expr string) {
	switch result.kind {
	case kindSpell:
		fmt.Fprintf(out, "%s = %s\n\n", expr, result.value)
	case kindEffect:
		fmt.Fprintf(out, "%s = %s of %s\n\n", result.trail, result.value, result.title())
	default:
		fmt.Fprintf(out, "%s = %s\n\n", result.trail, result.value)
	}

	if result.doc != "" {
		for _, line := range strings.Split(result.doc, "\n") {
			fmt.Fprintln(out, strings.TrimRight("    "+line, " "))
		}
		fmt.Fprintln(out)
	}

	for i, accessor := range result.accessors {
		label := ""
		if i == 0 {
			label = "accessors"
		}
		fmt.Fprintf(out, "%-9s %s\n", label, accessor)
	}
	if len(result.accessors) > 0 {
		fmt.Fprintln(out)
	}

	writeTitledText(out, result.title(), result.spell, result.readEffect)
}

// The rank a caller means out of several rows with one name: the highest one the client states, and
// the lowest id among rows that state no rank, which is the ability itself ahead of the item and set
// rows that share its name.
func highestRank(matches []*spelldata.Spell) *spelldata.Spell {
	best := matches[0]
	for _, s := range matches[1:] {
		if s.RankNumber() > best.RankNumber() {
			best = s
		}
	}
	return best
}

func writeText(out io.Writer, s *spelldata.Spell, read int) {
	writeTitledText(out, title(s), s, read)
}

func writeTitledText(out io.Writer, heading string, s *spelldata.Spell, read int) {
	fmt.Fprintln(out, join(heading, strings.Join(ladderRefs(s.ID), "  ")))
	for _, line := range header(s) {
		fmt.Fprintln(out, line)
	}
	if proc := procSummary(s); proc != "" {
		fmt.Fprintf(out, "%-9s %s\n", "proc", proc)
	}
	if refs := refList(s); len(refs) > 0 {
		fmt.Fprintf(out, "%-9s %s\n", "refs", strings.Join(refs, ", "))
	}

	effects := effectLines(s)
	if len(effects) > 0 {
		fmt.Fprintln(out)
	}
	for i, line := range effects {
		label := fmt.Sprintf("effect %-2d", i+1)
		if i+1 == read {
			label = fmt.Sprintf("effect %d (read)", i+1)
		}
		fmt.Fprintf(out, "%s %s\n", label, line.Human)
		fmt.Fprintf(out, "%9s %s\n", "", line.Literal)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, wowheadURL(s.ID))
}

type spellJSON struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Rank string `json:"rank"`

	// The heading the text form prints, and the ladder calls a class file reaches this id through.
	Title  string   `json:"title"`
	Ladder []string `json:"ladder"`

	Header  []string `json:"header"`
	Effects []Line   `json:"effects"`

	// Empty on a row that is not a proc and on one the tooltip names no spell from.
	Proc string   `json:"proc"`
	Refs []string `json:"refs"`

	Wowhead string `json:"wowhead"`
}

// The row the chain reached, and what the chain itself answered: the kind of value, the chain with
// every name substituted, the value as it prints, the doc comment of the accessor that answered it, the
// effect the chain read and, where it stopped on an effect, the accessors that read something off it.
type exprJSON struct {
	spellJSON
	Expr     string `json:"expr"`
	Resolved int32  `json:"resolved"`

	Kind       string   `json:"kind"`
	Trail      string   `json:"trail"`
	Value      string   `json:"value"`
	Doc        string   `json:"doc"`
	ReadEffect int      `json:"read_effect"`
	Accessors  []string `json:"accessors"`
}

func asJSON(s *spelldata.Spell) spellJSON {
	effects := effectLines(s)
	if effects == nil {
		effects = []Line{}
	}
	ladder := ladderRefs(s.ID)
	if ladder == nil {
		ladder = []string{}
	}
	return spellJSON{
		ID:      s.ID,
		Name:    s.Name,
		Rank:    s.Rank,
		Title:   title(s),
		Ladder:  ladder,
		Header:  header(s),
		Effects: effects,
		Proc:    procSummary(s),
		Refs:    refList(s),
		Wowhead: wowheadURL(s.ID),
	}
}

func writeJSON(out io.Writer, v any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

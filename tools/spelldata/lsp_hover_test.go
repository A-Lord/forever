package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"
	"path/filepath"
	"strings"
	"testing"
)

const executeGo = `package warrior

var executeRank = spellData.Execute.Highest()

var executeBaseDamage = executeRank.EffectN(1).Average(core.CharacterLevel)

var executeDamagePerRage = float64(executeRank.EffectN(1).ChainAmp) * 10

func (warrior *Warrior) registerExecute() {
	baseDamage := executeBaseDamage + executeDamagePerRage*extraRage
}
`

func warriorURI(t *testing.T, name string) string {
	t.Helper()
	root, err := moduleRoot()
	if err != nil {
		t.Fatal(err)
	}
	return pathURI(filepath.Join(root, "sim", "warrior", name))
}

func positionOf(t *testing.T, text, needle string, offset int) (int, int) {
	t.Helper()
	for i, line := range strings.Split(text, "\n") {
		if at := strings.Index(line, needle); at >= 0 {
			return i, at + offset
		}
	}
	t.Fatalf("%q is not in the text", needle)
	return 0, 0
}

func hoverOn(t *testing.T, ws *workspace, text, uri, needle string, offset int) (string, []string, bool) {
	t.Helper()
	line, col := positionOf(t, text, needle, offset)
	return ws.hover(text, line, col, uri)
}

func wantHover(t *testing.T, text, needle string, offset int, wants ...string) string {
	t.Helper()
	markdown, trace, ok := hoverOn(t, newWorkspace(), text, warriorURI(t, "execute.go"), needle, offset)
	if !ok {
		t.Fatalf("no hover on %q:\n%s", needle, strings.Join(trace, "\n"))
	}
	for _, want := range wants {
		if !strings.Contains(markdown, want) {
			t.Errorf("the hover on %q lacks %q:\n%s", needle, want, markdown)
		}
	}
	return markdown
}

func TestHoverSpellIDShapes(t *testing.T) {
	cases := []struct {
		line, needle, uri string
		id                int
	}{
		{"\tvar rend = spelldata.MustFind(11574)", "MustFind(11574", "a.go", 11574},
		{"\tspelldata.Find(116)", "Find(116", "a.go", 116},
		{"\tspellData.Rend.ByID(11574)", "ByID(11574", "a.go", 11574},
		{"\tActionID: core.ActionID{SpellID: 11574},", "SpellID: 11574", "a.go", 11574},
		{"\t\t\t\"spellId\": 11574", "\"spellId\": 11574", "apl.json", 11574},
		{"\tActionId.fromSpellId(23563)", "fromSpellId(23563", "a.ts", 23563},
		{"\t\tspellId: 23563,", "spellId: 23563", "a.tsx", 23563},
	}
	for _, c := range cases {
		markdown, trace, ok := hoverOn(t, newWorkspace(), c.line, "file:///tmp/"+c.uri, c.needle, len(c.needle))
		if !ok || !strings.HasPrefix(markdown, fmt.Sprintf("### %d ", c.id)) {
			t.Errorf("%s answered %v %q:\n%s", c.line, ok, markdown, strings.Join(trace, "\n"))
		}
		if trace[1] != fmt.Sprintf("id %d", c.id) {
			t.Errorf("%s traced %v", c.line, trace)
		}
	}

	if _, _, ok := Hover("\tspell.ApplyEffects = nil", 0, 12, "file:///tmp/a.ts"); ok {
		t.Error("a TS line with no id answered a hover")
	}
	if _, _, ok := Hover("\tspelldata.Find(116) // the rank 1", 0, 30, "file:///tmp/a.ts"); ok {
		t.Error("a column outside the match answered a hover")
	}
}

// Columns are UTF-16 code units: the dash before the id is one unit and three bytes.
func TestHoverColumnIsUTF16(t *testing.T) {
	line := "// — spelldata.MustFind(11574)"
	col := len([]rune(line[:strings.Index(line, "11574")]))
	if markdown, _, ok := Hover(line, 0, col, "file:///tmp/a.go"); !ok || !strings.HasPrefix(markdown, "### 11574 ") {
		t.Errorf("the id after a dash answered %v %q", ok, markdown)
	}
}

func TestHoverFamily(t *testing.T) {
	markdown := wantHover(t, executeGo, "Execute.Highest", 3,
		"### warrior/Execute\n\n| id | name | rank | call |\n|--:|:--|:--|:--|\n",
		"| 20662 | Execute | Rank 5 | `Highest()` |",
		"### 20662 Execute · Rank 5\n")
	if !strings.Contains(markdown, "| **school** | physical |") {
		t.Errorf("the family's highest rank states no header table:\n%s", markdown)
	}
}

// The card a rank pick hovers as: a heading, the ladder call, the row's columns under its name, and the
// effects below a rule with each wording over its client row.
const executeCard = "### 20662 Execute · Rank 5\n" +
	"`warrior spellData.Execute.Highest()`  \n" +
	"\n" +
	"| | Execute (Rank 5) |\n" +
	"|--:|:--|\n" +
	"| **school** | physical |\n" +
	"| **defense** | melee |\n" +
	"| **gcd** | 1.5 s |\n" +
	"| **cost** | 15 rage |\n" +
	"| **range** | 5 yd |\n" +
	"| **stance** | Battle, Berserker |\n" +
	"| **equip** | needs a melee weapon |\n" +
	"| **attrs** | refund on miss |\n" +
	"| **labels** | 25 |\n" +
	"\n" +
	"---\n" +
	"| # | effect |\n" +
	"|--:|:--|\n" +
	"| 1 | a dummy effect holding 600 to the enemy<br>`E_DUMMY base=600 sp=1 target=[6,0]` |\n" +
	"| 2 | casts 26651 Execute to the caster<br>`E_TRIGGER_SPELL base=0 trigger=26651 target=[1,0]` |\n" +
	"\n" +
	"[Wowhead](https://www.wowhead.com/forever/spell=20662)\n"

func TestHoverRankPick(t *testing.T) {
	if got := wantHover(t, executeGo, "executeRank =", 4); got != executeCard {
		t.Errorf("the rank pick hovers as\n%s\nwant\n%s", got, executeCard)
	}
}

func TestHoverValue(t *testing.T) {
	markdown := wantHover(t, executeGo, "executeBaseDamage =", 4,
		"`executeBaseDamage` = **600**\n\n`spellData.Execute.Highest().EffectN(1).Average(60)`",
		"`spellData.Execute.Highest().EffectN(1).Average(60)`\n\n### 20662 Execute · Rank 5\n",
		"| 1 ▶ | a dummy effect holding 600 to the enemy<br>",
		"| 2 | casts 26651 Execute to the caster<br>")
	if strings.Contains(markdown, "The amount at a caster level") {
		t.Errorf("the name's hover states the accessor's doc:\n%s", markdown)
	}

	wantHover(t, executeGo, "executeBaseDamage +", 4, "`executeBaseDamage` = **600**")
}

const crueltyGo = `package warrior

func (warrior *Warrior) applyCruelty() {
	crueltyRank := spellData.Cruelty.Rank(3)
	crueltyCrit := crueltyRank.EffectN(1).BaseValue()
}
`

// A talent rank hovers as the rank Talent builds from the curve, titled with its rank.
func TestHoverTalentRank(t *testing.T) {
	wantHover(t, crueltyGo, "crueltyRank :=", 4,
		"### 12320 Cruelty · rank 3 of 5\n",
		"| | Cruelty (rank 3 of 5) |",
		"| 1 | +3% physical crit<br>`E_APPLY_AURA A_MOD_WEAPON_CRIT_PERCENT base=3 target=[1,0]` |")
	wantHover(t, crueltyGo, "crueltyCrit :=", 4,
		"`crueltyCrit` = **3**\n\n`spellData.Cruelty.Rank(3).EffectN(1).BaseValue()`",
		"### 12320 Cruelty · rank 3 of 5\n",
		"| 1 ▶ | +3% physical crit<br>")
	wantHover(t, crueltyGo, "Cruelty.Rank", 3,
		"### warrior/Cruelty\n",
		"| 12320 | Cruelty | rank 3 of 5 | `Rank(3)` | effect 1 = 3 |",
		"### 12320 Cruelty · rank 5 of 5\n")
}

func TestHoverSegments(t *testing.T) {
	wantHover(t, executeGo, "EffectN(1).Average", 2,
		"`EffectN(1)` = **effect 1** of 20662 Execute (Rank 5)",
		"`spellData.Execute.Highest().EffectN(1)`",
		"---\n| # | effect |\n|--:|:--|\n| 1 ▶ | a dummy effect holding 600 to the enemy<br>`E_DUMMY base=600 sp=1 target=[6,0]` |\n",
		"`Average(60) = 600`  \n")

	markdown := wantHover(t, executeGo, "Average(core", 2,
		"`Average(60)` = **600**",
		"`spellData.Execute.Highest().EffectN(1).Average(60)`",
		"The amount at a caster level",
		"| 1 ▶ |")
	if again := wantHover(t, executeGo, "core.CharacterLevel", 2); again != markdown {
		t.Errorf("the argument hovers differently from its call:\n%s", again)
	}

	wantHover(t, executeGo, "executeRank.EffectN(1).Average", 4, executeCard)
	wantHover(t, executeGo, "executeRank.EffectN(1).ChainAmp", 4, executeCard)
}

func TestHoverTrace(t *testing.T) {
	_, trace, _ := hoverOn(t, newWorkspace(), executeGo, warriorURI(t, "execute.go"), "executeBaseDamage =", 4)
	want := []string{
		"sim/warrior/execute.go:5:9",
		"declarations sim/warrior: cache miss",
		"ident executeBaseDamage",
		"  executeBaseDamage = executeRank.EffectN(1).Average(core.CharacterLevel)  (sim/warrior/execute.go:5)",
		"  executeRank = spellData.Execute.Highest()  (sim/warrior/execute.go:3)",
		"expr spellData.Execute.Highest().EffectN(1).Average(core.CharacterLevel)",
		"✓ 20662 Execute (Rank 5) → effect 1 → 600",
	}
	if len(trace) != len(want) {
		t.Fatalf("the trace is\n%s", strings.Join(trace, "\n"))
	}
	for i := range want {
		if !strings.HasPrefix(trace[i], want[i]) {
			t.Errorf("trace line %d is %q, want %q", i, trace[i], want[i])
		}
	}
}

func TestHoverSilent(t *testing.T) {
	cases := []struct {
		needle string
		reason string
	}{
		{"extraRage", "✗ no ladder-shaped declaration of extraRage in sim/warrior"},
		{"executeDamagePerRage =", "✗ executeDamagePerRage is bound to no chain the evaluator reads"},
		{"registerExecute", "✗ no ladder-shaped declaration of registerExecute"},
		{"func", "✗ no spell id, family or name under the cursor"},
	}
	for _, c := range cases {
		markdown, trace, ok := hoverOn(t, newWorkspace(), executeGo, warriorURI(t, "execute.go"), c.needle, 1)
		if ok {
			t.Errorf("%s answered\n%s", c.needle, markdown)
		}
		if last := trace[len(trace)-1]; !strings.HasPrefix(last, c.reason) {
			t.Errorf("%s stopped with %q, want %q", c.needle, last, c.reason)
		}
	}

	if _, _, ok := hoverOn(t, newWorkspace(), executeGo, "file:///tmp/execute.ts", "executeRank =", 2); ok {
		t.Error("a TS file answered a Go name")
	}
}

func TestDeclarations(t *testing.T) {
	file := strings.Join([]string{
		"\tsecondRank := spellData.Execute.Rank(2)",
		"\tnamed := spellData.Execute.ByID(20658)",
		"\tbare := Cruelty.Rank(3)",
		"\tif executeRank == spellData.Execute.Highest() {",
		"// var stubbedRank = spellData.MangleBear.ByID(33987)",
		"\tconfig := spelldata.SpellConfig(&warrior.Unit, executeRank, spelldata.Melee(cost))",
		"\tmaxRage := warrior.MaximumRage() - spell.Cost.GetCurrentCost()",
		"\tticks := executeRank.EffectN(1).Average(level)",
		"var executeRank = spellData.Execute.Highest() // the top rank",
	}, "\n")

	found := map[string]string{}
	for _, d := range declarationsIn(file, "a.go") {
		found[d.name] = d.chain.text(len(d.chain.segments))
	}
	want := map[string]string{
		"secondRank":  "spellData.Execute.Rank(2)",
		"named":       "spellData.Execute.ByID(20658)",
		"bare":        "Cruelty.Rank(3)",
		"executeRank": "spellData.Execute.Highest()",
	}
	if len(found) != len(want) {
		t.Errorf("declared %v, want %v", found, want)
	}
	for name, chain := range want {
		if found[name] != chain {
			t.Errorf("%s is declared as %q, want %q", name, found[name], chain)
		}
	}
}

func declared(chains map[string]string) map[string]declaration {
	out := map[string]declaration{}
	for name, text := range chains {
		c, err := parseChainAt(text, 0)
		if err != nil {
			panic(err)
		}
		out[name] = declaration{name: name, chain: c, file: "a.go", line: 1}
	}
	return out
}

func TestResolveChain(t *testing.T) {
	decls := declared(map[string]string{
		"executeRank":       "spellData.Execute.Highest()",
		"executeBaseDamage": "executeRank.EffectN(1).Average(core.CharacterLevel)",
	})
	resolves := map[string]string{
		"executeBaseDamage":      "spellData.Execute.Highest().EffectN(1).Average(core.CharacterLevel)",
		"executeRank":            "spellData.Execute.Highest()",
		"executeRank.EffectN(1)": "spellData.Execute.Highest().EffectN(1)",
		"Cruelty.Rank(3)":        "spellData.Cruelty.Rank(3)",
		"spellData.Rend.Highest().EffectN(1).Period()": "spellData.Rend.Highest().EffectN(1).Period()",
	}
	for chain, want := range resolves {
		got, err := resolveChain(chain, decls, maxSubstitutions, map[string]bool{}, &tracer{})
		if err != nil || got != want {
			t.Errorf("%s resolved to %q, %v; want %q", chain, got, err, want)
		}
	}

	refuses := map[string]string{
		"warrior.MaximumRage()":  "no ladder-shaped declaration of warrior",
		"spellData.Execute":      "names a family, not a rank",
		"executeRank.EffectN(n)": "n is not a literal",
	}
	for chain, want := range refuses {
		if got, err := resolveChain(chain, decls, maxSubstitutions, map[string]bool{}, &tracer{}); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s resolved to %q, %v; want %q", chain, got, err, want)
		}
	}
}

func TestResolveChainGuards(t *testing.T) {
	deep := declared(map[string]string{
		"a": "spellData.Execute.Highest()",
		"b": "a.EffectN(1)",
		"c": "b.Average(60)",
		"d": "c.Min(60)",
		"e": "d.Min(60)",
	})
	if got, err := resolveChain("c", deep, maxSubstitutions, map[string]bool{}, &tracer{}); got != "spellData.Execute.Highest().EffectN(1).Average(60)" {
		t.Errorf("c resolved to %q, %v", got, err)
	}
	if _, err := resolveChain("e", deep, maxSubstitutions, map[string]bool{}, &tracer{}); err == nil || !strings.Contains(err.Error(), "more than 4 names") {
		t.Errorf("e resolved past the depth limit: %v", err)
	}

	cyclic := declared(map[string]string{
		"loop":  "other.EffectN(1)",
		"other": "loop.EffectN(2)",
		"self":  "self.EffectN(1)",
	})
	for _, name := range []string{"loop", "self"} {
		if _, err := resolveChain(name, cyclic, maxSubstitutions, map[string]bool{}, &tracer{}); err == nil || !strings.Contains(err.Error(), "stands on itself") {
			t.Errorf("%s resolved through a cycle: %v", name, err)
		}
	}
}

func TestWorkspaceBuffers(t *testing.T) {
	ws := newWorkspace()
	other := warriorURI(t, "zz_hover_test_buffer.go")
	ws.setBuffer(other, "package warrior\n\nvar hoverTestRank = spellData.Rend.Highest()\n")

	use := "package warrior\n\nvar x = hoverTestRank\n"
	uri := warriorURI(t, "execute.go")
	markdown, trace, ok := hoverOn(t, ws, use, uri, "hoverTestRank", 2)
	if !ok || !strings.HasPrefix(markdown, "### 11574 Rend · Rank 7\n") {
		t.Fatalf("the buffer's declaration was not read:\n%s\n%s", markdown, strings.Join(trace, "\n"))
	}

	_, trace, _ = hoverOn(t, ws, use, uri, "hoverTestRank", 2)
	if !strings.HasSuffix(trace[1], "cache hit") {
		t.Errorf("the second hover missed the cache: %v", trace)
	}

	ws.setBuffer(other, "package warrior\n\nvar hoverTestRank = spellData.Execute.Highest()\n")
	markdown, trace, _ = hoverOn(t, ws, use, uri, "hoverTestRank", 2)
	if !strings.HasPrefix(trace[1], "declarations sim/warrior: cache miss") || !strings.HasPrefix(markdown, "### 20662 Execute") {
		t.Errorf("the edit did not reach the hover:\n%s\n%s", markdown, strings.Join(trace, "\n"))
	}

	ws.dropBuffer(other)
	if _, _, ok := hoverOn(t, ws, use, uri, "hoverTestRank", 2); ok {
		t.Error("a closed buffer's declaration is still read")
	}
}

func frame(t *testing.T, msg map[string]any) string {
	t.Helper()
	body, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)
}

func readFrames(t *testing.T, out io.Reader) []rpcMessage {
	t.Helper()
	reader := textproto.NewReader(bufio.NewReader(out))
	var msgs []rpcMessage
	for {
		body, err := readFrame(reader)
		if err == io.EOF {
			return msgs
		}
		if err != nil {
			t.Fatal(err)
		}
		var msg rpcMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			t.Fatal(err)
		}
		msgs = append(msgs, msg)
	}
}

func TestLSPRoundTrip(t *testing.T) {
	uri := warriorURI(t, "execute.go")
	line, col := positionOf(t, executeGo, "executeBaseDamage =", 4)
	position := map[string]any{"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": line, "character": col}}

	var in strings.Builder
	for _, msg := range []map[string]any{
		{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"initializationOptions": map[string]any{"trace": "on"}}},
		{"jsonrpc": "2.0", "method": "initialized", "params": map[string]any{}},
		{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": map[string]any{
			"textDocument": map[string]any{"uri": uri, "languageId": "go", "version": 1, "text": executeGo},
		}},
		{"jsonrpc": "2.0", "id": 2, "method": "textDocument/hover", "params": position},
		{"jsonrpc": "2.0", "id": 3, "method": "textDocument/definition", "params": position},
		{"jsonrpc": "2.0", "method": "$/cancelRequest", "params": map[string]any{"id": 9}},
		{"jsonrpc": "2.0", "id": 4, "method": "shutdown"},
		{"jsonrpc": "2.0", "method": "exit"},
	} {
		in.WriteString(frame(t, msg))
	}

	var out bytes.Buffer
	shutdown, err := serveLSP(strings.NewReader(in.String()), &out)
	if err != nil || !shutdown {
		t.Fatalf("the server stopped with %v, shutdown %v", err, shutdown)
	}

	responses := map[string]rpcMessage{}
	var logs []string
	for _, msg := range readFrames(t, &out) {
		if msg.Method == "window/logMessage" {
			var params struct {
				Type    int    `json:"type"`
				Message string `json:"message"`
			}
			_ = json.Unmarshal(msg.Params, &params)
			if params.Type != messageTypeLog {
				t.Errorf("a log message of type %d", params.Type)
			}
			logs = append(logs, params.Message)
			continue
		}
		responses[string(msg.ID)] = msg
	}

	var initialized struct {
		Capabilities struct {
			HoverProvider    bool `json:"hoverProvider"`
			TextDocumentSync struct {
				Change int `json:"change"`
			} `json:"textDocumentSync"`
		} `json:"capabilities"`
	}
	if err := json.Unmarshal(responses["1"].Result, &initialized); err != nil || !initialized.Capabilities.HoverProvider || initialized.Capabilities.TextDocumentSync.Change != 1 {
		t.Errorf("initialize answered %s", responses["1"].Result)
	}

	var hover struct {
		Contents struct {
			Kind  string `json:"kind"`
			Value string `json:"value"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(responses["2"].Result, &hover); err != nil {
		t.Fatalf("hover answered %s", responses["2"].Result)
	}
	if hover.Contents.Kind != "markdown" || !strings.HasPrefix(hover.Contents.Value, "`executeBaseDamage` = **600**") ||
		!strings.Contains(hover.Contents.Value, "| 1 ▶ |") {
		t.Errorf("hover answered %s:\n%s", hover.Contents.Kind, hover.Contents.Value)
	}

	if responses["3"].Error == nil || responses["3"].Error.Code != codeMethodNotFound {
		t.Errorf("an unknown method answered %+v", responses["3"])
	}
	if string(responses["4"].Result) != "null" || responses["4"].Error != nil {
		t.Errorf("shutdown answered %+v", responses["4"])
	}
	if len(logs) != 1 || !strings.Contains(logs[0], "✓ 20662 Execute (Rank 5) → effect 1 → 600") {
		t.Errorf("the trace logged %q", logs)
	}
}

func TestLSPTraceOff(t *testing.T) {
	uri := warriorURI(t, "execute.go")
	var in strings.Builder
	in.WriteString(frame(t, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"initializationOptions": map[string]any{"trace": "off"}}}))
	in.WriteString(frame(t, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "textDocument/hover", "params": map[string]any{
		"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 7, "character": 6},
	}}))

	var out bytes.Buffer
	if shutdown, err := serveLSP(strings.NewReader(in.String()), &out); err != nil || shutdown {
		t.Fatalf("the server stopped with %v, shutdown %v", err, shutdown)
	}
	for _, msg := range readFrames(t, &out) {
		if msg.Method == "window/logMessage" {
			t.Errorf("trace off still logged %s", msg.Params)
		}
	}
}

func TestLSPInvocation(t *testing.T) {
	for _, args := range [][]string{{"-lsp"}, {"--lsp"}, {"-lsp", "--stdio"}} {
		if !isLSPInvocation(args) {
			t.Errorf("%q should start the language server", args)
		}
	}
	for _, args := range [][]string{{}, {"11574"}, {"-lsp", "11574"}, {"--stdio", "-lsp"}} {
		if isLSPInvocation(args) {
			t.Errorf("%q should not start the language server", args)
		}
	}
}

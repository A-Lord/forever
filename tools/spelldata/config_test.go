package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
)

func TestCoreConstantsMatchCore(t *testing.T) {
	cases := []struct {
		typeName string
		name     string
		want     uint64
	}{
		{"SpellFlag", "SpellFlagAPL", uint64(core.SpellFlagAPL)},
		{"SpellFlag", "SpellFlagMeleeMetrics", uint64(core.SpellFlagMeleeMetrics)},
		{"SpellFlag", "SpellFlagNoOnCastComplete", uint64(core.SpellFlagNoOnCastComplete)},
		{"SpellFlag", "SpellFlagPassiveSpell", uint64(core.SpellFlagPassiveSpell)},
		{"SpellFlag", "SpellFlagHelpful", uint64(core.SpellFlagHelpful)},
		{"ProcMask", "ProcMaskMeleeMHSpecial", uint64(core.ProcMaskMeleeMHSpecial)},
		{"ProcMask", "ProcMaskMeleeOHSpecial", uint64(core.ProcMaskMeleeOHSpecial)},
		{"ProcMask", "ProcMaskMelee", uint64(core.ProcMaskMelee)},
		{"ProcMask", "ProcMaskSpellDamage", uint64(core.ProcMaskSpellDamage)},
		{"SpellSchool", "SpellSchoolPhysical", uint64(core.SpellSchoolPhysical)},
		{"SpellSchool", "SpellSchoolFrostfire", uint64(core.SpellSchoolFrostfire)},
		{"DefenseType", "DefenseTypeMelee", uint64(core.DefenseTypeMelee)},
		{"DefenseType", "DefenseTypeRanged", uint64(core.DefenseTypeRanged)},
	}
	for _, c := range cases {
		found := false
		for _, named := range coreConstants(c.typeName) {
			if named.name == c.name {
				found = true
				if named.value != c.want {
					t.Errorf("%s resolved to %d, core states %d", c.name, named.value, c.want)
				}
			}
		}
		if !found {
			t.Errorf("%s is not among the %s constants", c.name, c.typeName)
		}
	}

	if got := strings.Join(bitNames("SpellFlag", uint64(core.SpellFlagAPL|core.SpellFlagMeleeMetrics)), " | "); got != "SpellFlagMeleeMetrics | SpellFlagAPL" &&
		got != "SpellFlagAPL | SpellFlagMeleeMetrics" {
		t.Errorf("the bits read %q", got)
	}
	if got := exactName("DefenseType", uint64(core.DefenseTypeMelee)); got != "DefenseTypeMelee" {
		t.Errorf("DefenseTypeMelee reads %q", got)
	}
}

const executeConfigGo = `package warrior

var executeRank = spellData.Execute.Highest()

func (warrior *Warrior) registerExecute() {
	config := spelldata.SpellConfig(&warrior.Unit, executeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Flags(core.SpellFlagNoOnCastComplete))
	config.ClassSpellMask = SpellMaskExecute

	ww := spelldata.SpellConfig(&warrior.Unit, spellData.Whirlwind.Highest(),
		spelldata.Melee(core.ProcMaskMeleeOHSpecial), // the off hand
		spelldata.Proc(), spelldata.Tag(2))

	odd := spelldata.SpellConfig(&warrior.Unit, executeRank, spelldata.Flags(flags), spelldata.Label(3), spelldata.Flags(core.SpellFlagAPL|core.SpellFlagHelpful))
}
`

func TestHoverSpellConfig(t *testing.T) {
	markdown, trace, ok := hoverOn(t, newWorkspace(), executeConfigGo, warriorURI(t, "execute.go"), "SpellConfig(&warrior.Unit, executeRank, spelldata.Melee", 3)
	if !ok {
		t.Fatalf("no hover:\n%s", strings.Join(trace, "\n"))
	}
	for _, want := range []string{
		"`SpellConfig` of 20662 Execute (Rank 5)\n\n`spellData.Execute.Highest()`",
		"---\n| field | value | from |",
		"| **ActionID** | `SpellID 20662` | row |",
		"| **Rank** | `5` | row |",
		"| **SpellSchool** | `SpellSchoolPhysical` | row |",
		"| **DefenseType** | `DefenseTypeMelee` | row |",
		"| **ProcMask** | `ProcMaskMeleeMHSpecial` | Melee(ProcMaskMeleeMHSpecial) |",
		"| **Cast.DefaultCast.GCD** | `1.5s` | row |",
		"| **RageCost.Cost** | `15` | row |",
		"| **DamageMultiplier** | `1` | Melee(ProcMaskMeleeMHSpecial) |",
		"| **MaxRange** | `5` | row |",
		"Assignments to the config after the call are not folded in.",
	} {
		if !strings.Contains(markdown, want) {
			t.Errorf("the hover lacks %q:\n%s", want, markdown)
		}
	}

	flags := ""
	for _, line := range strings.Split(markdown, "\n") {
		if strings.HasPrefix(line, "| **Flags** |") {
			flags = line
		}
	}
	for _, want := range []string{"SpellFlagMeleeMetrics", "SpellFlagAPL", "SpellFlagNoOnCastComplete", "Melee(ProcMaskMeleeMHSpecial), Flags(SpellFlagNoOnCastComplete)"} {
		if !strings.Contains(flags, want) {
			t.Errorf("the Flags row %q lacks %q", flags, want)
		}
	}

	wantTrace := []string{"SpellConfig", "  row executeRank", "  executeRank = spellData.Execute.Highest()", "  option Melee(ProcMaskMeleeMHSpecial)", "✓ 20662 Execute (Rank 5) → "}
	joined := strings.Join(trace, "\n")
	for _, want := range wantTrace {
		if !strings.Contains(joined, want) {
			t.Errorf("the trace lacks %q:\n%s", want, joined)
		}
	}
}

func TestHoverSpellConfigAcrossLines(t *testing.T) {
	markdown, trace, ok := hoverOn(t, newWorkspace(), executeConfigGo, warriorURI(t, "execute.go"), "SpellConfig(&warrior.Unit, spellData.Whirlwind", 3)
	if !ok {
		t.Fatalf("no hover:\n%s", strings.Join(trace, "\n"))
	}
	for _, want := range []string{
		"| **ActionID** | `SpellID 1680, Tag 2` | row, Tag(2) |",
		"| **ProcMask** | `ProcMaskMeleeOHSpecial` | Melee(ProcMaskMeleeOHSpecial) |",
		"SpellFlagPassiveSpell",
		"Proc()",
	} {
		if !strings.Contains(markdown, want) {
			t.Errorf("the hover lacks %q:\n%s", want, markdown)
		}
	}
	if strings.Contains(markdown, "RageCost") || strings.Contains(markdown, "Cast.DefaultCast.GCD") {
		t.Errorf("Proc() left a cost or a cast behind:\n%s", markdown)
	}
}

func TestHoverSpellConfigUnevaluated(t *testing.T) {
	markdown, trace, ok := hoverOn(t, newWorkspace(), executeConfigGo, warriorURI(t, "execute.go"), "SpellConfig(&warrior.Unit, executeRank, spelldata.Flags(flags)", 3)
	if !ok {
		t.Fatalf("no hover:\n%s", strings.Join(trace, "\n"))
	}
	for _, want := range []string{
		"unevaluated: `spelldata.Flags(flags)`",
		"unevaluated: `spelldata.Label(3)` (spelldata.Label is not an option this reads)",
		"Flags(SpellFlagAPL \\| SpellFlagHelpful)",
	} {
		if !strings.Contains(markdown, want) {
			t.Errorf("the hover lacks %q:\n%s", want, markdown)
		}
	}
}

func TestConfigText(t *testing.T) {
	var out bytes.Buffer
	call := "spelldata.SpellConfig(&warrior.Unit, executeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))"
	if err := run([]string{"-config", call, "-package", "warrior"}, &out); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"SpellConfig of 20662 Execute (Rank 5)\n", "ProcMaskMeleeMHSpecial", "Melee(ProcMaskMeleeMHSpecial)", configFootnote} {
		if !strings.Contains(text, want) {
			t.Errorf("-config lacks %q:\n%s", want, text)
		}
	}
}

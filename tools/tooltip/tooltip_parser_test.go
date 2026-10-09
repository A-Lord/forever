package tooltip

import (
	"fmt"
	"os"
	"testing"

	"github.com/wowsims/forever/tools/database/dbc"
)

var db *dbc.DBC

// The game files inputs load from paths relative to the repository root.
func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err := os.Stat("./assets/db_inputs/dbc/spells.json"); err != nil {
		fmt.Println("no game files DBC inputs in assets/db_inputs/dbc - run `make db` from a local WoW install to enable these tests")
		os.Exit(0)
	}
	db = dbc.GetDBC()
	os.Exit(m.Run())
}

func Test_WhenInvalidTernaryGiven_ThenProperlyApplyFixes(t *testing.T) {
	tp, error := ParseTooltip("Damage every ${$16914d3/10}.2 seconds$?$w1!=0[ and movement slowed by $w1%][].",
		NewTestDataProvider(CharacterConfig{SpellDamage: 1000}),
		16914,
	)

	if error != nil {
		t.Fatal()
	}

	if tp.String() != "Damage every 1.00 seconds." {
		t.Fail()
	}
}

func Test_WhenLConditionGiven_ThenProperlyEvaluate(t *testing.T) {
	SimpleTooltipTest(370, "Purges the enemy target, removing 1 magic effect.", t)
}

func Test_WhenGivenDurationShorter60_ThenRenderSeconds(t *testing.T) {
	SimpleTooltipTest(56, "Stuns target for 3s.", t)
}

func Test_WhenGivenDurationLongerThan2Hours_ThenRnderHrs(t *testing.T) {
	SimpleTooltipTest(1289654,
		"Increases mounted movement speed by 10% for 360hrs, stacking with all other mounted movement speed bonuses. This expires while offline.",
		t,
	)
}

// Uppercase $H is the proc chance as much as lowercase $h: Invulnerable Mail's 16620 states 5.
func Test_WhenUppercaseProcChanceGiven_ThenRenderTheChance(t *testing.T) {
	GeneratorTooltipTest(16620,
		"When struck in combat has a 5% chance to make you invulnerable to melee damage for 3s. This effect can only occur once every 10 sec.",
		t,
	)
}

// A reference into another spell renders the amount that spell's row states, with no spell power on
// top: Arcane Infused Gem's 23721 names Arcane Detonation 23722's 200, which carries a coefficient of 1.
func Test_WhenOtherSpellsDamageReferenced_ThenRenderItsStatedAmount(t *testing.T) {
	GeneratorTooltipTest(23721,
		"Infuses you with Arcane energy, causing your next Arcane Shot fired within 10s to detonate at the target. The Arcane Detonation will deal 200 damage to enemies near the target.",
		t,
	)
}

// A computed value is unsigned like a plain one, since the tooltip writes its own words for a reduction:
// 22778 states -30 tenths of rage as "reduced by $/10;s1".
func Test_WhenNegativeValueComputed_ThenRenderItUnsigned(t *testing.T) {
	GeneratorTooltipTest(22778, "Hamstring Rage cost reduced by 3.0.", t)
}

func SimpleTooltipTest(spellId int, expectedDescription string, t *testing.T) {
	renderTooltipTest(spellId, NewTestDataProvider(CharacterConfig{}), expectedDescription, t)
}

// The tooltip as the item generator renders it, through the game files' data provider alone.
func GeneratorTooltipTest(spellId int, expectedDescription string, t *testing.T) {
	renderTooltipTest(spellId, DBCTooltipDataProvider{DBC: db}, expectedDescription, t)
}

func renderTooltipTest(spellId int, dataProvider TooltipDataProvider, expectedDescription string, t *testing.T) {
	t.Helper()
	tp, error := ParseTooltip(db.Spells[spellId].Description, dataProvider, int64(spellId))

	if error != nil {
		t.Fatal(error)
	}

	if got := tp.String(); got != expectedDescription {
		t.Errorf("%d renders %q, want %q", spellId, got, expectedDescription)
	}
}

func NewTestDataProvider(config CharacterConfig) *TestDataProvider {
	return &TestDataProvider{
		DBCTooltipDataProvider: &DBCTooltipDataProvider{
			DBC: db,
		},
		Character: &config,
	}
}

// add here over time to overwrite fixed values for tests
type CharacterConfig struct {
	SpellDamage float64
}

type TestDataProvider struct {
	*DBCTooltipDataProvider
	Character *CharacterConfig
}

func (t TestDataProvider) GetSpellDamage() float64 {
	return t.Character.SpellDamage
}

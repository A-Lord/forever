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
		"Increases mounted movement speed by 10%, stacking with all other mounted movement speed bonuses. Lasts 360hrs real time.\r\n",
		t,
	)
}

func SimpleTooltipTest(spellId int, expectedDescription string, t *testing.T) {
	spell := db.Spells[spellId]
	tp, error := ParseTooltip(spell.Description,
		NewTestDataProvider(CharacterConfig{}),
		int64(spellId),
	)

	if error != nil {
		t.Fatal()
	}

	if tp.String() != expectedDescription {
		t.Fail()
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

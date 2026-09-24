package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	smokedDesertDumplingsID int32 = 20452
	elixirOfTheMongooseID   int32 = 13452
	scrollOfIntellectIVID   int32 = 10308
	swiftfeatherTokenID     int32 = 1400
)

func registerStatConsumable(t *testing.T, id int32) {
	t.Helper()
	if core.GetConsumableByID(id).Id == 0 {
		t.Skipf("consumable %d is not in the loaded database", id)
	}
	if !core.HasConsumableEffect(id) {
		NewSpellDataStatConsumable(id)
	}
}

func newConsumablesSim(consumes *proto.ConsumesSpec, raidBuffs *proto.RaidBuffs) (*core.Simulation, *core.Character) {
	player := &proto.Player{
		Name:        "Caster",
		Class:       proto.Class_ClassShaman,
		Race:        proto.Race_RaceTroll,
		Buffs:       &proto.IndividualBuffs{},
		Spec:        &proto.Player_ElementalShaman{},
		Equipment:   &proto.EquipmentSpec{},
		Consumables: consumes,
	}

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}},
			Buffs:   raidBuffs,
		},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 60}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

func statGain(consumes *proto.ConsumesSpec, raidBuffs *proto.RaidBuffs) stats.Stats {
	_, without := newConsumablesSim(&proto.ConsumesSpec{}, raidBuffs)
	_, with := newConsumablesSim(consumes, raidBuffs)
	return with.GetStats().Subtract(without.GetStats())
}

// Nutritious Food 1248401 hands its 20 to Well Fed 1248422, whose own row states 1 Strength. Well
// Fed persists through death, so it is up for the whole iteration.
func TestWellFedTakesTheValueItsTriggerHandsOver(t *testing.T) {
	registerStatConsumable(t, smokedDesertDumplingsID)

	gain := statGain(&proto.ConsumesSpec{FoodId: smokedDesertDumplingsID}, &proto.RaidBuffs{})
	if gain[stats.Strength] != 20 {
		t.Errorf("Smoked Desert Dumplings add %v Strength, want 20", gain[stats.Strength])
	}

	_, character := newConsumablesSim(&proto.ConsumesSpec{FoodId: smokedDesertDumplingsID}, &proto.RaidBuffs{})
	aura := character.GetAuraByID(core.ActionID{ItemID: smokedDesertDumplingsID})
	if aura == nil || !aura.IsActive() || aura.Duration != core.NeverExpires {
		t.Errorf("Well Fed is %+v, want up for the whole iteration", aura)
	}
}

// 17538 states 25 Agility and 2% critical strike chance for 30 minutes; the flat stats the catalogue
// carries name only the Agility.
func TestElixirTakesEveryEffectAndTheDurationOfItsRow(t *testing.T) {
	registerStatConsumable(t, elixirOfTheMongooseID)

	gain := statGain(&proto.ConsumesSpec{BattleElixirId: elixirOfTheMongooseID}, &proto.RaidBuffs{})
	if gain[stats.Agility] != 25 || gain[stats.PhysicalCritPercent] < 2-1e-9 {
		t.Errorf("Elixir of the Mongoose adds %v Agility and %v%% crit, want 25 and at least 2",
			gain[stats.Agility], gain[stats.PhysicalCritPercent])
	}

	_, character := newConsumablesSim(&proto.ConsumesSpec{BattleElixirId: elixirOfTheMongooseID}, &proto.RaidBuffs{})
	aura := character.GetAuraByID(core.ActionID{ItemID: elixirOfTheMongooseID})
	if aura == nil || !aura.IsActive() || aura.Duration != 30*time.Minute {
		t.Errorf("Elixir of the Mongoose is %+v, want up from the pull for 30 minutes", aura)
	}
}

// Scroll of Intellect IV states 16 Intellect, which Arcane Brilliance's 40 overrides rather than adds
// to.
func TestScrollIsExclusiveWithTheRaidBuffOfItsStat(t *testing.T) {
	registerStatConsumable(t, scrollOfIntellectIVID)

	if gain := statGain(&proto.ConsumesSpec{ExtraIds: []int32{scrollOfIntellectIVID}}, &proto.RaidBuffs{}); gain[stats.Intellect] != 16 {
		t.Errorf("Scroll of Intellect IV adds %v Intellect, want 16", gain[stats.Intellect])
	}

	brilliance := &proto.RaidBuffs{ArcaneBrilliance: true}
	if gain := statGain(&proto.ConsumesSpec{ExtraIds: []int32{scrollOfIntellectIVID}}, brilliance); gain[stats.Intellect] != 0 {
		t.Errorf("Scroll of Intellect IV adds %v Intellect under Arcane Brilliance, want 0", gain[stats.Intellect])
	}
}

// Swiftfeather Token's buff lasts 30 s against the 5 min category cooldown the item shares, so it is
// used in combat rather than put up before the pull.
func TestShortBuffIsUsedOnItsCooldown(t *testing.T) {
	registerStatConsumable(t, swiftfeatherTokenID)

	_, character := newConsumablesSim(&proto.ConsumesSpec{FoodId: swiftfeatherTokenID}, &proto.RaidBuffs{})
	mcd := character.GetMajorCooldown(core.ActionID{ItemID: swiftfeatherTokenID})
	if mcd == nil {
		t.Fatal("Swiftfeather Token registers no major cooldown")
	}
	if mcd.Spell.SharedCD.Duration != 5*time.Minute {
		t.Errorf("Swiftfeather Token shares a %v cooldown, want 5m", mcd.Spell.SharedCD.Duration)
	}

	aura := character.GetAuraByID(core.ActionID{ItemID: swiftfeatherTokenID})
	if aura == nil || aura.IsActive() || aura.Duration != 30*time.Second {
		t.Errorf("Swiftfeather Token's buff is %+v, want down at the pull and lasting 30s", aura)
	}
}

// R.O.I.D.S. 8410 (10667, 25 Strength) and Ground Scorpok Assay 8412 (10669, 25 Agility) share
// category 103's 1 h, as long as each buff lasts, so only the first the spec names is up. Spirit of
// Zanza 20079 (24382, 50 Spirit and 50 Stamina, category 79's 3 s) and Lung Juice Cocktail 8411
// (10668, 25 Stamina) state nothing that keeps them apart, so both are up.
func TestOnlyACategoryCooldownAsLongAsTheBuffKeepsConsumablesApart(t *testing.T) {
	const roids, scorpok, zanza, lungJuice int32 = 8410, 8412, 20079, 8411
	for _, id := range []int32{roids, scorpok, zanza, lungJuice} {
		registerStatConsumable(t, id)
	}

	for _, tc := range []struct {
		name  string
		ids   []int32
		stat  stats.Stat
		gain  float64
		other stats.Stat
	}{
		{"R.O.I.D.S. first", []int32{roids, scorpok}, stats.Strength, 25, stats.Agility},
		{"Ground Scorpok Assay first", []int32{scorpok, roids}, stats.Agility, 25, stats.Strength},
	} {
		gain := statGain(&proto.ConsumesSpec{ExtraIds: tc.ids}, &proto.RaidBuffs{})
		if gain[tc.stat] != tc.gain || gain[tc.other] != 0 {
			t.Errorf("%s: %s rose by %v and %s by %v, want %v and 0", tc.name, tc.stat.StatName(), gain[tc.stat],
				tc.other.StatName(), gain[tc.other], tc.gain)
		}
	}

	if gain := statGain(&proto.ConsumesSpec{ExtraIds: []int32{zanza, lungJuice}}, &proto.RaidBuffs{}); gain[stats.Stamina] != 75 {
		t.Errorf("Spirit of Zanza and Lung Juice Cocktail add %v Stamina, want 75", gain[stats.Stamina])
	}
}

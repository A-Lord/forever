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

	if gain := statGain(&proto.ConsumesSpec{ScrollInt: true}, &proto.RaidBuffs{}); gain[stats.Intellect] != 16 {
		t.Errorf("Scroll of Intellect IV adds %v Intellect, want 16", gain[stats.Intellect])
	}

	brilliance := &proto.RaidBuffs{ArcaneBrilliance: true}
	if gain := statGain(&proto.ConsumesSpec{ScrollInt: true}, brilliance); gain[stats.Intellect] != 0 {
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

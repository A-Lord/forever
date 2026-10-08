//go:build with_db

package forever_test

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/core/testagent"
)

func init() {
	testagent.Register(proto.Player_HolyPaladin{}, proto.Spec_SpecHolyPaladin, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_HolyPaladin)
	}, func(agent *testagent.Agent) {
		agent.EnableManaBar()
	})
}

// A character wearing one item, against a level 60 target: the warrior stand-in, with a paladin the
// stand-in registered as the holy paladin, which has a mana bar, or with a rogue the stand-in
// registered as the rogue, which shoots its ranged weapon.
func itemWearer(t *testing.T, class proto.Class, slot proto.ItemSlot, itemID int32) (*core.Simulation, *core.Character) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[slot].Id = itemID
	player := &proto.Player{
		Name: "Wearer", Class: class, Race: proto.Race_RaceHuman,
		Equipment:   &proto.EquipmentSpec{Items: items},
		Consumables: &proto.ConsumesSpec{},
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{},
	}
	switch class {
	case proto.Class_ClassPaladin:
		core.WithSpec(player, &proto.Player_HolyPaladin{HolyPaladin: &proto.HolyPaladin{}})
	case proto.Class_ClassRogue:
		core.WithSpec(player, &proto.Player_Rogue{Rogue: &proto.Rogue{}})
	default:
		core.WithSpec(player, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}})
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Whether the damage is the roll's range, or that range at the spell crit multiplier on a crit.
func damageInRange(metrics core.SpellMetrics, low, high float64) bool {
	if metrics.Crits > 0 {
		low, high = low*1.5, high*1.5
	}
	return metrics.TotalDamage >= low && metrics.TotalDamage <= high
}

// Aegis of the Scarlet Commander 7726's 1293539 deals 14 Holy (13 to 15) to the attacker of a melee
// hit the wearer blocks, at most once every 5 s (the row's RecoveryTime); a hit not blocked deals
// nothing.
func TestAegisOfTheScarletCommanderDealsHolyDamageOnBlock(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotOffHand, 7726)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	enemy := sim.Encounter.ActiveTargetUnits[0]
	retribution := character.GetSpell(core.ActionID{SpellID: 1293539})
	if retribution == nil {
		t.Fatal("7726 registers no damage spell")
	}

	takeHit(t, sim, character, meleeSwing(enemy))
	if got := retribution.SpellMetrics[enemy.UnitIndex].Casts; got != 0 {
		t.Errorf("a hit not blocked cast the damage %d times, want none", got)
	}

	block := func() {
		character.OnSpellHitTaken(sim, meleeSwing(enemy), &core.SpellResult{Target: &character.Unit, Outcome: core.OutcomeHit | core.OutcomeBlock, Damage: 50})
		stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	}
	firstBlock := sim.CurrentTime
	block()
	metrics := retribution.SpellMetrics[enemy.UnitIndex]
	if metrics.Casts != 1 || !damageInRange(metrics, 13, 15) {
		t.Errorf("a block cast the damage %d times for %v, want once for 13 to 15", metrics.Casts, metrics.TotalDamage)
	}
	block()
	if got := retribution.SpellMetrics[enemy.UnitIndex].Casts; got != 1 {
		t.Errorf("a second block at once cast the damage %d times in all, want once still", got)
	}
	stepTo(t, sim, firstBlock+5*time.Second)
	block()
	if got := retribution.SpellMetrics[enemy.UnitIndex].Casts; got != 2 {
		t.Errorf("a block 5 s after the first cast the damage %d times in all, want 2", got)
	}
}

// Black Grasp of the Destroyer 22194 restores 8 mana (29471) on each landed melee or ranged hit, and
// nothing on a spell; a wearer without mana registers nothing.
func TestBlackGraspOfTheDestroyerRestoresManaPerHit(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassPaladin, proto.ItemSlot_ItemSlotHands, 22194)
	character.SpendMana(sim, 100, character.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen}))
	before := character.CurrentMana()

	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if got := character.CurrentMana() - before; got != 8 {
		t.Errorf("a melee hit restored %v mana, want 8", got)
	}
	dealHit(t, sim, character, fireSpell(&character.Unit))
	if got := character.CurrentMana() - before; got != 8 {
		t.Errorf("after a spell hit the mana restored is %v, want 8 still", got)
	}

	_, warrior := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotHands, 22194)
	if warrior.GetAura("Black Grasp of the Destroyer") != nil {
		t.Errorf("a wearer without mana registers the proc")
	}
}

// Thick Obsidian Breastplate 22196's 27539 has a 30% chance, on a spell hit taken, to shield the
// wearer for 6 s against that spell's school: 400 (300 to 500) absorbed, the Fire one 27533's. The
// chance has a 10 s internal cooldown, and a Physical spell, which has no shield, does not spend it.
func TestThickObsidianBreastplateShieldsTheSchoolHit(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotChest, 22196)
	enemy := sim.Encounter.ActiveTargetUnits[0]
	fireShield := character.GetAuraByID(core.ActionID{SpellID: 27533})
	frostShield := character.GetAuraByID(core.ActionID{SpellID: 27534})
	if fireShield == nil || frostShield == nil {
		t.Fatal("22196 registers no Fire or Frost shield")
	}
	spell := func(school core.SpellSchool) *core.Spell {
		return &core.Spell{ProcMask: core.ProcMaskSpellDamage, SpellSchool: school, DefenseType: core.DefenseTypeMagic, Unit: enemy}
	}
	shieldsUp := func() int {
		up := 0
		for _, aura := range character.GetAuras() {
			if aura.IsActive() && aura.ActionID.SpellID >= 27533 && aura.ActionID.SpellID <= 27540 {
				up++
			}
		}
		return up
	}
	hitUntil := func(school core.SpellSchool, shield *core.Aura) {
		t.Helper()
		for range 200 {
			takeHit(t, sim, character, spell(school))
			if shield.IsActive() {
				return
			}
		}
		t.Fatalf("200 %v spells taken raised no %v shield", school, school)
	}

	for range 100 {
		takeHit(t, sim, character, spell(core.SpellSchoolPhysical))
		takeHit(t, sim, character, meleeSwing(enemy))
	}
	if got := shieldsUp(); got != 0 {
		t.Fatalf("Physical spells and melee hits raised %d shields, want none", got)
	}

	hitUntil(core.SpellSchoolFire, fireShield)
	raised := sim.CurrentTime
	if got := fireShield.GetStacks(); got < 300 || got > 500 {
		t.Errorf("the Fire shield absorbs %d, want 300 to 500", got)
	}
	if got := fireShield.Duration; got != 6*time.Second {
		t.Errorf("the Fire shield lasts %v, want 6s", got)
	}
	stepTo(t, sim, raised+6*time.Second+time.Millisecond)
	if fireShield.IsActive() {
		t.Errorf("the Fire shield is still up after 6 s")
	}

	for range 100 {
		takeHit(t, sim, character, spell(core.SpellSchoolFrost))
	}
	if sim.CurrentTime >= raised+10*time.Second || frostShield.IsActive() {
		t.Errorf("a Frost shield went up within the 10 s cooldown")
	}

	stepTo(t, sim, raised+10*time.Second)
	hitUntil(core.SpellSchoolFrost, frostShield)
	if got := shieldsUp(); got != 1 {
		t.Errorf("%d shields are up after a Frost spell, want the Frost one alone", got)
	}
}

// Demon Forged Breastplate 12628's 16611 has a 3% chance, on a melee hit taken, to cast 16608 on the
// attacker: 60 Shadow a second for 4 s, each tick healing the wearer for the damage it dealt.
func TestDemonForgedBreastplateDrainsTheAttacker(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotChest, 12628)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	character.RemoveHealth(sim, 1000)
	enemy := sim.Encounter.ActiveTargetUnits[0]
	drain := character.GetSpell(core.ActionID{SpellID: 16608})
	if drain == nil {
		t.Fatal("12628 registers no drain")
	}

	hits := 0
	for ; !drain.Dot(enemy).IsActive() && hits < 1000; hits++ {
		takeHit(t, sim, character, meleeSwing(enemy))
	}
	if hits == 1000 {
		t.Fatalf("1000 melee hits taken cast no drain")
	}
	dot := drain.Dot(enemy)
	if dot.BaseTickCount != 4 || dot.BaseTickLength != time.Second {
		t.Errorf("the drain ticks %d times every %v, want 4 every 1s", dot.BaseTickCount, dot.BaseTickLength)
	}

	health := character.CurrentHealth()
	stepTo(t, sim, sim.CurrentTime+4*time.Second+time.Millisecond)
	metrics := drain.SpellMetrics[enemy.UnitIndex]
	if metrics.TotalDamage != 240 {
		t.Errorf("the drain dealt %v, want 240", metrics.TotalDamage)
	}
	if got := character.CurrentHealth() - health; got != metrics.TotalDamage {
		t.Errorf("the drain healed the wearer for %v, want the %v it dealt", got, metrics.TotalDamage)
	}
}

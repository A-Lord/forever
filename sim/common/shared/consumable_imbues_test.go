package shared

import (
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// A test oil shaped like Brilliant Wizard Oil 20749, enchant 2628, carrying 36 healing, 36 spell
// damage and 1% spell crit: its stats apply once for each hand carrying it.
func TestTemporaryEnchantStatsApplyPerEnchantedHand(t *testing.T) {
	oil := core.Consumable{Id: 20749, Name: "Brilliant Wizard Oil", TemporaryEnchant: &core.Enchant{
		EffectID:    2628,
		Stats:       stats.Stats{stats.HealingPower: 36, stats.SpellDamage: 36},
		PseudoStats: pseudoStats(proto.PseudoStat_PseudoStatSpellCritPercent, 1),
	}}

	character := newTestAgent().character
	character.Equipment[proto.ItemSlot_ItemSlotMainHand] = core.Item{ID: 1, TempEnchant: 2628}
	before := character.GetStats()
	applyTemporaryEnchant(character, oil)
	gained := character.GetStats().Subtract(before)

	if gained[stats.SpellDamage] != 36 || gained[stats.HealingPower] != 36 || gained[stats.SpellCritPercent] != 1 {
		t.Errorf("one oiled hand gained %v spell damage, %v healing, %v%% spell crit, want 36, 36 and 1",
			gained[stats.SpellDamage], gained[stats.HealingPower], gained[stats.SpellCritPercent])
	}

	bare := newTestAgent().character
	before = bare.GetStats()
	applyTemporaryEnchant(bare, oil)
	if gained := bare.GetStats().Subtract(before); gained != (stats.Stats{}) {
		t.Errorf("a character with no hand carrying the enchant gained %v", gained)
	}
}

// Shadow Oil 3824, enchant 25: its combat spell 350026 is cast off the enchanted weapon's hits, at the
// chance the store writes onto 350026's row from the enchantment, with the 10 s internal cooldown the
// override states.
func TestTemporaryEnchantCombatSpellRegistersAWeaponProc(t *testing.T) {
	shadowOil := core.Consumable{Id: 3824, Name: "Shadow Oil", TemporaryEnchant: &core.Enchant{
		EffectID:       25,
		EnchantEffects: []*proto.ItemEffect{{BuffId: 350026}},
	}}

	character := newTestAgent().character
	character.Equipment[proto.ItemSlot_ItemSlotMainHand] = core.Item{ID: 1, TempEnchant: 25}
	applyTemporaryEnchant(character, shadowOil)

	listener := auraByLabel(character, "Shadow Oil")
	if listener == nil {
		t.Fatal("no listener casts the oil's combat spell")
	}
	if listener.Icd == nil || listener.Icd.Duration != 10*time.Second {
		t.Errorf("the oil's listener has internal cooldown %v, want 10 s", listener.Icd)
	}
	if character.GetSpell(core.ActionID{SpellID: 350026}) == nil {
		t.Error("the combat spell 350026 was not registered")
	}
	if chance := spelldata.MustFind(350026).StatedChance(); chance <= 0 || chance >= 1 {
		t.Errorf("350026 rolls %v, want the share of a hit the store writes from enchant 25", chance)
	}
}

// Scroll of Imbue Manablade's enchant 8715 names 1308913, a chance on the wearer's melee swings and
// abilities with a ProcCategoryRecovery, firing 1306613's E_ENERGIZE mana; Scroll of Imbue Spark's
// 8707 names 1307644, a chance on the melee hits the wearer takes with a lockout of its own, firing
// 1306610's E_MODIFY_COOLDOWNS_BY_CATEGORY, a reduction of one category's cooldown.
const (
	manabladeProc int32 = 1308913
	sparkProc     int32 = 1307644
)

// Whether procs out of rolls lie within four standard deviations of what the chance expects, with at
// least one landed for the checks that follow each proc.
func plausibleProcs(procs int32, rolls int, chance float64) bool {
	mean := float64(rolls) * chance
	return procs > 0 && math.Abs(float64(procs)-mean) <= 4*math.Sqrt(mean*(1-chance))
}

// The spell a proc row fires.
func firedSpell(procRow int32) *spelldata.Spell {
	return spelldata.MustFind(procRow).FirstAura(dbcenums.A_PROC_TRIGGER_SPELL).Trigger()
}

// A player of the given class wearing two test one-handers, the main hand imbued with a test
// consumable carrying the enchant whose equip spell is the proc.
func newImbuedSim(t *testing.T, player *proto.Player, consumableID, enchantID, procSpellID int32) (*core.Simulation, *testCaster) {
	t.Helper()
	const mainHandID, offHandID int32 = 991501, 991502
	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{testOneHander(mainHandID), testOneHander(offHandID)},
		Consumables: []*proto.Consumable{{Id: consumableID, Name: "Test Imbue", Type: proto.ConsumableType_ConsumableTypeImbue,
			TemporaryEnchant: &proto.SimEnchant{EffectId: enchantID}}},
	})
	NewSpellDataTemporaryEnchant(consumableID, SchoolPiercing{}, procSpellID)

	player.Name = "Wearer"
	player.Buffs = &proto.IndividualBuffs{}
	player.Equipment = &proto.EquipmentSpec{Items: testHands(&proto.ItemSpec{Id: mainHandID}, &proto.ItemSpec{Id: offHandID})}
	player.Consumables = &proto.ConsumesSpec{MhImbueId: consumableID}
	player.Rotation = &proto.APLRotation{}

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 600},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*testCaster)
}

// A landed hit, and the spell batch window the proc it set off resolves after.
func dealHit(t *testing.T, sim *core.Simulation, caster *testCaster, mask core.ProcMask) {
	caster.OnSpellHitDealt(sim, &core.Spell{ProcMask: mask, SpellSchool: core.SpellSchoolPhysical, Unit: &caster.Unit},
		&core.SpellResult{Target: sim.Encounter.ActiveTargetUnits[0], Outcome: core.OutcomeHit, Damage: 100})
	stepPast(t, sim, sim.CurrentTime+core.SpellBatchWindow)
}

func takeHit(t *testing.T, sim *core.Simulation, caster *testCaster) {
	caster.OnSpellHitTaken(sim, &core.Spell{ProcMask: core.ProcMaskMeleeMHAuto, SpellSchool: core.SpellSchoolPhysical, Unit: sim.Encounter.ActiveTargetUnits[0]},
		&core.SpellResult{Target: &caster.Unit, Outcome: core.OutcomeHit, Damage: 100})
	stepPast(t, sim, sim.CurrentTime+core.SpellBatchWindow)
}

// A proc row stating one A_PROC_TRIGGER_SPELL is built where the spell it fires restores mana to the
// wearer or changes one of the wearer's category cooldowns, and refused where that spell does neither.
// Three store rows, reshaped for the test, stand in for the spells fired.
func TestFiredSpellUnsupported(t *testing.T) {
	const gainID, cooldownID, damageID int32 = 1306613, 1306610, 1293421
	onWearer := [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_CASTER}
	reshape := func(id int32, effect spelldata.Effect) {
		effect.SpellID = id
		editRow(t, id, func(s *spelldata.Spell) { s.Effects = []spelldata.Effect{effect} })
	}
	reshape(gainID, spelldata.Effect{Type: dbcenums.E_ENERGIZE, Misc: int32(dbcenums.POWER_MANA), BasePoints: 40, Target: onWearer})
	reshape(cooldownID, spelldata.Effect{Type: dbcenums.E_MODIFY_COOLDOWNS_BY_CATEGORY, Misc: 19, BasePoints: -100000, Target: onWearer})
	reshape(damageID, spelldata.Effect{Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 100,
		Target: [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_TARGET_ENEMY}})
	firing := func(fired int32) *spelldata.Spell {
		const procID int32 = 991520
		return &spelldata.Spell{ID: procID, Name: "Test Proc", Effects: []spelldata.Effect{{SpellID: procID,
			Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PROC_TRIGGER_SPELL, TriggerID: fired, Target: onWearer}}}
	}

	for _, fired := range []int32{gainID, cooldownID} {
		if got := FiredSpellUnsupported(firing(fired)); len(got) > 0 {
			t.Errorf("a proc firing %d is refused: %v", fired, got)
		}
	}
	got := FiredSpellUnsupported(firing(damageID))
	if !slices.ContainsFunc(got, func(reason string) bool { return strings.Contains(reason, "neither restores") }) {
		t.Errorf("a proc firing damage is refused for %v, want the spell it fires named", got)
	}
}

// Each main-hand hit outside the trigger's lockout rolls its chance for the mana the fired spell
// rolls, a hit inside it or off the other hand restores nothing, and a wearer with no mana bar gets
// no listener at all. A copy of the trigger fixes the lockout at 3 s.
func TestManabladeRestoresManaOnTheImbuedHandsHits(t *testing.T) {
	const imbueID int32 = 991503
	editRow(t, manabladeProc, func(s *spelldata.Spell) { s.ICDMs = 3000 })
	trigger := spelldata.MustFind(manabladeProc)
	lockout, gain := trigger.ICD(), firedSpell(manabladeProc).EnergizeEffect()
	needAmount(t, "1306613's mana", gain.Max())
	sim, mage := newImbuedSim(t, testMage(), imbueID, 8715, manabladeProc)
	spent := mage.NewManaMetrics(core.ActionID{SpellID: 1})

	var procs int32
	for range 50 {
		stepPast(t, sim, sim.CurrentTime+lockout)
		mage.SpendMana(sim, mage.CurrentMana(), spent)
		before, _ := gainedFrom(mage, imbueID)
		dealHit(t, sim, mage, core.ProcMaskMeleeOHAuto)
		if got, _ := gainedFrom(mage, imbueID); got != before {
			t.Fatalf("an off-hand hit restored %v mana, want none", got-before)
		}
		dealHit(t, sim, mage, core.ProcMaskMeleeMHAuto)
		gained, events := gainedFrom(mage, imbueID)
		if events == procs {
			continue
		}
		procs = events
		if got := gained - before; got < gain.Min() || got > gain.Max() {
			t.Errorf("a proc restored %v mana, want %v to %v", got, gain.Min(), gain.Max())
		}
		mage.SpendMana(sim, mage.CurrentMana(), spent)
		for range 20 {
			dealHit(t, sim, mage, core.ProcMaskMeleeMHSpecial)
		}
		if _, got := gainedFrom(mage, imbueID); got != procs {
			t.Errorf("hits inside the %v lockout restored mana %d more times, want none", lockout, got-procs)
		}
	}
	if chance := trigger.StatedChance(); !plausibleProcs(procs, 50, chance) {
		t.Errorf("%d procs over 50 main-hand hits, want about %v at %v", procs, 50*chance, chance)
	}

	_, warrior := newImbuedSim(t, testWarrior(), 991504, 8715, manabladeProc)
	if auraByLabel(&warrior.Character, "Test Imbue") != nil {
		t.Error("a wearer with no mana bar gets a Manablade listener")
	}
}

// A melee hit the wearer takes rolls the chance that takes the fired spell's reduction off its
// category's cooldown, the one Fire Blast runs on; the wearer's own hits do not, and neither does a
// hit inside the trigger's lockout. The category is held for the whole reduction, which a proc makes
// ready at once. Copies of the rows fix the lockout at 4 s and the reduction at 100 s.
func TestSparkResetsItsCategoryOnHitsTaken(t *testing.T) {
	editRow(t, sparkProc, func(s *spelldata.Spell) { s.ICDMs = 4000 })
	editRow(t, firedSpell(sparkProc).ID, func(s *spelldata.Spell) {
		for i := range s.Effects {
			if s.Effects[i].Type == dbcenums.E_MODIFY_COOLDOWNS_BY_CATEGORY {
				s.Effects[i].BasePoints = -100000
			}
		}
	})
	trigger := spelldata.MustFind(sparkProc)
	cooldown := firedSpell(sparkProc).CategoryCooldownEffect()
	if cooldown == spelldata.NilEffect {
		t.Fatalf("%d fires no category cooldown change", sparkProc)
	}
	lockout, held := trigger.ICD(), -cooldown.TimeValue()
	sim, mage := newImbuedSim(t, testMage(), 991505, 8707, sparkProc)
	timer := mage.CategoryTimer(cooldown.Misc)

	timer.Set(sim.CurrentTime + held)
	for range 50 {
		dealHit(t, sim, mage, core.ProcMaskMeleeMHAuto)
	}
	if timer.IsReady(sim) {
		t.Fatalf("the wearer's own hits reset category %d", cooldown.Misc)
	}

	var procs int32
	for range 50 {
		stepPast(t, sim, sim.CurrentTime+lockout)
		timer.Set(sim.CurrentTime + held)
		takeHit(t, sim, mage)
		if !timer.IsReady(sim) {
			continue
		}
		procs++
		timer.Set(sim.CurrentTime + held)
		for range 20 {
			takeHit(t, sim, mage)
		}
		if timer.IsReady(sim) {
			t.Errorf("a hit taken inside the %v lockout reset category %d", lockout, cooldown.Misc)
		}
	}
	if chance := trigger.StatedChance(); !plausibleProcs(procs, 50, chance) {
		t.Errorf("%d procs over 50 hits taken, want about %v at %v", procs, 50*chance, chance)
	}
}

// A stone shaped like Dense Sharpening Stone 12404 (Effect 54, enchant 1643's 8 weapon damage) and a
// poison shaped like Instant Poison VI 8928 (Effect 360, enchant 625's combat spell 11337): one hand
// carries both.
func TestStoneAndPoisonShareAHand(t *testing.T) {
	const stone, poison int32 = 991513, 991514
	const mainHandID, offHandID int32 = 991511, 991512
	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{testOneHander(mainHandID), testOneHander(offHandID)},
		Consumables: []*proto.Consumable{
			{Id: stone, Name: "Test Stone", Type: proto.ConsumableType_ConsumableTypeImbue,
				TemporaryEnchant: &proto.SimEnchant{EffectId: 1643, WeaponDamage: 8}},
			{Id: poison, Name: "Test Poison", Type: proto.ConsumableType_ConsumableTypeWeaponImbue,
				TemporaryEnchant: &proto.SimEnchant{EffectId: 625, EnchantEffects: []*proto.ItemEffect{{BuffId: 11337}}}},
		},
	})
	NewSpellDataTemporaryEnchant(stone, SchoolPiercing{})
	NewSpellDataTemporaryEnchant(poison, SchoolPiercing{})

	player := testMage()
	player.Name = "Wearer"
	player.Buffs = &proto.IndividualBuffs{}
	player.Equipment = &proto.EquipmentSpec{Items: testHands(&proto.ItemSpec{Id: mainHandID}, &proto.ItemSpec{Id: offHandID})}
	player.Consumables = &proto.ConsumesSpec{MhImbueId: stone, MhWeaponImbueId: poison}
	player.Rotation = &proto.APLRotation{}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 600},
	}, simsignals.CreateSignals())
	character := sim.Raid.Parties[0].Players[0].GetCharacter()

	mainHand := character.Equipment[proto.ItemSlot_ItemSlotMainHand]
	if mainHand.TempEnchant != 1643 || mainHand.WeaponImbue != 625 {
		t.Fatalf("main hand carries enchant %d and imbue %d, want the stone's 1643 and the poison's 625", mainHand.TempEnchant, mainHand.WeaponImbue)
	}
	if got := character.WeaponFromMainHand().BaseDamageMin; got != 58 {
		t.Errorf("main hand deals %v minimum damage, want the sword's 50 and the stone's 8", got)
	}
	if auraByLabel(character, "Test Poison") == nil {
		t.Error("no listener casts the poison's combat spell beside the stone")
	}
}

func pseudoStats(stat proto.PseudoStat, value float64) []float64 {
	pseudo := make([]float64, stats.PseudoStatsLen)
	pseudo[stat] = value
	return pseudo
}

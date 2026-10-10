package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Vengeance 20049 states Can Proc From Procs, and its harmful spell bit hears a crit of any school.
const vengeanceTalent int32 = 20049

// Storm Gauntlets 12632's 16615 casts Lightning Strike 16614 off each melee hit, and Blazefury
// Medallion 17111's 7711 casts Fire Strike 7712 the same way. Their hits are procs, so an aura hears
// them only where it carries Can Proc From Procs: Vengeance's listener, built from its row as the
// paladin builds it, procs off their crits, and the same listener without the attribute does not.
func TestItemDamageProcCritsReachListenersThatHearProcs(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		itemID, trigger, damage int32
		listenerItemID          int32
		itemType                proto.ItemType
		itemSlot                proto.ItemSlot
	}{
		{"Storm Gauntlets", 991601, 16615, 16614, 991602, proto.ItemType_ItemTypeHands, proto.ItemSlot_ItemSlotHands},
		{"Blazefury Medallion", 991603, 7711, 7712, 991604, proto.ItemType_ItemTypeTrinket, proto.ItemSlot_ItemSlotTrinket1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{
				{Id: tc.itemID, Name: "Test " + tc.name, Type: tc.itemType,
					ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}},
				{Id: tc.listenerItemID, Name: "Test Vengeance", Type: proto.ItemType_ItemTypeNeck,
					ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}},
			}})
			registerSpellDataDamageProc(SpellDataProc{Name: "Test " + tc.name, ItemID: tc.itemID,
				TriggerSpellID: tc.trigger, BuffSpellID: tc.damage})

			heard := map[bool]int{}
			core.NewItemEffect(tc.listenerItemID, func(agent core.Agent) {
				character := agent.GetCharacter()
				for _, canProcFromProcs := range []bool{true, false} {
					trigger := spelldata.ProcTrigger(character, spelldata.MustFind(vengeanceTalent),
						func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
							if spell.ActionID.SpellID == tc.damage {
								heard[canProcFromProcs]++
							}
						})
					trigger.Name = "Test Vengeance"
					if !canProcFromProcs {
						trigger.Name += " Without Procs"
					}
					trigger.ProcMask = core.ProcMaskUnknown
					trigger.CanProcFromProcs = canProcFromProcs
					trigger.TriggerImmediately = true
					character.MakeProcTriggerAura(trigger)
				}
			})

			items := testHands(&proto.ItemSpec{}, &proto.ItemSpec{})
			items[tc.itemSlot] = &proto.ItemSpec{Id: tc.itemID}
			items[proto.ItemSlot_ItemSlotNeck] = &proto.ItemSpec{Id: tc.listenerItemID}
			sim := newTestCasterSim(items, nil)
			caster := sim.Raid.Parties[0].Players[0].(*testCaster)
			target := sim.Encounter.ActiveTargetUnits[0]
			caster.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100, stats.SpellCritPercent: 100})

			landMainHandSwing(sim, caster, target)

			damageSpell := caster.GetSpell(core.ActionID{SpellID: tc.damage})
			if damageSpell == nil || damageSpell.SpellMetrics[target.UnitIndex].Crits != 1 {
				t.Fatalf("the swing did not land one %d crit", tc.damage)
			}
			if heard[true] != 1 {
				t.Errorf("the listener with Can Proc From Procs heard %d crits, want 1", heard[true])
			}
			if heard[false] != 0 {
				t.Errorf("the listener without it heard %d crits, want none", heard[false])
			}
		})
	}
}

// Windfury Weapon's passive 439431 hears melee auto attacks and abilities, without Can Proc From Procs.
const windfuryWeaponPassive int32 = 439431

func landSwing(sim *core.Simulation, caster *testCaster, target *core.Unit, hand core.ProcMask) {
	caster.OnSpellHitDealt(sim, &core.Spell{ProcMask: hand},
		&core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
}

// Bloodpike 13057's chance-on-hit Rend 18202 applies a melee bleed and deals nothing up front, and
// Hanzo Sword 8190's Wound 16405 deals melee damage. A test one-hander carries the row in both hands,
// at a chance below 1 so that a strike rolling its own weapon again ends. Each strike is a special
// hit of the hand whose hit cast it: the weapon procs of that hand hear it, its own included, and
// the other hand's do not. A weapon proc on auto attacks alone does not hear it, and neither does
// Windfury Weapon's passive, an aura without Can Proc From Procs.
func TestWeaponProcStrikeIsASpecialHitOfItsHand(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		weaponID, neckID, strike int32
	}{
		{"Rend", 991606, 991607, 18202},
		{"Wound", 991608, 991609, 16405},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const swings = 20
			editRow(t, tc.strike, func(s *spelldata.Spell) {
				s.ProcChance, s.ProcChanceSource = 90, spelldata.ProcChanceColumn
			})
			core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{testOneHander(tc.weaponID),
				{Id: tc.neckID, Name: "Test Listeners", Type: proto.ItemType_ItemTypeNeck,
					ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})
			registerSpellDataDamageProc(SpellDataProc{Name: "Test " + tc.name + " Weapon", ItemID: tc.weaponID,
				TriggerSpellID: tc.strike, BuffSpellID: tc.strike, IsWeaponProc: true})

			heard := map[string]map[int32]int{}
			listen := func(name string) core.ProcHandler {
				heard[name] = map[int32]int{}
				return func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					heard[name][spell.ActionID.SpellID]++
				}
			}
			core.NewItemEffect(tc.neckID, func(agent core.Agent) {
				character := agent.GetCharacter()
				for name, mask := range map[string]core.ProcMask{
					"main-hand oil": core.ProcMaskMeleeMH,
					"off-hand oil":  core.ProcMaskMeleeOH,
					"auto attacks":  core.ProcMaskMeleeWhiteHit,
				} {
					character.MakeProcTriggerAura(core.ProcTrigger{
						Name:               "Test " + name,
						Callback:           core.CallbackOnSpellHitDealt,
						ProcMask:           mask,
						Outcome:            core.OutcomeLanded,
						IsWeaponProc:       true,
						ProcChance:         1,
						TriggerImmediately: true,
						Handler:            listen(name),
					})
				}
				windfury := spelldata.ProcTrigger(character, spelldata.MustFind(windfuryWeaponPassive),
					listen("Windfury Weapon"), spelldata.Chance(1))
				windfury.Name = "Test Windfury Weapon"
				windfury.ICD = 0
				windfury.TriggerImmediately = true
				character.MakeProcTriggerAura(windfury)
			})

			items := testHands(&proto.ItemSpec{Id: tc.weaponID}, &proto.ItemSpec{Id: tc.weaponID})
			items[proto.ItemSlot_ItemSlotNeck] = &proto.ItemSpec{Id: tc.neckID}
			sim := newTestCasterSim(items, nil)
			caster := sim.Raid.Parties[0].Players[0].(*testCaster)
			target := sim.Encounter.ActiveTargetUnits[0]

			for range swings {
				landSwing(sim, caster, target, core.ProcMaskMeleeMHAuto)
			}
			strike := caster.GetSpell(core.ActionID{SpellID: tc.strike})
			if casts := strike.SpellMetrics[target.UnitIndex].Casts; casts <= swings {
				t.Errorf("%d main-hand swings cast %d strikes, want more: a strike re-rolls its own weapon", swings, casts)
			}
			if heard["main-hand oil"][tc.strike] == 0 {
				t.Errorf("the main-hand weapon proc heard no main-hand strike")
			}
			if n := heard["off-hand oil"][tc.strike]; n != 0 {
				t.Errorf("the off-hand weapon proc heard %d main-hand strikes, want none", n)
			}

			mainHandHeard := heard["main-hand oil"][tc.strike]
			for range swings {
				landSwing(sim, caster, target, core.ProcMaskMeleeOHAuto)
			}
			if heard["off-hand oil"][tc.strike] == 0 {
				t.Errorf("the off-hand weapon proc heard no off-hand strike")
			}
			if n := heard["main-hand oil"][tc.strike]; n != mainHandHeard {
				t.Errorf("the main-hand weapon proc heard %d off-hand strikes, want none", n-mainHandHeard)
			}

			for _, name := range []string{"auto attacks", "Windfury Weapon"} {
				if n := heard[name][tc.strike]; n != 0 {
					t.Errorf("the %s listener heard %d strikes, want none", name, n)
				}
				if n := heard[name][0]; n != 2*swings {
					t.Errorf("the %s listener heard %d swings, want %d", name, n, 2*swings)
				}
			}
		})
	}
}

// Every landed Rend 18202 puts its bleed on the target, a Rend that its own strike chained included.
// A test one-hander carries the row at a chance below 1, and each swing lands once the last bleed has
// run out.
func TestWeaponProcStrikeChainAppliesEachLandedBleed(t *testing.T) {
	const weaponID, neckID, rend int32 = 991612, 991613, 18202
	const swings = 10
	editRow(t, rend, func(s *spelldata.Spell) {
		s.ProcChance, s.ProcChanceSource = 90, spelldata.ProcChanceColumn
	})
	duration := spelldata.MustFind(rend).Duration()
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: neckID, Name: "Test Oil",
		Type: proto.ItemType_ItemTypeNeck, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})
	landed := 0
	core.NewItemEffect(neckID, func(agent core.Agent) {
		agent.GetCharacter().MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Test Oil",
			Callback:           core.CallbackOnSpellHitDealt,
			ProcMask:           core.ProcMaskMeleeMH,
			Outcome:            core.OutcomeLanded,
			IsWeaponProc:       true,
			ProcChance:         1,
			TriggerImmediately: true,
			Handler: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				if spell.ActionID.SpellID == rend {
					landed++
				}
			},
		})
	})
	sim, caster, target := newSlowProcSim(t, weaponID, registerSpellDataDamageProc, rend, &proto.ItemSpec{Id: neckID})
	strike := caster.GetSpell(core.ActionID{SpellID: rend})

	for range swings {
		stepPast(t, sim, sim.CurrentTime+duration+time.Millisecond)
		before := landed
		landMainHandSwing(sim, caster, target)
		if landed > before && !strike.Dot(target).IsActive() {
			t.Fatalf("%d Rends landed off one swing and none left its bleed on the target", landed-before)
		}
	}
	if landed == 0 {
		t.Fatalf("no Rend landed in %d swings", swings)
	}
}

// A weapon proc of magic defense is no weapon's strike: Blaze 16898's hit is heard by no weapon proc.
func TestMagicWeaponProcIsHeardByNoWeaponProc(t *testing.T) {
	const weaponID, neckID, blaze int32 = 991610, 991611, 16898
	editRow(t, blaze, alwaysProcs)
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: neckID, Name: "Test Oil",
		Type: proto.ItemType_ItemTypeNeck, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})
	heard := map[int32]int{}
	core.NewItemEffect(neckID, func(agent core.Agent) {
		agent.GetCharacter().MakeProcTriggerAura(core.ProcTrigger{
			Name:               "Test Oil",
			Callback:           core.CallbackOnSpellHitDealt,
			Outcome:            core.OutcomeLanded,
			IsWeaponProc:       true,
			ProcChance:         1,
			TriggerImmediately: true,
			Handler: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				heard[spell.ActionID.SpellID]++
			},
		})
	})
	sim, caster, target := newSlowProcSim(t, weaponID, registerSpellDataDamageProc, blaze, &proto.ItemSpec{Id: neckID})

	landMainHandSwing(sim, caster, target)
	damage := caster.GetSpell(core.ActionID{SpellID: blaze})
	if damage == nil || damage.SpellMetrics[target.UnitIndex].Casts != 1 {
		t.Fatalf("the swing did not cast %d once", blaze)
	}
	if heard[0] != 1 {
		t.Errorf("the weapon proc heard %d swings, want 1", heard[0])
	}
	if heard[blaze] != 0 {
		t.Errorf("the weapon proc heard %d of %d's hits, want none", heard[blaze], blaze)
	}
}

// Blazing Rapier 12777's chance-on-hit Blaze 16898 deals Fire every period for its duration and
// nothing up front, and states no rate. Rolled at 100% for the test, the damage proc puts the damage
// over time on the target the swing hit, and it ticks the row's duration over its period.
func TestWeaponProcDealingDamageOverTimeAloneAppliesIt(t *testing.T) {
	const weaponID, blaze int32 = 991605, 16898
	editRow(t, blaze, alwaysProcs)
	row := spelldata.MustFind(blaze)
	duration := row.Duration()
	wantTicks := int32(duration / row.PeriodicDamageEffect().Period())
	sim, caster, target := newSlowProcSim(t, weaponID, registerSpellDataDamageProc, blaze, nil)

	start := sim.CurrentTime
	landMainHandSwing(sim, caster, target)
	damage := caster.GetSpell(core.ActionID{SpellID: blaze})
	if damage == nil || !damage.Dot(target).IsActive() {
		t.Fatalf("the swing put no %d damage over time on the target", blaze)
	}

	stepPast(t, sim, start+duration+time.Millisecond)
	if damage.Dot(target).IsActive() {
		t.Errorf("the damage over time is still up after the row's %v", duration)
	}
	metrics := damage.SpellMetrics[target.UnitIndex]
	if ticks := metrics.Ticks + metrics.ResistedTicks + metrics.CritTicks + metrics.ResistedCritTicks; wantTicks < 2 || ticks != wantTicks {
		t.Errorf("the damage over time ticked %d times, want the row's %d", ticks, wantTicks)
	}
}

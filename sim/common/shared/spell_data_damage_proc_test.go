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

// The deferred re-rolls of the strikes the swings so far cast, which a chance below 1 ends.
func settleChains(t *testing.T, sim *core.Simulation) {
	t.Helper()
	stepPast(t, sim, sim.CurrentTime+time.Second)
}

func mustGetSpell(t *testing.T, caster *testCaster, id int32) *core.Spell {
	t.Helper()
	spell := caster.GetSpell(core.ActionID{SpellID: id})
	if spell == nil {
		t.Fatalf("no spell %d is registered", id)
	}
	return spell
}

func weaponProcListener(name string, mask core.ProcMask, handler core.ProcHandler) core.ProcTrigger {
	return core.ProcTrigger{
		Name:               name,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           mask,
		Outcome:            core.OutcomeLanded,
		IsWeaponProc:       true,
		ProcChance:         1,
		TriggerImmediately: true,
		Handler:            handler,
	}
}

// Bloodpike 13057's Rend 18202 (a bleed, no hit) and Hanzo Sword 8190's Wound 16405 (a hit) at a chance
// below 1. Each strike is a main-hand special whichever hand cast it: the main hand's weapon procs hear
// it, the off hand's, auto attacks' and Windfury Weapon's (Wound only, it needs damage) do not.
func TestWeaponProcStrikeIsAMainHandSpecial(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		weaponID, neckID, strike int32
		windfury                 bool
	}{
		{"Rend", 991606, 991607, 18202, false},
		{"Wound", 991608, 991609, 16405, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const swings = 20
			editRow(t, tc.strike, func(s *spelldata.Spell) {
				s.ProcChance, s.ProcChanceSource = 90, spelldata.ProcChanceColumn
			})
			core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{testOneHander(tc.weaponID)}})
			withAuraItem(tc.neckID, proto.ItemType_ItemTypeNeck)
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
					character.MakeProcTriggerAura(weaponProcListener("Test "+name, mask, listen(name)))
				}
				if tc.windfury {
					windfury := spelldata.ProcTrigger(character, spelldata.MustFind(windfuryWeaponPassive),
						listen("Windfury Weapon"), spelldata.Chance(1))
					windfury.Name = "Test Windfury Weapon"
					windfury.ICD = 0
					windfury.TriggerImmediately = true
					character.MakeProcTriggerAura(windfury)
				}
			})

			items := testHands(&proto.ItemSpec{Id: tc.weaponID}, &proto.ItemSpec{Id: tc.weaponID})
			items[proto.ItemSlot_ItemSlotNeck] = &proto.ItemSpec{Id: tc.neckID}
			sim := newTestCasterSim(items, nil)
			caster := sim.Raid.Parties[0].Players[0].(*testCaster)
			target := sim.Encounter.ActiveTargetUnits[0]

			for range swings {
				landSwing(sim, caster, target, core.ProcMaskMeleeMHAuto)
			}
			settleChains(t, sim)
			strike := mustGetSpell(t, caster, tc.strike)
			if casts := strike.SpellMetrics[target.UnitIndex].Casts; casts <= swings {
				t.Errorf("%d main-hand swings cast %d strikes, want more: a strike re-rolls the main-hand weapon", swings, casts)
			}
			if heard["main-hand oil"][tc.strike] == 0 {
				t.Errorf("the main-hand weapon proc heard no strike a main-hand swing cast")
			}

			mainHandHeard := heard["main-hand oil"][tc.strike]
			for range swings {
				landSwing(sim, caster, target, core.ProcMaskMeleeOHAuto)
			}
			settleChains(t, sim)
			if heard["main-hand oil"][tc.strike] == mainHandHeard {
				t.Errorf("the main-hand weapon proc heard no strike an off-hand swing cast")
			}
			if n := heard["off-hand oil"][tc.strike]; n != 0 {
				t.Errorf("the off-hand weapon proc heard %d strikes, want none", n)
			}

			names := []string{"auto attacks"}
			if tc.windfury {
				names = append(names, "Windfury Weapon")
			}
			for _, name := range names {
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

// Two test one-handers, each with its own row and rate: Wound 16405 in the main hand and Rend 18202
// in the off hand, each measured on the weapon carrying it, so a wielder that dual wields carries
// them. A strike is a main-hand special whichever hand's swing cast it, so it rolls the main-hand
// weapon's proc and never the off-hand one's: off-hand swings cast Rends whose hits roll Wound, and
// main-hand swings cast Wounds that roll Wound again and never Rend. The swings are the test's own,
// so the wielder's auto attacks are off.
func TestWeaponProcStrikeRollsTheMainHandWeapon(t *testing.T) {
	const mainHandID, offHandID, wound, rend int32 = 991614, 991615, 16405, 18202
	const swings = 20
	editRow(t, wound, func(s *spelldata.Spell) { s.RPPM = 18 })
	editRow(t, rend, func(s *spelldata.Spell) { s.RPPM = 12 })
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{testOneHander(mainHandID), testOneHander(offHandID)}})
	registerSpellDataDamageProc(SpellDataProc{Name: "Test Wound Weapon", ItemID: mainHandID,
		TriggerSpellID: wound, BuffSpellID: wound, IsWeaponProc: true})
	registerSpellDataDamageProc(SpellDataProc{Name: "Test Rend Weapon", ItemID: offHandID,
		TriggerSpellID: rend, BuffSpellID: rend, IsWeaponProc: true})

	sim := newTestCasterSimAgainst(testHands(&proto.ItemSpec{Id: mainHandID}, &proto.ItemSpec{Id: offHandID}), nil, 1,
		func(player *proto.Player) { player.Spec = &proto.Player_EnhancementShaman{} })
	caster := sim.Raid.Parties[0].Players[0].(*testCaster)
	caster.AutoAttacks.CancelAutoSwing(sim)
	target := sim.Encounter.ActiveTargetUnits[0]
	woundSpell, rendSpell := mustGetSpell(t, caster, wound), mustGetSpell(t, caster, rend)
	casts := func(spell *core.Spell) int32 { return spell.SpellMetrics[target.UnitIndex].Casts }

	for range swings {
		landSwing(sim, caster, target, core.ProcMaskMeleeOHAuto)
	}
	settleChains(t, sim)
	if n := casts(rendSpell); n == 0 || n > swings {
		t.Errorf("%d off-hand swings cast %d Rends, want some and at most one each: a Rend never rolls the off-hand weapon", swings, n)
	}
	if casts(woundSpell) == 0 {
		t.Errorf("no Rend an off-hand swing cast rolled the main-hand weapon's Wound")
	}

	rends, wounds := casts(rendSpell), casts(woundSpell)
	for range swings {
		landSwing(sim, caster, target, core.ProcMaskMeleeMHAuto)
	}
	settleChains(t, sim)
	if n := casts(rendSpell) - rends; n != 0 {
		t.Errorf("%d main-hand swings cast %d Rends, want none", swings, n)
	}
	if n := casts(woundSpell) - wounds; n <= swings {
		t.Errorf("%d main-hand swings cast %d Wounds, want more: a Wound rolls the main-hand weapon again", swings, n)
	}
}

// A strike whose own weapon procs on every hit and that always lands chains for as long as it is
// rolled, each re-roll one spell batch after the strike that asked for it rather than inside it.
func TestWeaponProcStrikeThatAlwaysProcsChainsOneBatchApart(t *testing.T) {
	const weaponID, wound int32 = 991616, 16405
	const batches = 5
	editRow(t, wound, alwaysProcs)
	sim, caster, target := newSlowProcSim(t, weaponID, registerSpellDataDamageProc, wound, nil)
	caster.AddStatsDynamic(sim, stats.Stats{stats.PhysicalHitPercent: 100, stats.ExpertisePercent: 100})
	strike := mustGetSpell(t, caster, wound)
	metrics := &strike.SpellMetrics[target.UnitIndex]

	landMainHandSwing(sim, caster, target)
	if metrics.Casts != 1 {
		t.Fatalf("the swing cast %d strikes at once, want 1: a re-roll waits a batch", metrics.Casts)
	}
	stepPast(t, sim, sim.CurrentTime+batches*core.SpellBatchWindow+time.Millisecond)
	if metrics.Casts != 1+batches {
		t.Errorf("%d batches after the swing %d strikes were cast, want %d", batches, metrics.Casts, 1+batches)
	}
	if avoided := metrics.Misses + metrics.Dodges + metrics.Parries; avoided != 0 {
		t.Errorf("%d strikes were avoided, want every one to land", avoided)
	}
}

// Every landed Rend 18202 puts its bleed on the target, a Rend that its own strike chained included,
// and no weapon proc hears the bleed's ticks. A test one-hander carries the row at a chance below 1,
// and each swing lands once the last bleed has run out.
func TestWeaponProcStrikeChainAppliesEachLandedBleed(t *testing.T) {
	const weaponID, neckID, rend int32 = 991612, 991613, 18202
	const swings = 10
	editRow(t, rend, func(s *spelldata.Spell) {
		s.ProcChance, s.ProcChanceSource = 90, spelldata.ProcChanceColumn
	})
	duration := spelldata.MustFind(rend).Duration()
	withAuraItem(neckID, proto.ItemType_ItemTypeNeck)
	landed := 0
	core.NewItemEffect(neckID, func(agent core.Agent) {
		character := agent.GetCharacter()
		oil := spelldata.ProcTrigger(character, spelldata.MustFind(rend),
			func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				if spell.ActionID.SpellID == rend {
					landed++
				}
			}, spelldata.WeaponProc(), spelldata.Chance(1))
		oil.Name = "Test Oil"
		oil.TriggerImmediately = true
		character.MakeProcTriggerAura(oil)
	})
	sim, caster, target := newSlowProcSim(t, weaponID, registerSpellDataDamageProc, rend, &proto.ItemSpec{Id: neckID})
	strike := mustGetSpell(t, caster, rend)

	for range swings {
		stepPast(t, sim, sim.CurrentTime+duration+time.Millisecond)
		before := landed
		landMainHandSwing(sim, caster, target)
		settleChains(t, sim)
		if landed > before && !strike.Dot(target).IsActive() {
			t.Fatalf("%d Rends landed off one swing and none left its bleed on the target", landed-before)
		}
	}
	stepPast(t, sim, sim.CurrentTime+duration+time.Millisecond)
	if landed == 0 {
		t.Fatalf("no Rend landed in %d swings", swings)
	}
	metrics := strike.SpellMetrics[target.UnitIndex]
	if metrics.Ticks == 0 {
		t.Fatalf("no bleed ticked")
	}
	if landed > int(metrics.Casts) {
		t.Errorf("the weapon proc heard %d Rends of %d cast: it heard the bleed's %d ticks", landed, metrics.Casts, metrics.Ticks)
	}
}

// A weapon proc of magic defense is no weapon's strike: Blaze 16898's hit is heard by no weapon proc.
func TestMagicWeaponProcIsHeardByNoWeaponProc(t *testing.T) {
	const weaponID, neckID, blaze int32 = 991610, 991611, 16898
	editRow(t, blaze, alwaysProcs)
	withAuraItem(neckID, proto.ItemType_ItemTypeNeck)
	heard := map[int32]int{}
	core.NewItemEffect(neckID, func(agent core.Agent) {
		agent.GetCharacter().MakeProcTriggerAura(weaponProcListener("Test Oil", core.ProcMaskUnknown,
			func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				heard[spell.ActionID.SpellID]++
			}))
	})
	sim, caster, target := newSlowProcSim(t, weaponID, registerSpellDataDamageProc, blaze, &proto.ItemSpec{Id: neckID})

	landMainHandSwing(sim, caster, target)
	damage := mustGetSpell(t, caster, blaze)
	if damage.SpellMetrics[target.UnitIndex].Casts != 1 {
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

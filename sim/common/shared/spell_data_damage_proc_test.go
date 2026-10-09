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

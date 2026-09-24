package warrior

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

const stanceEffectCategory = "Stance"

var battleStanceRank = spellData.BattleStance.Highest()
var defensiveStanceRank = spellData.DefensiveStance.Highest()
var berserkerStanceRank = spellData.BerserkerStance.Highest()

// The passive each stance carries, which is where the client states what standing in it is worth.
var battleStancePassive = spellData.BattleStancePassive.Highest()
var defensiveStancePassive = spellData.DefensiveStancePassive.Highest()
var berserkerStancePassive = spellData.BerserkerStancePassive.Highest()

func (warrior *Warrior) makeStanceSpell(flags core.ClassFlags, rank *spelldata.Spell, aura *core.Aura, stanceCD *core.Timer) *core.Spell {
	form := rank.ShapeshiftForm()
	actionID := aura.ActionID
	rageMetrics := warrior.NewRageMetrics(actionID)
	maxRetainedRage := spellData.TacticalMastery.ValueAt(1) + spellData.ImprovedTacticalMastery.ValueAt(warrior.Talents.ImprovedTacticalMastery)

	return warrior.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		DefenseType: rank.DefenseTypeCore(),
		ClassFlags:  flags,
		Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    stanceCD,
				Duration: cooldownOf(rank),
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warrior.ShapeshiftForm != form
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			if warrior.WarriorInputs.StanceSnapshot {
				// Delayed, so same-GCD casts are affected by the current aura.
				// Alternatively, those casts could just (artificially) happen before the stance change.
				pa := sim.GetConsumedPendingActionFromPool()
				pa.NextActionAt = sim.CurrentTime + 10*time.Millisecond
				pa.OnAction = aura.Activate
				sim.AddPendingAction(pa)
			} else {
				aura.Activate(sim)
			}

			if warrior.CurrentRage() > maxRetainedRage {
				warrior.SpendRage(sim, warrior.CurrentRage()-maxRetainedRage, rageMetrics)
			}

			warrior.ShapeshiftForm = form
		},

		RelatedSelfBuff: aura,
	})
}

func (warrior *Warrior) registerBattleStanceAura() *core.Aura {
	actionID := core.ActionID{SpellID: battleStanceRank.ID}

	aura := warrior.RegisterAura(core.Aura{
		Label:      "Battle Stance",
		ActionID:   actionID,
		Duration:   core.NeverExpires,
		BuildPhase: core.Ternary(warrior.DefaultStance == proto.WarriorStance_WarriorStanceBattle, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone),
	})
	spelldata.ParseEffects(&warrior.Character, aura, battleStancePassive)

	aura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})

	return aura
}

func (warrior *Warrior) registerDefensiveStanceAura() *core.Aura {
	actionID := core.ActionID{SpellID: defensiveStanceRank.ID}

	aura := warrior.RegisterAura(core.Aura{
		Label:      "Defensive Stance",
		ActionID:   actionID,
		Duration:   core.NeverExpires,
		BuildPhase: core.Ternary(warrior.DefaultStance == proto.WarriorStance_WarriorStanceDefensive, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone),
	})
	spelldata.ParseEffects(&warrior.Character, aura, defensiveStancePassive)

	if warrior.Talents.Defiance > 0 {
		// The stance the talent's threat applies in is the aura it hangs on; the shield the
		// tooltip asks for is stated nowhere in the row, so it is the caller's condition, re-read
		// on an off-hand swap.
		defiance := spelldata.ParseEffects(&warrior.Character, aura,
			spellData.Defiance.Rank(warrior.Talents.Defiance),
			spelldata.Conditional(func() bool { return warrior.PseudoStats.CanBlock }))

		warrior.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotOffHand}, func(sim *core.Simulation, _ proto.ItemSlot) {
			defiance.Refresh(sim)
		})
	}

	aura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})

	return aura
}

func (warrior *Warrior) registerBerserkerStanceAura() *core.Aura {
	actionId := core.ActionID{SpellID: berserkerStanceRank.ID}

	aura := warrior.RegisterAura(core.Aura{
		Label:      "Berserker Stance",
		ActionID:   actionId,
		Duration:   core.NeverExpires,
		BuildPhase: core.Ternary(warrior.DefaultStance == proto.WarriorStance_WarriorStanceBerserker, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone),
	})
	// The row's fourth effect is an attack power percentage of 0, which has no sim kind and nothing
	// to apply either way.
	spelldata.ParseEffects(&warrior.Character, aura, berserkerStancePassive)

	aura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})

	return aura
}

func (warrior *Warrior) registerStances() {
	stanceCD := warrior.NewTimer()
	battleStanceAura := warrior.registerBattleStanceAura()
	defensiveStanceAura := warrior.registerDefensiveStanceAura()
	berserkerStanceAura := warrior.registerBerserkerStanceAura()
	warrior.BattleStance = warrior.makeStanceSpell(SpellFlagsBattleStance, battleStanceRank, battleStanceAura, stanceCD)
	warrior.DefensiveStance = warrior.makeStanceSpell(SpellFlagsDefensiveStance, defensiveStanceRank, defensiveStanceAura, stanceCD)
	warrior.BerserkerStance = warrior.makeStanceSpell(SpellFlagsBerserkerStance, berserkerStanceRank, berserkerStanceAura, stanceCD)

	switch warrior.DefaultStance {
	case proto.WarriorStance_WarriorStanceBattle:
		core.MakePermanent(warrior.BattleStance.RelatedSelfBuff)
	case proto.WarriorStance_WarriorStanceDefensive:
		core.MakePermanent(warrior.DefensiveStance.RelatedSelfBuff)
	case proto.WarriorStance_WarriorStanceBerserker:
		core.MakePermanent(warrior.BerserkerStance.RelatedSelfBuff)
	}
}

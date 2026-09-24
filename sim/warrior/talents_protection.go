package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (warrior *Warrior) registerProtectionTalents() {
	// Tier 1
	warrior.registerShieldSpecialization()
	warrior.registerAnticipation()

	// Tier 2
	// Improved Bloodrage: bloodrage.go
	warrior.registerToughness()
	warrior.registerImprovedThunderClap()

	// Tier 3
	warrior.registerLastStand()
	warrior.registerMasterOfDefense()
	warrior.registerImprovedRevenge()
	// Defiance: stances.go

	// Tier 4
	warrior.registerImprovedSunderArmor()
	warrior.registerImprovedDisarm()
	// Vanguard: charge.go

	// Tier 5
	warrior.registerImprovedShieldWall()
	warrior.registerConcussionBlow()
	warrior.registerImprovedShieldBash()
	warrior.registerBastion()

	// Tier 6
	warrior.registerFocusedRage()

	// Tier 7
	warrior.registerShieldSlam()
}

func (warrior *Warrior) registerAnticipation() {
	if warrior.Talents.Anticipation == 0 {
		return
	}

	warrior.AddStat(stats.DefenseRating, spellData.Anticipation.ValueAt(warrior.Talents.Anticipation)*core.DefenseRatingPerDefenseLevel)
}

var shieldSpecializationEnergize = spellData.ShieldSpecializationTriggered.Highest()

func (warrior *Warrior) registerShieldSpecialization() {
	if warrior.Talents.ShieldSpecialization == 0 {
		return
	}

	// Effect 1 is the block bonus; effect 2 is the proc trigger, which the parse skips.
	spelldata.ParseStatic(&warrior.Character,
		spellData.ShieldSpecialization.Rank(warrior.Talents.ShieldSpecialization))

	warrior.registerRageOnAvoid(spellData.ShieldSpecialization.Rank(warrior.Talents.ShieldSpecialization),
		shieldSpecializationEnergize, core.OutcomeBlock, nil)
}

// A chance to gain rage when an incoming attack is blocked, dodged or parried. The rate is the
// effect ladder the tooltip's $m names, and the 100 in the proc chance column is noise; the
// energize the triggered spell states is on the client's 0-1000 rage bar.
//
// No proc mask states an outcome, so which of the three it is stays the caller's; that the hit
// carries no damage is the row's, off its outcome hint.
func (warrior *Warrior) registerRageOnAvoid(driver *spelldata.Spell, energize *spelldata.Spell, outcome core.HitOutcome, extra core.ProcExtraCondition) {
	rage := energize.EnergizeEffect().Tenths()
	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: energize.ID})

	trigger := spelldata.ProcTrigger(&warrior.Character, driver,
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			warrior.AddRage(sim, rage, rageMetrics)
		})
	trigger.Outcome = outcome
	trigger.TriggerImmediately = true
	trigger.ExtraCondition = extra

	warrior.MakeProcTriggerAura(trigger)
}

func (warrior *Warrior) registerToughness() {
	if warrior.Talents.Toughness == 0 {
		return
	}

	// The client states the ladder twice, once on base armor and once on bonus armor; the parse
	// takes the base armor effect, which the tooltip words as armor from items, onto the equipment
	// share of the sim's single Armor stat, and skips the bonus armor one.
	spelldata.ParseStatic(&warrior.Character, spellData.Toughness.Rank(warrior.Talents.Toughness))
}

var lastStandRank = spellData.LastStand.Highest()
var lastStandBuff = spellData.LastStandTriggered.Highest()

func (warrior *Warrior) registerLastStand() {
	if !warrior.Talents.LastStand {
		return
	}

	actionID := core.ActionID{SpellID: lastStandRank.ID}
	healthMetrics := warrior.NewHealthMetrics(actionID)

	var bonusHealth float64
	aura := warrior.RegisterAura(core.Aura{
		Label:    "Last Stand",
		ActionID: actionID,
		Duration: lastStandBuff.Duration(),
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			bonusHealth = warrior.MaxHealth() * lastStandBuff.Effect(dbcenums.A_MOD_MAX_HEALTH, 0).Percent()
			warrior.UpdateMaxHealth(sim, bonusHealth, healthMetrics)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			warrior.UpdateMaxHealth(sim, -bonusHealth, healthMetrics)
		},
	})

	config := spelldata.SpellConfig(&warrior.Unit, lastStandRank)

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		aura.Activate(sim)
	}

	config.RelatedSelfBuff = aura

	spell := warrior.RegisterSpell(config)

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeSurvival,
		BuffAura: &core.StatBuffAura{
			Aura:            aura,
			BuffedStatTypes: []stats.Stat{stats.Health},
		},
	})
}

func (warrior *Warrior) registerImprovedSunderArmor() {
	if warrior.Talents.ImprovedSunderArmor == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.ImprovedSunderArmor.Rank(warrior.Talents.ImprovedSunderArmor))
}

func (warrior *Warrior) registerImprovedShieldWall() {
	if warrior.Talents.ImprovedShieldWall == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.ImprovedShieldWall.Rank(warrior.Talents.ImprovedShieldWall))
}

var concussionBlowRank = spellData.ConcussionBlow.Highest()

// TODO: In-game testing if this generates threat
func (warrior *Warrior) registerConcussionBlow() {
	if !warrior.Talents.ConcussionBlow {
		return
	}

	config := spelldata.SpellConfig(&warrior.Unit, concussionBlowRank,
		spelldata.Flags(core.SpellFlagMeleeMetrics|core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskMeleeMHSpecial

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}

var shieldSlamRank = spellData.ShieldSlam.Highest()

func (warrior *Warrior) registerShieldSlam() {
	if !warrior.Talents.ShieldSlam {
		return
	}

	config := spelldata.SpellConfig(&warrior.Unit, shieldSlamRank, spelldata.Melee(core.ProcMaskMeleeOHSpecial))
	// TODO: In-game testing needed for threat multiplier / flat threat
	config.FlatThreatBonus = 0

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return warrior.PseudoStats.CanBlock
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := shieldSlamRank.DamageEffect().Roll(sim, core.CharacterLevel) + warrior.BlockDamageReduction()
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}

func (warrior *Warrior) registerFocusedRage() {
	if warrior.Talents.FocusedRage == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.FocusedRage.Rank(warrior.Talents.FocusedRage))
}

var masterOfDefenseEnergize = spellData.MasterOfDefenseTriggered.Highest()

func (warrior *Warrior) registerMasterOfDefense() {
	if warrior.Talents.MasterOfDefense == 0 {
		return
	}

	// A shield has to be equipped, which the row does not state.
	warrior.registerRageOnAvoid(spellData.MasterOfDefense.Rank(warrior.Talents.MasterOfDefense),
		masterOfDefenseEnergize, core.OutcomeDodge|core.OutcomeParry,
		func(_ *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
			return warrior.PseudoStats.CanBlock
		})
}

func (warrior *Warrior) registerImprovedRevenge() {
	if warrior.Talents.ImprovedRevenge == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.ImprovedRevenge.Rank(warrior.Talents.ImprovedRevenge))
}

func (warrior *Warrior) registerImprovedDisarm() {
	if warrior.Talents.ImprovedDisarm == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.ImprovedDisarm.Rank(warrior.Talents.ImprovedDisarm))
}

var improvedShieldBashSilence = spellData.ImprovedShieldBashTriggered.Highest()

func (warrior *Warrior) registerImprovedShieldBash() {
	if warrior.Talents.ImprovedShieldBash == 0 {
		return
	}

	// TODO: nothing in the sim reads a silence on an enemy, so the aura only shows up in metrics.
	silenceAuras := warrior.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			Label:    "Shield Bash - Silence",
			ActionID: core.ActionID{SpellID: improvedShieldBashSilence.ID},
			Duration: improvedShieldBashSilence.Duration(),
		})
	})

	// The rate is the effect ladder the tooltip's $m1 names, and the 100 in the proc chance column
	// is noise. The one ability it fires on is a shape no proc mask states, so the row's listener
	// is narrowed to Shield Bash by hand.
	trigger := spelldata.ProcTrigger(&warrior.Character,
		spellData.ImprovedShieldBash.Rank(warrior.Talents.ImprovedShieldBash),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			silenceAuras.Get(result.Target).Activate(sim)
		})
	trigger.ClassSpellMask = SpellMaskShieldBash
	trigger.TriggerImmediately = true

	warrior.MakeProcTriggerAura(trigger)
}

func (warrior *Warrior) registerBastion() {
	if warrior.Talents.Bastion == 0 {
		return
	}

	// The row raises the physical school and states no shield of its own: the tooltip's shield
	// requirement is the caller's condition, re-read on an off-hand swap.
	parsed := spelldata.ParseStatic(&warrior.Character, spellData.Bastion.Rank(warrior.Talents.Bastion),
		spelldata.Conditional(func() bool { return warrior.PseudoStats.CanBlock }))

	warrior.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotOffHand}, func(sim *core.Simulation, slot proto.ItemSlot) {
		parsed.Refresh(sim)
	})
}

func (warrior *Warrior) registerImprovedThunderClap() {
	if warrior.Talents.ImprovedThunderClap == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.ImprovedThunderClap.Rank(warrior.Talents.ImprovedThunderClap))
}

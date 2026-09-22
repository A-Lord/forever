package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (warrior *Warrior) registerFuryTalents() {
	// Tier 1
	// Booming Voice (12321) widens the shout radius only, which the sim does not model.
	warrior.registerCruelty()

	// Tier 2
	warrior.registerIronWill()
	warrior.registerUnbridledWrath()

	// Tier 3
	warrior.registerImprovedCleave()
	warrior.registerPiercingHowl()
	warrior.registerBloodCraze()
	// Boundless Rage: warrior.go, when it enables the rage bar

	// Tier 4
	warrior.registerDualWieldSpecialization()
	warrior.registerRagingBlows()
	warrior.registerEnrage()
	warrior.registerImprovedExecute()

	// Tier 5
	warrior.registerPrecision()
	warrior.registerDeathWish()
	warrior.registerImprovedIntercept()

	// Tier 6
	// Improved Berserker Rage: berserker_rage.go
	warrior.registerFlurry()

	// Tier 7
	warrior.registerBloodthirst()
}

func (warrior *Warrior) registerCruelty() {
	if warrior.Talents.Cruelty == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.Cruelty.Rank(warrior.Talents.Cruelty))
}

var unbridledWrathRank = spellData.UnbridledWrathTriggered.Highest()

var unbridledWrathRage = unbridledWrathRank.EnergizeEffect().Tenths()

func (warrior *Warrior) registerUnbridledWrath() {
	if warrior.Talents.UnbridledWrath == 0 {
		return
	}

	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: unbridledWrathRank.ID})

	// The rate is the effect ladder the tooltip's $m1 names - 12/24/36/48/60 by rank, where the
	// proc chance column reads a flat 60 - and the row's mask is the white hits the tooltip means.
	warrior.MakeProcTriggerAura(spelldata.ProcTrigger(&warrior.Character,
		spellData.UnbridledWrath.Rank(warrior.Talents.UnbridledWrath),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			// The tooltip of 12322 doubles the rage for a two-handed weapon.
			twoHanded := warrior.GetMainHandType() == proto.HandType_HandTypeTwoHand
			warrior.AddRage(sim, unbridledWrathRage*core.TernaryFloat64(twoHanded, 2, 1), rageMetrics)
		}))
}

func (warrior *Warrior) registerDualWieldSpecialization() {
	if warrior.Talents.DualWieldSpecialization == 0 {
		return
	}

	rank := spellData.DualWieldSpecialization.Rank(warrior.Talents.DualWieldSpecialization)

	// Effect 3 states A_MOD_HIT_CHANCE with nothing narrowing it, which the parse reads as hit on
	// every attack; the tooltip states the chance to hit with off-hand attacks, so it stays a mod
	// on the off-hand's own hits.
	spelldata.ParseStatic(&warrior.Character, rank, spelldata.SkipEffects(3))

	warrior.AddStaticMod(core.SpellModConfig{
		ProcMask:   core.ProcMaskMeleeOH,
		Kind:       core.SpellMod_BonusHit_Percent,
		FloatValue: rank.EffectN(3).BaseValue(),
	})

	// The off-hand rage the tooltip's $m2 states is the dummy at index 2, which the parse skips.
	warrior.SetOffHandRageMultiplier(1 + rank.EffectN(2).Percent())
}

func (warrior *Warrior) registerIronWill() {
	if warrior.Talents.IronWill == 0 {
		return
	}
	spelldata.ParseStatic(&warrior.Character, spellData.IronWill.Rank(warrior.Talents.IronWill))
}

func (warrior *Warrior) registerImprovedExecute() {
	if warrior.Talents.ImprovedExecute == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.ImprovedExecute.Rank(warrior.Talents.ImprovedExecute))
}

var enrageBuff = spellData.EnrageTriggered.Highest()

func (warrior *Warrior) registerEnrage() {
	if warrior.Talents.Enrage == 0 {
		return
	}

	warrior.EnrageAura = warrior.GetOrRegisterAura(spelldata.AuraConfig(enrageBuff))

	// 12880 states the bucket and the duration, and the sim's only Enrage buff row states a flat
	// 10; the percentage per rank is the talent's own ladder, which is what 12317's tooltip reads
	// as $m1%. So the value is the talent's and only the bucket - the physical school's damage
	// dealt multiplier, not a per-spell mod - comes off the buff row.
	warrior.EnrageAura.AttachMultiplicativePseudoStatBuff(
		&warrior.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical],
		spellData.Enrage.MultiplierAt(warrior.Talents.Enrage),
	)

	warrior.EnrageAura.NewExclusiveEffect("Enrage", true, core.ExclusiveEffect{Priority: spellData.Enrage.ValueAt(warrior.Talents.Enrage)})

	// The rate is the shape where the tooltip carries $h%: a real 30% roll on damage taken, which
	// is the proc chance column.
	trigger := spelldata.ProcTrigger(&warrior.Character, spellData.Enrage.Rank(warrior.Talents.Enrage),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			warrior.EnrageAura.Activate(sim)
		})
	trigger.Name = "Enrage - Trigger"

	warrior.MakeProcTriggerAura(trigger)
}

var flurryBuff = spellData.FlurryTriggered.Highest()

func (warrior *Warrior) registerFlurry() {
	if warrior.Talents.Flurry == 0 {
		return
	}

	// 12966 supplies the duration and the three charges. The haste is the talent's own ladder:
	// 12319 reads "Increases your melee attack speed by $m1% for your next $12966n swings", so the
	// buff row is named for the swing count alone.
	//
	// TODO: Ingame test needed: the talent ladder gives 5% per point (25% at rank 5) while the
	// applied buff 12966 carries a flat 30%.
	flurryAura := warrior.RegisterAura(spelldata.AuraConfig(flurryBuff)).
		AttachMultiplyMeleeSpeed(spellData.Flurry.MultiplierAt(warrior.Talents.Flurry))

	// The proc shape with no roll: 12319 states its rate as "always" and the crit is the
	// condition, which this handler reads off the result rather than the trigger because the same
	// listener has to see the white hits that spend a charge. The row's mask reaches ranged and
	// spell hits too; the tooltip says a melee critical strike, so the mask stays the caller's.
	trigger := spelldata.ProcTrigger(&warrior.Character, spellData.Flurry.Rank(warrior.Talents.Flurry),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.Matches(SpellMaskWhirlwindOh) {
				return
			}

			if result.Outcome.Matches(core.OutcomeCrit) {
				flurryAura.Activate(sim)
				flurryAura.SetStacks(sim, int32(flurryBuff.ProcCharges))
				return
			}

			if flurryAura.IsActive() && spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
				flurryAura.RemoveStack(sim)
			}
		})
	trigger.Name = "Flurry - Trigger"
	trigger.ProcMask = core.ProcMaskMelee
	trigger.TriggerImmediately = true

	warrior.MakeProcTriggerAura(trigger)
}

func (warrior *Warrior) registerPrecision() {
	if warrior.Talents.Precision == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.Precision.Rank(warrior.Talents.Precision))
}

var bloodthirstRank = spellData.Bloodthirst.ByID(23894)

func (warrior *Warrior) registerBloodthirst() {
	if !warrior.Talents.Bloodthirst {
		return
	}

	apShare := bloodthirstRank.EffectN(2).Percent()

	config := spelldata.SpellConfig(&warrior.Unit, bloodthirstRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ClassSpellMask = SpellMaskBloodthirst

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := spell.MeleeAttackPower(target)*apShare + bloodthirstRank.DamageEffect().Roll(sim, core.CharacterLevel)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}

var piercingHowlRank = spellData.PiercingHowl.Highest()

// TODO: In-game test required if there's any threat interaction
func (warrior *Warrior) registerPiercingHowl() {
	if !warrior.Talents.PiercingHowl {
		return
	}

	config := spelldata.SpellConfig(&warrior.Unit, piercingHowlRank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty

	warrior.RegisterSpell(config)
}

var bloodCrazeHot = spellData.BloodCrazeTriggered.Highest()

func (warrior *Warrior) registerBloodCraze() {
	if warrior.Talents.BloodCraze == 0 {
		return
	}

	healthFraction := spellData.BloodCraze.EffectAt(1).FractionAt(warrior.Talents.BloodCraze)
	hitThreshold := spellData.BloodCraze.EffectAt(2).FractionAt(warrior.Talents.BloodCraze)
	tick := bloodCrazeHot.PeriodicEffect()

	config := spelldata.SpellConfig(&warrior.Unit, bloodCrazeHot, spelldata.Proc())
	config.ProcMask = core.ProcMaskSpellHealing
	config.DamageMultiplier = 1
	config.ThreatMultiplier = 1

	config.Hot = spelldata.DotConfig(bloodCrazeHot, tick)
	config.Hot.SelfOnly = true
	config.Hot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		healPerTick := warrior.MaxHealth() * healthFraction / float64(dot.ExpectedTickCount())
		dot.Spell.CalcAndDealPeriodicHealing(sim, target, healPerTick, dot.OutcomeTick)
	}

	bloodCraze := warrior.RegisterSpell(config)

	// One row, two listeners: 16487 decodes to hits taken and hits dealt at once, and the tooltip
	// gives each its own condition. Its crit hint covers only half of the taken one - a crit taken
	// or a hit over a share of maximum health - so both take the landed outcome and the condition
	// by hand, and the dealt one is narrowed to the Bloodthirst the tooltip names.
	apply := func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		bloodCraze.SelfHot().Apply(sim)
	}
	rank := spellData.BloodCraze.Rank(warrior.Talents.BloodCraze)

	taken := spelldata.ProcTrigger(&warrior.Character, rank, apply)
	taken.Name = "Blood Craze - Damage Taken"
	taken.Callback = core.CallbackOnSpellHitTaken
	taken.Outcome = core.OutcomeLanded
	taken.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
		return result.Outcome.Matches(core.OutcomeCrit) || result.Damage > warrior.MaxHealth()*hitThreshold
	}
	warrior.MakeProcTriggerAura(taken)

	dealt := spelldata.ProcTrigger(&warrior.Character, rank, apply)
	dealt.Name = "Blood Craze - Bloodthirst"
	dealt.Callback = core.CallbackOnSpellHitDealt
	dealt.Outcome = core.OutcomeLanded
	dealt.ClassSpellMask = SpellMaskBloodthirst
	warrior.MakeProcTriggerAura(dealt)
}

func (warrior *Warrior) registerRagingBlows() {
	if !warrior.Talents.RagingBlows {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.RagingBlows.Rank(1))
}

var deathWishRank = spellData.DeathWish.Highest()

func (warrior *Warrior) registerDeathWish() {
	if !warrior.Talents.DeathWish {
		return
	}

	deathWishAura := warrior.RegisterAura(spelldata.AuraConfig(deathWishRank))
	spelldata.ParseEffects(&warrior.Character, deathWishAura, deathWishRank)

	// Grants immunity to Fear effects, which the row states as A_MECHANIC_IMMUNITY and the parse
	// skips.
	deathWishAura.AttachFearImmunity()

	config := spelldata.SpellConfig(&warrior.Unit, deathWishRank,
		spelldata.Flags(core.SpellFlagCastWhileIncapacitated))

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		deathWishAura.Activate(sim)
		warrior.WaitUntil(sim, sim.CurrentTime+core.GCDDefault)
	}

	config.RelatedSelfBuff = deathWishAura

	deathWishSpell := warrior.RegisterSpell(config)

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: deathWishSpell,
		Type:  core.CooldownTypeDPS,
	})
}

func (warrior *Warrior) registerImprovedIntercept() {
	if warrior.Talents.ImprovedIntercept == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.ImprovedIntercept.Rank(warrior.Talents.ImprovedIntercept))
}

func (warrior *Warrior) registerImprovedCleave() {
	if warrior.Talents.ImprovedCleave == 0 {
		return
	}
	spelldata.ParseStatic(&warrior.Character, spellData.ImprovedCleave.Rank(warrior.Talents.ImprovedCleave))
}

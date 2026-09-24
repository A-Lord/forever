package core

import (
	"fmt"
	"strconv"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// applyRaidDebuffEffects applies all raid-level debuffs based on the provided Debuffs proto.
func applyDebuffEffects(target *Unit, targetIdx int, debuffs *proto.Debuffs, raid *proto.Raid) {

	if debuffs.BloodFrenzy {
		MakePermanent(BloodFrenzyAura(target, 2))
	}

	if debuffs.CurseOfElements != proto.TristateEffect_TristateEffectMissing {
		ranks := GetTristateValueInt32(debuffs.CurseOfElements, 0, 3)
		MakePermanent(CurseOfElementsAura(target, -1, ranks))
	}

	if debuffs.CurseOfRecklessness {
		MakePermanent(CurseOfRecklessnessAura(target, -1))
	}

	if debuffs.DemoralizingRoar != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(DemoralizingRoarAura(target, GetTristateValueInt32(debuffs.DemoralizingRoar, 0, 5)))
	}

	if debuffs.DemoralizingShout != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(DemoralizingShoutAura(target, 5, GetTristateValueInt32(debuffs.DemoralizingShout, 0, 5)))
	}

	if debuffs.ExposeWeaknessUptime > 0.0 {
		aura := ExposeWeaknessAura(target, func() float64 {
			return debuffs.ExposeWeaknessHunterAgility
		})
		ApplyFixedUptimeAura(aura, debuffs.ExposeWeaknessUptime, aura.Duration, 1)
	}

	if debuffs.FaerieFire != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(FaerieFireAura(target, TernaryFloat64(IsImproved(debuffs.FaerieFire), 3, 0)))
	}

	if debuffs.HemorrhageUptime > 0.0 {
		HemorrhageAura(target, debuffs.HemorrhageUptime)
	}

	if debuffs.GiftOfArthas {
		MakePermanent(GiftOfArthasAura(target))
	}

	if debuffs.HuntersMark != proto.TristateEffect_TristateEffectMissing {
		aura := HuntersMarkAura(target, GetTristateValueInt32(debuffs.HuntersMark, 0, 5))
		ApplyFixedUptimeAura(aura, 1, aura.Duration, 1)

		ScheduledAura(aura, PeriodicActionOptions{
			Period:   time.Second * 1,
			NumTicks: 5,
			Priority: ActionPriorityDOT,
			OnAction: func(sim *Simulation) {
				aura.Activate(sim)
				if aura.IsActive() {
					aura.SetStacks(sim, aura.GetStacks()+6)
				}
			},
		}, raid)
	}

	if debuffs.ImprovedScorch {
		aura := MakePermanent(ImprovedScorchAura(target))

		ScheduledAura(aura, PeriodicActionOptions{
			Period:          time.Millisecond * 1200,
			NumTicks:        5,
			TickImmediately: true,
			Priority:        ActionPriorityDOT, // High prio so it comes before actual warrior sunders.
			OnAction: func(sim *Simulation) {
				aura.Activate(sim)
				if aura.IsActive() {
					aura.AddStack(sim)
				}
			},
		}, raid)

	}

	if debuffs.JudgementOfTheCrusader {
		MakePermanent(JudgementOfTheCrusaderAura(target, JudgementOfTheCrusaderMaxRank))
	}

	if debuffs.InsectSwarm {
		MakePermanent((InsectSwarmAura(target)))
	}

	if debuffs.IsbUptime > 0.0 {
		ImprovedShadowBoltAura(target, debuffs.IsbUptime, 5)
	}

	if debuffs.JudgementOfLight {
		MakePermanent(JudgementOfLightAura(target, JudgementOfLightMaxRank))
	}

	if debuffs.JudgementOfWisdom {
		MakePermanent(JudgementOfWisdomAura(target, JudgementOfWisdomMaxRank))
	}

	if debuffs.Mangle {
		MakePermanent(MangleAura(target))
	}

	if debuffs.Misery {
		MakePermanent(MiseryAura(target, 5))
	}

	if debuffs.ScorpidSting {
		MakePermanent(ScorpidStingAura(target))
	}

	if debuffs.Screech {
		MakePermanent(ScreechAura(target))
	}

	if debuffs.ShadowEmbrace {
		MakePermanent(ShadowEmbraceAura(target, 5))
	}

	if debuffs.ShadowWeaving {
		aura := MakePermanent(ShadowWeavingAura(target))

		ScheduledAura(aura, PeriodicActionOptions{
			Period:          time.Millisecond * 1500,
			NumTicks:        5,
			TickImmediately: true,
			Priority:        ActionPriorityDOT,
			OnAction: func(sim *Simulation) {
				aura.Activate(sim)
				aura.AddStack(sim)

			},
		}, raid)

	}

	if debuffs.ExposeArmor != proto.TristateEffect_TristateEffectMissing {
		aura := MakePermanent(ExposeArmorAura(target, func() int32 { return 5 }, GetTristateValueInt32(debuffs.ExposeArmor, 0, 2)))

		ScheduledAura(aura, PeriodicActionOptions{
			Period:   time.Second * 10,
			NumTicks: 1,
			OnAction: func(sim *Simulation) {
				aura.Activate(sim)
			},
		}, raid)
	}

	if debuffs.SunderArmor {
		aura := MakePermanent(SunderArmorAura(target))

		ScheduledAura(aura, PeriodicActionOptions{
			Period:          GCDDefault,
			NumTicks:        5,
			TickImmediately: true,
			Priority:        ActionPriorityDOT, // High prio so it comes before actual warrior sunders.
			OnAction: func(sim *Simulation) {
				aura.Activate(sim)
				if aura.IsActive() {
					aura.AddStack(sim)
				}
			},
		}, raid)
	}

	if debuffs.WintersChill {
		MakePermanent(WintersChillAura(target, 5))
	}

	if debuffs.ThunderClap != proto.TristateEffect_TristateEffectMissing {
		MakePermanent(ThunderClapAura(target))
	}
}

func ScheduledAura(aura *Aura, options PeriodicActionOptions, raid *proto.Raid) {
	aura.OnReset = func(aura *Aura, sim *Simulation) {
		aura.Duration = NeverExpires
		StartPeriodicAction(sim, options)
	}
}

// Physical and Armor Related Debuffs
func BloodFrenzyAura(target *Unit, points int32) *Aura {
	return damageTakenDebuff(target, 0,
		"Blood Frenzy",
		29859,
		[]stats.SchoolIndex{stats.SchoolIndexPhysical},
		1+0.02*float64(points),
		NeverExpires,
	)
}

// Damage Taken Debuffs
func CurseOfElementsAura(target *Unit, casterIndex int32, ranks int32) *Aura {
	multiplier := 1.10 + 0.01*float64(ranks)

	aura := damageTakenDebuff(
		target,
		casterIndex,
		fmt.Sprintf("Curse of the Elements (%s)", Ternary(casterIndex == -1, "External", "Self")),
		27228,
		[]stats.SchoolIndex{
			stats.SchoolIndexArcane,
			stats.SchoolIndexFire,
			stats.SchoolIndexFrost,
			stats.SchoolIndexShadow,
		},
		multiplier,
		time.Minute*5,
	)
	aura.AttachStatsBuff(stats.Stats{
		stats.ArcaneResistance: -88,
		stats.FireResistance:   -88,
		stats.FrostResistance:  -88,
		stats.ShadowResistance: -88,
	})

	aura.NewExclusiveEffect("CurseOfElements", true, ExclusiveEffect{
		Priority: multiplier,
	})

	return aura
}

func CurseOfRecklessnessAura(target *Unit, casterIndex int32) *Aura {
	aura := statsDebuff(
		target,
		casterIndex,
		fmt.Sprintf("Curse of Recklessness (%s)", Ternary(casterIndex == -1, "External", "Self")),
		27226,
		stats.Stats{
			stats.Armor:       -800,
			stats.AttackPower: 135,
		},
		time.Minute*2,
	)

	aura.NewExclusiveEffect("CurseOfRecklessness", true, ExclusiveEffect{})

	return aura
}

func DemoralizingRoarAura(target *Unit, feralAggressionPoints int32) *Aura {
	apReduction := 248.0 * (1 + 0.08*float64(feralAggressionPoints))

	aura := target.GetOrRegisterAura(Aura{
		Label:    "Demoralizing Roar",
		ActionID: ActionID{SpellID: 26998},
		Duration: time.Second * 30,
	})

	effect := aura.NewExclusiveEffect(DemoralizingEffectCategory, true, ExclusiveEffect{
		Priority: apReduction,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.AttackPower, -ee.Priority)
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.AttackPower, ee.Priority)
		},
	})

	if effect.Priority < apReduction {
		effect.Priority = apReduction
	}

	return aura
}

func DemoralizingShoutAura(target *Unit, boomingVoicePoints int32, improvedDemoShoutPoints int32) *Aura {
	apReduction := 300.0 * (1 + 0.1*float64(improvedDemoShoutPoints))
	duration := time.Duration(float64(time.Second*30) * (1 + 0.1*float64(boomingVoicePoints)))

	aura := target.GetOrRegisterAura(Aura{
		Label:    "Demoralizing Shout",
		ActionID: ActionID{SpellID: 25203},
		Duration: duration,
	})

	effect := aura.NewExclusiveEffect(DemoralizingEffectCategory, true, ExclusiveEffect{
		Priority: apReduction,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.AttackPower, -ee.Priority)
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.AttackPower, ee.Priority)
		},
	})

	if effect.Priority < apReduction {
		effect.Priority = apReduction
	}
	if aura.Duration < duration {
		aura.Duration = duration
	}

	return aura
}

func SlowAura(target *Unit) *Aura {
	return castSlowReductionAura(target, "Slow", 31589, 1.5, time.Second*15)
}

func castSlowReductionAura(target *Unit, label string, spellID int32, multiplier float64, duration time.Duration) *Aura {
	aura := target.GetOrRegisterAura(Aura{Label: label, ActionID: ActionID{SpellID: spellID}, Duration: duration})
	aura.NewExclusiveEffect("CastSpdReduction", false, ExclusiveEffect{
		Priority: multiplier,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, 1/multiplier)
			ee.Aura.Unit.MultiplyRangedSpeed(sim, 1/multiplier)
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, multiplier)
			ee.Aura.Unit.MultiplyRangedSpeed(sim, multiplier)
		},
	})
	return aura
}

// A slow on casts alone, in the category Slow's cast and ranged slow takes: only the strongest applies.
func CastSpeedReductionEffect(aura *Aura, castTimeMultiplier float64) *ExclusiveEffect {
	return aura.NewExclusiveEffect("CastSpdReduction", false, ExclusiveEffect{
		Priority: castTimeMultiplier,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, 1/ee.Priority)
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, ee.Priority)
		},
	})
}

type ExposeWeaknessAgiFunc func() float64

func ExposeWeaknessAura(target *Unit, agilityFunc ExposeWeaknessAgiFunc) *Aura {
	aura := target.GetOrRegisterAura(Aura{
		Label:     "Expose Weakness",
		Tag:       "ExposeWeakness",
		ActionID:  ActionID{SpellID: 34503},
		Duration:  time.Second * 7,
		MaxStacks: 10000,
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.SetStacks(sim, int32(agilityFunc()*0.25))
		},
		OnStacksChange: func(aura *Aura, sim *Simulation, oldStacks int32, newStacks int32) {
			newValue := float64(newStacks - oldStacks)
			target.PseudoStats.BonusAttackPower += newValue
			target.PseudoStats.BonusRangedAttackPower += newValue
		},
	})

	return aura

}

func FaerieFireAura(target *Unit, improvedPoints float64) *Aura {
	armorValue := 610.0

	aura := target.GetOrRegisterAura(Aura{
		Label:    "Faerie Fire",
		ActionID: ActionID{SpellID: 26993},
		Duration: time.Second * 40,
	})

	effect := aura.NewExclusiveEffect("FaerieFireAura", true, ExclusiveEffect{
		Priority: improvedPoints,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.Armor, -armorValue)
			ee.Aura.Unit.PseudoStats.ReducedPhysicalHitTakenChance -= ee.Priority
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.Armor, armorValue)
			ee.Aura.Unit.PseudoStats.ReducedPhysicalHitTakenChance += ee.Priority
		},
	})

	if effect.Priority < improvedPoints {
		effect.Priority = improvedPoints
	}

	return aura
}

func GiftOfArthasAura(target *Unit) *Aura {
	var effect *ExclusiveEffect
	aura := target.GetOrRegisterAura(Aura{
		Label:    "Gift of Arthas",
		ActionID: ActionID{SpellID: 11374},
		Duration: time.Minute * 3,
		OnGain: func(aura *Aura, sim *Simulation) {
			effect.SetPriority(sim, 8)
		},
	})

	effect = aura.NewExclusiveEffect("GiftOfArthasAura", true, ExclusiveEffect{
		Priority: 0,
		OnGain: func(ee *ExclusiveEffect, s *Simulation) {
			ee.Aura.Unit.PseudoStats.BonusPhysicalDamageTaken += ee.Priority
		},
		OnExpire: func(ee *ExclusiveEffect, s *Simulation) {
			ee.Aura.Unit.PseudoStats.BonusPhysicalDamageTaken -= ee.Priority
		},
	})

	return aura
}

func HemorrhageAura(target *Unit, uptime float64) *Aura {
	hasAura := target.HasAura("Hemorrhage")
	aura := target.GetOrRegisterAura(Aura{
		Label:    "Hemorrhage",
		ActionID: ActionID{SpellID: 26864},
		Duration: time.Second * 15,
	})

	if !hasAura {
		aura.AttachAdditivePseudoStatBuff(&target.PseudoStats.BonusPhysicalDamageTaken, 42)
		ApplyFixedUptimeAura(aura, uptime, aura.Duration, 1)
	}

	return aura
}

func HuntersMarkAura(target *Unit, improved int32) *Aura {
	initialBonus := 110.0
	bonusPerStack := 11.0
	meleeBonus := initialBonus * 0.2 * float64(improved)

	var effect *ExclusiveEffect
	aura := target.RegisterAura(Aura{
		Label:     "Hunters Mark",
		Tag:       "HuntersMark",
		ActionID:  ActionID{SpellID: 14325},
		Duration:  time.Minute * 2,
		MaxStacks: 30,
		OnStacksChange: func(aura *Aura, sim *Simulation, oldStacks int32, newStacks int32) {
			effect.SetPriority(sim, initialBonus+bonusPerStack*float64(newStacks))
		},
	})

	effect = aura.NewExclusiveEffect("HuntersMark", true, ExclusiveEffect{
		Priority: initialBonus,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			if improved > 0 {
				target.PseudoStats.BonusAttackPower += meleeBonus
			}
			target.PseudoStats.BonusRangedAttackPower += ee.Priority
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			if improved > 0 {
				target.PseudoStats.BonusAttackPower -= meleeBonus
			}
			target.PseudoStats.BonusRangedAttackPower -= ee.Priority
		},
	})

	return aura
}

func ImprovedScorchAura(target *Unit) *Aura {
	fireBonus := 0.03
	var effect *ExclusiveEffect

	aura := target.GetOrRegisterAura(Aura{
		Label:     "Improved Scorch",
		ActionID:  ActionID{SpellID: 12873},
		Duration:  time.Second * 30,
		MaxStacks: 5,
		OnStacksChange: func(aura *Aura, sim *Simulation, oldStacks int32, newStacks int32) {
			effect.SetPriority(sim, 1.0+fireBonus*float64(newStacks))
		},
	})

	effect = aura.NewExclusiveEffect("ImprovedScorch", false, ExclusiveEffect{
		Priority: 1,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			target.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire] *= ee.Priority
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			target.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire] /= ee.Priority
		},
	})

	return aura
}

// One rank of a judgement debuff: the spell the target shows and the number its row states. The
// paladin registers a rank per row; the debuff panel applies the max rank, carried by the *MaxRank
// values with Rank 0.
type JudgementRank struct {
	SpellID int32
	Rank    int32
	Value   float64
}

var (
	JudgementOfTheCrusaderMaxRank = JudgementRank{SpellID: 20303, Value: 161}
	JudgementOfLightMaxRank       = JudgementRank{SpellID: 20346, Value: 61}
	JudgementOfWisdomMaxRank      = JudgementRank{SpellID: 20355, Value: 59}
)

const JudgementDuration = time.Second * 40

// Every judgement debuff carries the tag, so an effect that refreshes "all Judgement effects on the
// target" can find them whoever put them up.
const JudgementAuraTag = "JudgementAura"

// The client says the judgements proc on a chance; the sim keeps the 50% the TBC sim settled on
// until Forever testing says otherwise.
const judgementProcChance = 0.5

func judgementLabel(name string, rank JudgementRank) string {
	if rank.Rank > 0 {
		return fmt.Sprintf("%s Rank %d", name, rank.Rank)
	}
	return name
}

// Judgement of the Crusader raises the Holy damage the target takes by a flat amount. Every rank
// and every paladin share one exclusive category, so the strongest active one is the one that
// counts.
func JudgementOfTheCrusaderAura(target *Unit, rank JudgementRank) *Aura {
	bonus := rank.Value
	label := judgementLabel("Judgement of the Crusader", rank)
	if target.HasAura(label) {
		return target.GetAura(label)
	}

	aura := target.GetOrRegisterAura(Aura{
		Label:    label,
		ActionID: ActionID{SpellID: rank.SpellID},
		Tag:      JudgementAuraTag,
		Duration: JudgementDuration,
	})

	aura.NewExclusiveEffect("Judgement of the Crusader", true, ExclusiveEffect{
		Priority: bonus,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			target.PseudoStats.SchoolBonusSpellDamage[stats.SchoolIndexHoly] += bonus
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			target.PseudoStats.SchoolBonusSpellDamage[stats.SchoolIndexHoly] -= bonus
		},
	})

	return aura
}

func ImprovedShadowBoltAura(target *Unit, uptime float64, points int32) *Aura {
	bonus := 0.04 * float64(points)
	multiplier := 1 + bonus

	config := Aura{
		Label:     "ImprovedShadowBolt-" + strconv.Itoa(int(points)),
		Tag:       "ImprovedShadowBolt",
		ActionID:  ActionID{SpellID: 17800},
		Duration:  time.Second * 12,
		MaxStacks: 4,
	}

	if uptime == 0 {
		config.OnSpellHitTaken = func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if !spell.SpellSchool.Matches(SpellSchoolShadow) || !result.Landed() || result.Damage == 0 || !spell.ProcMask.Matches(ProcMaskSpellDamage) {
				return
			}
			aura.RemoveStack(sim)
		}
	}

	hasAura := target.HasAura(config.Label)
	aura := target.GetOrRegisterAura(config)
	if !hasAura {
		aura.AttachMultiplicativePseudoStatBuff(&target.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexShadow], multiplier)
		ApplyFixedUptimeAura(aura, uptime, aura.Duration, 1)
	}

	return aura
}

func InsectSwarmAura(target *Unit) *Aura {
	return statsDebuff(
		target,
		0,
		"Insect Swarm",
		27013,
		stats.Stats{
			stats.PhysicalHitPercent: -2,
			stats.SpellHitPercent:    -2,
		},
		time.Second*12,
	)
}

// Judgement of Light heals whoever lands a melee hit on the target.
func JudgementOfLightAura(target *Unit, rank JudgementRank) *Aura {
	healthMetrics := target.NewHealthMetrics(ActionID{SpellID: rank.SpellID})
	heal := rank.Value

	return target.GetOrRegisterAura(Aura{
		Label:    judgementLabel("Judgement of Light", rank),
		ActionID: ActionID{SpellID: rank.SpellID},
		Tag:      JudgementAuraTag,
		Duration: JudgementDuration,
		OnSpellHitTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if !spell.ProcMask.Matches(ProcMaskMelee) || !result.Landed() {
				return
			}

			if sim.Proc(judgementProcChance, "Judgement of Light - Heal") {
				spell.Unit.GainHealth(sim, heal, healthMetrics)
			}
		},
	})
}

// Judgement of Wisdom restores mana to whoever lands an attack or spell on the target.
func JudgementOfWisdomAura(target *Unit, rank JudgementRank) *Aura {
	actionID := ActionID{SpellID: rank.SpellID}
	mana := rank.Value
	label := judgementLabel("Judgement of Wisdom", rank)
	if target.HasAura(label) {
		return target.GetAura(label)
	}

	return target.GetOrRegisterAura(Aura{
		Label:    label,
		ActionID: actionID,
		Tag:      JudgementAuraTag,
		Duration: JudgementDuration,
	}).AttachProcTrigger(ProcTrigger{
		ProcChance: judgementProcChance,
		ProcMask:   ProcMaskDirect,
		Callback:   CallbackOnSpellHitTaken,
		Handler: func(sim *Simulation, spell *Spell, result *SpellResult) {
			// Melee claim that wisdom can proc on misses.
			if !spell.ProcMask.Matches(ProcMaskMeleeOrRanged) && !result.Landed() {
				return
			}

			unit := spell.Unit
			if unit.HasManaBar() {
				if unit.JowManaMetrics == nil {
					unit.JowManaMetrics = unit.NewManaMetrics(actionID)
				}
				unit.AddMana(sim, mana, unit.JowManaMetrics)
			}
		},
	})
}

func MangleAura(target *Unit) *Aura {
	multiplier := 1.3

	aura := target.GetOrRegisterAura(Aura{
		Label:    "Mangle",
		ActionID: ActionID{SpellID: 33876},
		Duration: time.Second * 12,
	})

	aura.NewExclusiveEffect("Mangle", true, ExclusiveEffect{
		Priority: multiplier,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.PseudoStats.PeriodicPhysicalDamageTakenMultiplier *= ee.Priority
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.PseudoStats.PeriodicPhysicalDamageTakenMultiplier /= ee.Priority
		},
	})

	return aura
}

func MiseryAura(target *Unit, ranks int32) *Aura {
	multiplier := 1.0 + 0.01*float64(ranks)
	schools := []stats.SchoolIndex{
		stats.SchoolIndexArcane, stats.SchoolIndexFire, stats.SchoolIndexFrost,
		stats.SchoolIndexHoly, stats.SchoolIndexNature, stats.SchoolIndexShadow,
	}
	aura := target.GetOrRegisterAura(Aura{
		Label:    "Misery",
		ActionID: ActionID{SpellID: 33195},
		Duration: NeverExpires,
	})
	effect := aura.NewExclusiveEffect("MiseryBonus", true, ExclusiveEffect{
		Priority: multiplier,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			for _, school := range schools {
				target.PseudoStats.SchoolDamageTakenMultiplier[school] *= ee.Priority
			}
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			for _, school := range schools {
				target.PseudoStats.SchoolDamageTakenMultiplier[school] /= ee.Priority
			}
		},
	})
	if effect.Priority < multiplier {
		effect.Priority = multiplier
	}
	return aura
}

func ScorpidStingAura(target *Unit) *Aura {
	return statsDebuff(target, 0, "Scorpid Sting", 3043, stats.Stats{stats.PhysicalHitPercent: -5.0}, time.Second*20)
}

func ScreechAura(target *Unit) *Aura {
	return statsDebuff(target, 0, "Screech", 27051, stats.Stats{stats.AttackPower: -210}, time.Second*4)
}

func ShadowEmbraceAura(target *Unit, ranks int32) *Aura {
	return damageDealtDebuff(target, "Shadow Embrace", 32394, []stats.SchoolIndex{stats.SchoolIndexPhysical}, 1.0-(.01*float64(ranks)), NeverExpires)
}

func ShadowWeavingAura(target *Unit) *Aura {
	const shadowBonus = 0.02
	var effect *ExclusiveEffect

	aura := target.GetOrRegisterAura(Aura{
		Label:     "Shadow Weaving",
		ActionID:  ActionID{SpellID: 15334},
		Duration:  time.Second * 15,
		MaxStacks: 5,
		OnStacksChange: func(aura *Aura, sim *Simulation, oldStacks int32, newStacks int32) {
			effect.SetPriority(sim, 1.0+shadowBonus*float64(newStacks))
		},
	})

	effect = aura.NewExclusiveEffect("ShadowWeaving", false, ExclusiveEffect{
		Priority: 1.0,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			target.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexShadow] *= ee.Priority
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			target.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexShadow] /= ee.Priority
		},
	})

	return aura
}

func StormstrikeAura(target *Unit, uptime float64) *Aura {
	multiplier := 1.20
	hasAura := target.HasAura("Stormstrike")
	aura := damageTakenDebuff(target, 0, "Stormstrike", 17364, []stats.SchoolIndex{stats.SchoolIndexNature}, multiplier, time.Second*12)

	if !hasAura {
		ApplyFixedUptimeAura(aura, uptime, aura.Duration, 1)
	}

	return aura
}

var MajorArmorReductionEffectCategory = "MajorArmorReduction"

// Demoralizing Roar and Demoralizing Shout are mutually exclusive; other AP
// reduction debuffs (Screech, Curse of Recklessness, ...) stack with them.
var DemoralizingEffectCategory = "Demoralizing"

func ExposeArmorAura(target *Unit, getComboPoints func() int32, talents int32) *Aura {

	var effect *ExclusiveEffect
	aura := target.GetOrRegisterAura(Aura{
		Label:    "Expose Armor",
		ActionID: ActionID{SpellID: 26866},
		Duration: time.Second * 30,
		OnGain: func(aura *Aura, sim *Simulation) {
			eaValue := 410.0 * float64(getComboPoints())
			eaValue *= 1.0 + 0.25*float64(talents)
			effect.SetPriority(sim, eaValue)
		},
	})

	effect = aura.NewExclusiveEffect(MajorArmorReductionEffectCategory, true, ExclusiveEffect{
		Priority: 0,
		OnGain: func(ee *ExclusiveEffect, s *Simulation) {
			ee.Aura.Unit.AddStatDynamic(s, stats.Armor, -ee.Priority)
		},
		OnExpire: func(ee *ExclusiveEffect, s *Simulation) {
			ee.Aura.Unit.AddStatDynamic(s, stats.Armor, ee.Priority)
		},
	})

	return aura

}

func SunderArmorAura(target *Unit) *Aura {
	var effect *ExclusiveEffect
	aura := target.GetOrRegisterAura(Aura{
		Label:     "Sunder Armor",
		ActionID:  ActionID{SpellID: 25225},
		Duration:  time.Second * 30,
		MaxStacks: 5,
		OnStacksChange: func(aura *Aura, sim *Simulation, oldStacks int32, newStacks int32) {
			effect.SetPriority(sim, -520*float64(newStacks))
		},
	})

	effect = aura.NewExclusiveEffect(MajorArmorReductionEffectCategory, true, ExclusiveEffect{
		Priority: 0,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.Armor, ee.Priority)
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.AddStatDynamic(sim, stats.Armor, -ee.Priority)
		},
	})

	return aura
}

func WintersChillAura(target *Unit, startingStacks int32) *Aura {
	critBonus := 2.0

	dynamicMods := make(map[int32]*SpellMod, len(target.Env.AllUnits))

	for _, unit := range target.Env.AllUnits {
		if unit.Type == PlayerUnit || unit.Type == PetUnit {
			dynamicMods[unit.UnitIndex] = unit.AddDynamicMod(SpellModConfig{
				Kind:       SpellMod_BonusCrit_Percent,
				FloatValue: 0,
				School:     SpellSchoolFrost,
			})
		}
	}

	return target.GetOrRegisterAura(Aura{
		Label:     "Winter's Chill",
		ActionID:  ActionID{SpellID: 28595},
		Duration:  time.Second * 15,
		MaxStacks: 5,
		OnGain: func(aura *Aura, sim *Simulation) {
			aura.SetStacks(sim, startingStacks)
		},
		OnStacksChange: func(aura *Aura, sim *Simulation, oldStacks int32, newStacks int32) {
			for _, unit := range sim.AllUnits {
				if unit.Type == PlayerUnit || unit.Type == PetUnit {
					dynamicMods[unit.UnitIndex].Activate()
					dynamicMods[unit.UnitIndex].UpdateFloatValue(critBonus * float64(newStacks))
				}
			}
		},
	})
}

// Spell 11581: -20% melee haste for 30 s on every rank, the time between attacks 20% longer;
// Improved Thunder Clap discounts the rage cost and leaves the slow alone.
func ThunderClapAura(target *Unit) *Aura {
	aura := target.GetOrRegisterAura(Aura{
		Label:    "Thunder Clap",
		ActionID: ActionID{SpellID: 11581},
		Duration: time.Second * 30,
	})
	AtkSpeedReductionEffect(aura, SlowedTimeMultiplier(-20))
	return aura
}

// A slow the client states as a negative speed percentage makes the time between attacks, or a cast
// time, that much longer: -20 is 20% longer, which divides the speed by 1.2.
func SlowedTimeMultiplier(speedPercent float64) float64 {
	return 1 - speedPercent/100
}

// The priority is the slow, so SetPriority from an aura's OnGain can rescale it: the Conqueror's
// set raises Thunder Clap's by half.
func AtkSpeedReductionEffect(aura *Aura, speedMultiplier float64) *ExclusiveEffect {
	return aura.NewExclusiveEffect("AtkSpdReduction", false, ExclusiveEffect{
		Priority: speedMultiplier,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.MultiplyAttackSpeed(sim, 1/ee.Priority)
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.MultiplyAttackSpeed(sim, ee.Priority)
		},
	})
}

func damageTakenDebuff(target *Unit, casterIndex int32, label string, spellID int32, schools []stats.SchoolIndex, multiplier float64, duration time.Duration) *Aura {
	actionID := ActionID{SpellID: spellID}
	if casterIndex != 0 {
		actionID = actionID.WithTag(casterIndex)
	}

	return target.GetOrRegisterAura(Aura{
		Label:    label,
		ActionID: actionID,
		Duration: duration,
		OnGain: func(aura *Aura, sim *Simulation) {
			for _, school := range schools {
				target.PseudoStats.SchoolDamageTakenMultiplier[school] *= multiplier
			}
		},

		OnExpire: func(aura *Aura, sim *Simulation) {
			for _, school := range schools {
				target.PseudoStats.SchoolDamageTakenMultiplier[school] /= multiplier
			}
		},
	})
}

func damageDealtDebuff(target *Unit, label string, spellID int32, schools []stats.SchoolIndex, multiplier float64, duration time.Duration) *Aura {
	return target.GetOrRegisterAura(Aura{
		Label:    label,
		ActionID: ActionID{SpellID: spellID},
		Duration: duration,

		OnGain: func(aura *Aura, sim *Simulation) {
			for _, school := range schools {
				target.PseudoStats.SchoolDamageDealtMultiplier[school] *= multiplier
			}
		},

		OnExpire: func(aura *Aura, sim *Simulation) {
			for _, school := range schools {
				target.PseudoStats.SchoolDamageDealtMultiplier[school] /= multiplier
			}
		},
	})
}

func statsDebuff(target *Unit, casterIndex int32, label string, spellID int32, stats stats.Stats, duration time.Duration) *Aura {
	if duration == 0 {
		duration = time.Second * 30
	}

	actionID := ActionID{SpellID: spellID}
	if casterIndex != 0 {
		actionID = actionID.WithTag(casterIndex)
	}

	aura := target.GetAuraByID(actionID)
	if aura != nil {
		return aura
	}

	return target.RegisterAura(Aura{
		Label:    label,
		ActionID: actionID,
		Duration: duration,
	}).AttachStatsBuff(stats)
}

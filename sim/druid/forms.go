package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

type DruidForm uint8

const (
	Humanoid DruidForm = 1 << iota
	Bear
	Cat
	Moonkin
	Tree
	DireBear
)

func (form DruidForm) Matches(other DruidForm) bool {
	return (form & other) != 0
}

func (druid *Druid) setForm(form DruidForm) {
	switch {
	case form.Matches(DireBear):
		// Dire Bear keeps the bear bit so a bear-only cast and InForm(Bear) both match.
		druid.form = Bear | DireBear
		druid.ShapeshiftForm = shapeshiftOf(spellData.DireBearForm.Highest())
	case form.Matches(Bear):
		druid.form = Bear
		druid.ShapeshiftForm = shapeshiftOf(spellData.BearForm.Highest())
	case form.Matches(Cat):
		druid.form = Cat
		druid.ShapeshiftForm = shapeshiftOf(spellData.CatForm.Highest())
	case form.Matches(Moonkin):
		druid.form = Moonkin
		druid.ShapeshiftForm = shapeshiftOf(spellData.MoonkinForm.Highest())
	default:
		druid.form = Humanoid
		druid.ShapeshiftForm = 0
	}
}

// The shapeshift aura's misc value is the game files' form id.
func shapeshiftOf(rank *spelldata.Spell) dbcenums.ShapeshiftForm {
	return dbcenums.ShapeshiftForm(rank.EffectN(1).Misc)
}

func (druid *Druid) InForm(form DruidForm) bool {
	return druid.form.Matches(form)
}

func (druid *Druid) leaveForm(sim *core.Simulation) {
	switch {
	case druid.CatFormAura != nil && druid.CatFormAura.IsActive():
		druid.CatFormAura.Deactivate(sim)
	case druid.DireBearFormAura != nil && druid.DireBearFormAura.IsActive():
		druid.DireBearFormAura.Deactivate(sim)
	case druid.BearFormAura != nil && druid.BearFormAura.IsActive():
		druid.BearFormAura.Deactivate(sim)
	case druid.MoonkinFormAura != nil && druid.MoonkinFormAura.IsActive():
		druid.MoonkinFormAura.Deactivate(sim)
	}
}

func (druid *Druid) ClearForm(sim *core.Simulation) {
	druid.leaveForm(sim)
	druid.setForm(Humanoid)
	if druid.HasManaBar() {
		druid.SetCurrentPowerBar(core.ManaBar)
	}
}

func (druid *Druid) GetCatWeapon() core.Weapon {
	unscaledWeapon := druid.WeaponFromMainHand()
	return core.Weapon{
		BaseDamageMin:        unscaledWeapon.BaseDamageMin / unscaledWeapon.SwingSpeed,
		BaseDamageMax:        unscaledWeapon.BaseDamageMax / unscaledWeapon.SwingSpeed,
		SwingSpeed:           1.0,
		NormalizedSwingSpeed: 1.0,
		AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
		MaxRange:             core.MaxMeleeRange,
	}
}

func (druid *Druid) GetBearWeapon() core.Weapon {
	unscaledWeapon := druid.WeaponFromMainHand()
	return core.Weapon{
		BaseDamageMin:        unscaledWeapon.BaseDamageMin / unscaledWeapon.SwingSpeed * 2.5,
		BaseDamageMax:        unscaledWeapon.BaseDamageMax / unscaledWeapon.SwingSpeed * 2.5,
		SwingSpeed:           2.5,
		NormalizedSwingSpeed: 2.5,
		AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
		MaxRange:             core.MaxMeleeRange,
	}
}

func (druid *Druid) RegisterCatFormAura() {
	form := spellData.CatForm.Highest()
	druid.CatFormAura = druid.registerFormAura(form, druid.StartingForm.Matches(Cat))
	spelldata.ParseEffects(&druid.Character, druid.CatFormAura, spellData.CatFormTriggered.Highest())

	druid.CatFormAura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
		druid.setForm(Cat)
		druid.SetCurrentPowerBar(core.EnergyBar)
		if !druid.Env.MeasuringStats {
			druid.AutoAttacks.SetMH(druid.GetCatWeapon())
		}
		druid.restoreCatEnergy(sim)
	})
	druid.CatFormAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		druid.rememberCatEnergy(sim)
		druid.breakProwl(sim)
		druid.restoreHumanoidBody()
	})
	druid.CatFormAura.ApplyOnReset(func(_ *core.Aura, _ *core.Simulation) {
		druid.furorStoredValid = false
		druid.furorStoredEnergy = 0
		druid.furorLeftCatAt = 0
		druid.furorBearTime = 0
	})
}

func (druid *Druid) RegisterBearFormAura() {
	druid.BearFormAura = druid.registerBearAura(spellData.BearForm.Highest(), spellData.BearFormTriggered.Highest(), Bear, druid.StartingForm == Bear)
	druid.DireBearFormAura = druid.registerBearAura(spellData.DireBearForm.Highest(), spellData.DireBearFormTriggered.Highest(), DireBear, druid.StartingForm.Matches(DireBear))
}

func (druid *Druid) registerBearAura(form *spelldata.Spell, passive *spelldata.Spell, kind DruidForm, starting bool) *core.Aura {
	aura := druid.registerFormAura(form, starting)
	// A_MOD_BONUS_ARMOR_PCT is not parsed. Armor from items, health, attack power and the Faerie
	// Fire spell mods are.
	// TODO: Ingame research needed -- effect 4 of Bear Form and Dire Bear Form is aura 668, which core
	// does not name.
	spelldata.ParseEffects(&druid.Character, aura, passive)
	aura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
		druid.setForm(kind)
		druid.SetCurrentPowerBar(core.RageBar)
		if !druid.Env.MeasuringStats {
			druid.AutoAttacks.SetMH(druid.GetBearWeapon())
		}
		druid.furorRageOnEnterBear(sim)
	})
	aura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		druid.furorBearTime += sim.CurrentTime - aura.StartedAt()
		druid.restoreHumanoidBody()
	})
	return aura
}

func (druid *Druid) registerFormAura(form *spelldata.Spell, starting bool) *core.Aura {
	return druid.RegisterAura(spelldata.AuraConfig(form, func(aura *core.Aura) {
		if starting {
			aura.BuildPhase = core.CharacterBuildPhaseBase
		}
	}))
}

func (druid *Druid) restoreHumanoidBody() {
	if druid.Env != nil && !druid.Env.MeasuringStats && druid.AutoAttacks.MH() != nil {
		druid.AutoAttacks.SetMH(druid.WeaponFromMainHand())
	}
	if druid.HasManaBar() {
		druid.SetCurrentPowerBar(core.ManaBar)
	}
	druid.setForm(Humanoid)
}

func (druid *Druid) registerCatFormSpell() {
	druid.CatForm = druid.registerShapeshiftSpell(spellData.CatForm.Highest(), druid.CatFormAura)
}

func (druid *Druid) registerBearFormSpell() {
	druid.BearForm = druid.registerShapeshiftSpell(spellData.BearForm.Highest(), druid.BearFormAura)
	druid.DireBearForm = druid.registerShapeshiftSpell(spellData.DireBearForm.Highest(), druid.DireBearFormAura)
}

func (druid *Druid) registerShapeshiftSpell(rank *spelldata.Spell, aura *core.Aura) *DruidSpell {
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return aura != nil && !aura.IsActive()
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		druid.leaveForm(sim)
		aura.Activate(sim)
	}
	return druid.RegisterSpell(config)
}

// TODO: Verify a companion-triggered swing reset before adding a snake-weaving option.

// furorFraction is the rage chance, the tooltip's $m1: 0.2 at 1/5 through 1 at 5/5.
func (druid *Druid) furorFraction() float64 {
	return spellData.Furor.EffectAt(1).FractionAt(druid.Talents.Furor)
}

func (druid *Druid) rememberCatEnergy(sim *core.Simulation) {
	if !druid.HasEnergyBar() {
		return
	}
	druid.furorStoredEnergy = druid.CurrentEnergy()
	druid.furorStoredValid = true
	druid.furorLeftCatAt = sim.CurrentTime
	druid.furorBearTime = 0
}

// restoreCatEnergy is the tooltip's rule, priced with $m2: the energy left in Cat Form plus 10 for
// each second spent outside Bear, Cat and Dire Bear Form, up to 100, all scaled by the rank. With the
// talent untaken the cat comes back with none.
//
// TODO: Ingame research needed -- whether entering Bear Form loses all of it, and whether energy outside
// Cat Form regenerates at the Cat Form rate at 5/5. Tooltip 17056 states neither.
func (druid *Druid) restoreCatEnergy(sim *core.Simulation) {
	if !druid.HasEnergyBar() || druid.Env.MeasuringStats {
		return
	}
	// The energy bar ticks in every form, so the first entry of an iteration keeps the pull's bar.
	// TODO: Ingame research needed -- what death or cancelling the form does to the stored energy, and
	// what a first entry later in the fight starts with.
	if !druid.furorStoredValid {
		return
	}
	seconds := max(0, (sim.CurrentTime - druid.furorLeftCatAt - druid.furorBearTime).Seconds())
	fraction := spellData.Furor.EffectAt(2).FractionAt(druid.Talents.Furor)
	gained := fraction * min(druid.furorStoredEnergy+10*seconds, 100)
	delta := gained - druid.CurrentEnergy()
	if delta > 0 {
		druid.AddEnergy(sim, delta, druid.furorEnergyMetrics())
	} else if delta < 0 {
		druid.SpendEnergy(sim, -delta, druid.furorEnergyMetrics())
	}
}

func (druid *Druid) furorRageOnEnterBear(sim *core.Simulation) {
	chance := druid.furorFraction()
	if chance == 0 || !druid.HasRageBar() || druid.Env.MeasuringStats {
		return
	}
	if !sim.Proc(chance, "Furor") {
		return
	}
	rage := spellData.FurorTriggered.Highest().EnergizeEffect().Tenths()
	druid.AddRage(sim, rage, druid.furorRageMetrics())
}

func (druid *Druid) furorEnergyMetrics() *core.ResourceMetrics {
	if druid.furorEnergyMetric == nil {
		druid.furorEnergyMetric = druid.NewEnergyMetrics(core.ActionID{SpellID: spellData.Furor.Highest().ID})
	}
	return druid.furorEnergyMetric
}

func (druid *Druid) furorRageMetrics() *core.ResourceMetrics {
	if druid.furorRageMetric == nil {
		druid.furorRageMetric = druid.NewRageMetrics(core.ActionID{SpellID: spellData.FurorTriggered.Highest().ID})
	}
	return druid.furorRageMetric
}

func (druid *Druid) RegisterMoonkinFormAura() {
	if !druid.Talents.MoonkinForm {
		return
	}
	druid.MoonkinFormAura = druid.registerFormAura(spellData.MoonkinForm.Highest(), druid.StartingForm.Matches(Moonkin))
	// Moonkin Form (Passive) 24905 states the form's armor from items; its A_MOD_BONUS_ARMOR_PCT has no sim
	// kind and is reported as skipped. The party crit of Moonkin Aura 24907 is the raid buff AddPartyBuffs
	// turns on.
	spelldata.ParseEffects(&druid.Character, druid.MoonkinFormAura, spellData.MoonkinFormTriggered.ByID(24905))
	druid.MoonkinFormAura.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
		druid.setForm(Moonkin)
		druid.SetCurrentPowerBar(core.ManaBar)
	})
	druid.MoonkinFormAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		druid.restoreHumanoidBody()
	})
}

func (druid *Druid) RegisterMoonkinFormSpell() {
	if druid.MoonkinFormAura == nil {
		return
	}
	druid.MoonkinForm = druid.registerShapeshiftSpell(spellData.MoonkinForm.Highest(), druid.MoonkinFormAura)
}

package paladin

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
)

// Only one seal is up at a time.
const SealCategory = "PaladinSeal"

// One rank of one seal: the castable spell's row, the aura the cast puts up, the judgement it
// unleashes, and the per-hit effect Twist of Light's Echo can replay.
type sealConfig struct {
	row       shared.SpellData
	classMask int64
	spell     *core.Spell
	aura      *core.Aura
	judgement *core.Spell

	// Twist of Light: the Echo this seal leaves when replaced (0 for the seals that leave none),
	// and what the Echo does to the target of the next melee attack (nil where the seal's effect
	// has no place in the sim).
	echoID int32
	echo   func(sim *core.Simulation, target *core.Unit)
}

// The Echoes Twist of Light names, one per seal.
const (
	echoOfCommandID       = 1311703
	echoOfFuryID          = 1311701
	echoOfRighteousnessID = 1311704
	echoOfJusticeID       = 1311705
)

// One Echo: its aura on the paladin, and the seal rank whose effect it holds.
type sealEcho struct {
	aura *core.Aura
	seal *sealConfig
}

// The seals' rows as the client states them. Seal of Command is a talent and registers from
// registerTalentSpells.
var (
	SealOfRighteousnessRankMap = spellData.SealOfRighteousness
	SealOfCommandRankMap       = spellData.SealOfCommand
	SealOfLightRankMap         = spellData.SealOfLight
	SealOfWisdomRankMap        = spellData.SealOfWisdom
	SealOfJusticeRankMap       = spellData.SealOfJustice
	SealOfTheCrusaderRankMap   = spellData.SealOfTheCrusader
	SealOfFuryRankMap          = spellData.SealOfFury
)

func (paladin *Paladin) registerSeals() {
	SealOfRighteousnessRankMap.RegisterAll(paladin.registerSealOfRighteousness)
	SealOfLightRankMap.RegisterAll(paladin.registerSealOfLight)
	SealOfWisdomRankMap.RegisterAll(paladin.registerSealOfWisdom)
	SealOfJusticeRankMap.RegisterAll(paladin.registerSealOfJustice)
	SealOfTheCrusaderRankMap.RegisterAll(paladin.registerSealOfTheCrusader)
	SealOfFuryRankMap.RegisterAll(paladin.registerSealOfFury)
}

func sealLabel(name string, paladin *Paladin, row shared.SpellData) string {
	return fmt.Sprintf("%s%s Rank %d", name, paladin.Label, row.Rank)
}

// Every seal aura joins the seal category, so casting one seal drops the other.
func (paladin *Paladin) makeSealExclusive(aura *core.Aura) *core.Aura {
	aura.NewExclusiveEffect(SealCategory, true, core.ExclusiveEffect{})
	return aura
}

// The castable seal spell: instant, on the GCD, priced as the row says.
func (paladin *Paladin) registerSealSpell(cfg *sealConfig) *core.Spell {
	cfg.spell = paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: cfg.row.SpellID},
		SpellSchool:    cfg.row.SpellSchool,
		DefenseType:    cfg.row.DefenseType,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: cfg.classMask,
		Rank:           cfg.row.Rank,

		ManaCost: manaCost(cfg.row),
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: cfg.row.GCD,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			paladin.applySeal(sim, cfg)
		},

		RelatedSelfBuff: cfg.aura,
	})
	return cfg.spell
}

func (paladin *Paladin) applySeal(sim *core.Simulation, cfg *sealConfig) {
	previous := paladin.currentSeal
	if previous != nil && previous != cfg && previous.aura.IsActive() {
		if echo := paladin.echoes[previous.echoID]; echo != nil {
			echo.seal = previous
			echo.aura.Activate(sim)
		}
	}

	paladin.currentSeal = cfg
	// The seal category drops whatever seal was up; recasting the same seal refreshes it.
	cfg.aura.Activate(sim)
}

// The seal the paladin is under, if any.
func (paladin *Paladin) activeSeal() *sealConfig {
	if paladin.currentSeal != nil && paladin.currentSeal.aura.IsActive() {
		return paladin.currentSeal
	}
	return nil
}

// A seal's per-hit damage lands a batch window after the swing that fired it, the way the game
// resolves the proc after the melee hit.
func dealAfterBatch(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{
		DoAt:     sim.CurrentTime + core.SpellBatchWindow,
		Priority: core.ActionPriorityLow,
		OnAction: func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		},
	}))
}

// The judgement debuffs Light, Wisdom and the Crusader put up: the paladin's melee strikes
// refresh them, and Judgement has to find the aura for the target it judges.
func (paladin *Paladin) newJudgementAuras(makeAura func(target *core.Unit) *core.Aura) core.AuraArray {
	auras := paladin.NewEnemyAuraArray(makeAura)
	paladin.JudgementAuras = append(paladin.JudgementAuras, auras)
	return auras
}

// Refreshes every judgement debuff this paladin has on the target, which the client does on every
// melee strike that lands.
func (paladin *Paladin) refreshJudgements(sim *core.Simulation, target *core.Unit) {
	for _, auras := range paladin.JudgementAuras {
		if aura := auras.Get(target); aura.IsActive() {
			aura.Refresh(sim)
		}
	}
}

var sealDuration = time.Second * 30

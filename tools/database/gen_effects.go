package database

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/template"

	_ "github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/tooltip"
)

// Sets the minimum itemlevel that should be considered for this expansions
const MIN_EFFECT_ILVL = 50

// Enchantment IDs at or below this are not generated.
const MIN_ENCHANT_EFFECT_ID = 0

func isGeneratableEnchant(effectID int32) bool {
	return effectID > MIN_ENCHANT_EFFECT_ID
}

type ProcInfo struct {
	Outcome             core.HitOutcome
	Callback            core.AuraCallback
	ProcMask            core.ProcMask
	MaxCumulativeStacks int32
	RequireDamageDealt  bool
	ClassSpellsOnly     bool
	// The listener's Can Proc From Procs attribute: it also hears hits from proc spells.
	CanProcFromProcs bool
	// A "Chance on hit" item effect or a combat enchant, cast by the game off every eligible
	// weapon hit regardless of proc flags. Decided by the trigger type, not by the item slot.
	IsWeaponProc bool
	// An aura proc carrying Aura Is Weapon Proc, which skips hits that suppress weapon procs.
	HonoursWeaponProcSuppression bool
}

// A weapon proc ignores proc-ness and already skips hits that suppress weapon procs, so the two
// aura-side fields would be dead or redundant next to it. Cleared so the generated file states
// one rule per listener.
func (info *ProcInfo) setIsWeaponProc(isWeaponProc bool) {
	info.IsWeaponProc = isWeaponProc
	if isWeaponProc {
		info.CanProcFromProcs = false
		info.HonoursWeaponProcSuppression = false
	}
}

// Entry represents a effect with its Item ID, Spell ID and display name.
type Variant struct {
	ID      int
	SpellID int
	Name    string
}

type Entry struct {
	Variants  []*Variant
	Tooltip   []string
	ProcInfo  ProcInfo
	Supported bool
	// What adds a stack while the window is open, for the trinkets whose effect carries a
	// separate accumulating aura. Nil for everything else.
	StackProcInfo *ProcInfo
	// Set for an on-use whose window accumulates a separate aura, which needs the stacking
	// helper rather than the flat one. Carries what that helper cannot read from the database.
	StackingOnUse *StackingOnUse
	// Set for effects an ignore list deliberately excludes. These emit a comment only, so that
	// skipping them is visible in the generated file rather than silent.
	Skipped bool
	// Set when the proc is resolved from the client's own rows at run time, which is every proc
	// but the two shapes that need more than the rows state: a window that accumulates a second
	// aura, and an effect a hand-written constructor already covers.
	Proc *ProcRouting
	// Set when the effect deals flat damage instead of granting stats. Those resolve no stats, so
	// without this they are dropped before they are ever emitted.
	DealsDamage bool
	// The same for an effect that heals the wearer.
	Heals bool
	// The same for an effect that shields the wearer with an absorb.
	Absorbs bool
	// What a registered on-use leaves out, stated beside its call.
	NotSimulated string
}

// The literals a stacking on-use needs in the generated call. Everything else - stacks,
// per-stack stats, which aura is which - the helper reads from the database at runtime.
type StackingOnUse struct {
	Name       string
	CooldownMs int32
}

// The two spells an item or enchant proc is resolved from at run time, and what the rows say the
// sim cannot model. The reasons are the ones sim/core/spelldata answers, so the registration the
// generator writes and the listener the sim builds come from the same reading.
type ProcRouting struct {
	TriggerSpellID int
	// Zero where the trigger's own row is the buff.
	BuffSpellID int
	// A "Chance on hit" item effect or a combat enchant, cast by the game off every eligible weapon
	// hit whatever the row's proc flags say.
	IsWeaponProc bool
	// A combat enchant's chance, stated on the enchantment's row rather than the spell's.
	ProcChancePct int
	// The trigger's outcome where the enchant's grant states it rather than the aura's own row.
	ProcHint core.ProcHint
	// Set where the spell the proc applies deals damage instead of granting an aura, which is a
	// constructor of its own: there is no buff to build.
	Damage bool
	// The same for a spell that heals the wearer.
	Heal bool
	// The same for a spell that shields the wearer with an absorb.
	Absorb bool
	// Empty when the rows state enough to build the listener.
	Unsupported []string
	// What the rows resolve to, for the reader of the generated file.
	Summary string
}

func (r *ProcRouting) Supported() bool {
	return len(r.Unsupported) == 0
}

// Renders the reasons as the generated file states them, one clause per shape.
func (r *ProcRouting) Reason() string {
	return strings.Join(r.Unsupported, "; ")
}

// The rows behind an item effect, by the two ids the sim will look up: the spell carrying the proc
// and the spell it applies.
func routeProc(triggerSpellID int, buffSpellID int, isWeaponProc bool) *ProcRouting {
	routing := &ProcRouting{TriggerSpellID: triggerSpellID, IsWeaponProc: isWeaponProc}
	if buffSpellID != triggerSpellID {
		routing.BuffSpellID = buffSpellID
	}

	trigger := spelldata.Find(int32(triggerSpellID))
	routing.Unsupported = spelldata.ItemProcUnsupported(trigger, isWeaponProc)
	routing.Summary = procSummary(triggerSpellID, trigger, buffSpellID)

	return routing
}

// A proc whose spell deals damage rather than granting an aura. The spell is named separately
// because the client hangs it below the trigger rather than on it.
func (r *ProcRouting) asDamage(damageSpellID int32) {
	r.Damage = true
	r.BuffSpellID = int(damageSpellID)

	if damage := spelldata.Find(damageSpellID); damage == spelldata.Nil {
		r.Unsupported = append(r.Unsupported, "the damage spell has no row in the store")
	} else if damage.DamageEffect() == spelldata.NilEffect {
		r.Unsupported = append(r.Unsupported, "the damage spell's row states no damage")
	}

	r.Summary = procSummary(r.TriggerSpellID, spelldata.Find(int32(r.TriggerSpellID)), r.BuffSpellID)
}

// A proc whose spell heals the wearer, the E_HEAL_PCT, E_HEAL or A_PERIODIC_HEAL spell the client
// hangs below the trigger.
func (r *ProcRouting) asHeal(healSpellID int32) {
	r.Heal = true
	r.BuffSpellID = int(healSpellID)
	r.Unsupported = append(r.Unsupported, healUnsupported(spelldata.Find(healSpellID))...)
	r.Summary = procSummary(r.TriggerSpellID, spelldata.Find(int32(r.TriggerSpellID)), r.BuffSpellID)
}

func healUnsupported(heal *spelldata.Spell) []string {
	var unsupported []string
	effect := heal.ProcHealEffect()
	if effect.Target[0] != dbcenums.TARGET_UNIT_CASTER {
		unsupported = append(unsupported, fmt.Sprintf("the heal lands on implicit target %d, not the wearer", effect.Target[0]))
	}
	if effect.Aura == dbcenums.A_PERIODIC_HEAL && (effect.PeriodMs <= 0 || heal.DurationMs <= 0) {
		unsupported = append(unsupported, "the heal over time states no period or no duration to tick over")
	}
	return unsupported
}

// A proc whose spell shields the wearer with the A_SCHOOL_ABSORB aura the client hangs below the
// trigger.
func (r *ProcRouting) asAbsorb(absorbSpellID int32) {
	r.Absorb = true
	r.BuffSpellID = int(absorbSpellID)
	r.Unsupported = append(r.Unsupported, absorbUnsupported(spelldata.Find(absorbSpellID))...)
	r.Summary = procSummary(r.TriggerSpellID, spelldata.Find(int32(r.TriggerSpellID)), r.BuffSpellID)
}

// An absorb of at least this much beside an A_DUMMY effect is a script's: the dummy stands for the
// spells it absorbs, which the client does not list. Blood of the Broodmother 1287808 states
// 10000000000 beside the dummy that names Dragon Breath spells.
const scriptedAbsorbAmount = 1e9

func absorbUnsupported(absorb *spelldata.Spell) []string {
	var unsupported []string
	effect := absorb.AbsorbEffect()
	if effect.Target[0] != dbcenums.TARGET_UNIT_CASTER {
		unsupported = append(unsupported, fmt.Sprintf("the absorb lands on implicit target %d, not the wearer", effect.Target[0]))
	}
	if effect.BasePoints >= scriptedAbsorbAmount && slices.ContainsFunc(absorb.Effects, func(e spelldata.Effect) bool {
		return e.Type == dbcenums.E_APPLY_AURA && e.Aura == dbcenums.A_DUMMY
	}) {
		unsupported = append(unsupported, fmt.Sprintf(
			"the absorb of %.0f beside an A_DUMMY absorbs only the spells a script names, which the client does not list", effect.BasePoints))
	}
	if absorb.DurationMs == 0 {
		unsupported = append(unsupported, "the absorb states no duration")
	}
	for _, id := range absorbLinkedDamage(absorb) {
		unsupported = append(unsupported, fmt.Sprintf("the damage of %d (%s) the absorb's row names is not simulated", id, spelldata.Find(id).Name))
	}
	return unsupported
}

// The spells dealing damage that an absorb's row triggers or its tooltip references, which the shield
// alone does not deal: Adaptive Combat Assistant's 1291097 names 1291099, Nature damage when the
// shield breaks early. Registering the shield without it would model half the item.
func absorbLinkedDamage(absorb *spelldata.Spell) []int32 {
	linked := slices.Clone(absorb.RefIDs)
	for _, e := range absorb.Effects {
		if e.TriggerID != 0 {
			linked = append(linked, e.TriggerID)
		}
	}

	var damage []int32
	for _, id := range linked {
		s := spelldata.Find(id)
		if id != absorb.ID && s != spelldata.Nil && !slices.Contains(damage, id) &&
			(s.DamageEffect() != spelldata.NilEffect || s.PeriodicDamageEffect() != spelldata.NilEffect) {
			damage = append(damage, id)
		}
	}
	return damage
}

// The heal spell a proc casts: the spell itself, or one it triggers.
func procHealSpell(spellID int32) int32 {
	return findProcSpell(spellID, func(s *spelldata.Spell) bool { return s.ProcHealEffect() != spelldata.NilEffect }, map[int32]bool{})
}

// The same for the spell applying an absorb.
func procAbsorbSpell(spellID int32) int32 {
	return findProcSpell(spellID, func(s *spelldata.Spell) bool { return s.AbsorbEffect() != spelldata.NilEffect }, map[int32]bool{})
}

func findProcSpell(spellID int32, casts func(*spelldata.Spell) bool, seen map[int32]bool) int32 {
	s := spelldata.Find(spellID)
	if s == spelldata.Nil || seen[spellID] {
		return 0
	}
	seen[spellID] = true

	if casts(s) {
		return spellID
	}
	for i := range s.Effects {
		if s.Effects[i].TriggerID == 0 {
			continue
		}
		if found := findProcSpell(s.Effects[i].TriggerID, casts, seen); found != 0 {
			return found
		}
	}
	return 0
}

// The buff the proc applies has to last for something: an aura of no duration is one the sim
// refuses to activate, and the client states it on either row.
func (r *ProcRouting) requireABuffDuration() {
	trigger := spelldata.Find(int32(r.TriggerSpellID))
	buff := trigger
	if r.BuffSpellID != 0 {
		buff = spelldata.Find(int32(r.BuffSpellID))
	}

	if buff == spelldata.Nil {
		r.Unsupported = append(r.Unsupported, "the buff has no row in the store")
		return
	}

	if buff.DurationMs == 0 && trigger.DurationMs == 0 {
		r.Unsupported = append(r.Unsupported, "neither row states how long the buff lasts")
	}
}

// What the rows resolve to, as the sim's own constants, so the generated file states the reading
// rather than leaving it to be looked up.
func procSummary(triggerSpellID int, trigger *spelldata.Spell, buffSpellID int) string {
	if trigger == spelldata.Nil {
		return fmt.Sprintf("trigger %d is not in the store", triggerSpellID)
	}

	decoded := core.DecodeProcTypeMask(trigger.ProcFlags, trigger.ProcHint)
	summary := fmt.Sprintf("trigger %d (%s, %s, %s)", trigger.ID,
		procRateSummary(trigger), asCoreCallback(decoded.Callback), asCoreProcMask(decoded.ProcMask))

	if int(trigger.ID) != buffSpellID {
		summary += fmt.Sprintf(" -> buff %d", buffSpellID)
	}

	return summary
}

func procRateSummary(trigger *spelldata.Spell) string {
	switch {
	case trigger.RPPM > 0:
		return fmt.Sprintf("%v ppm", trigger.RPPM)
	case trigger.ItemProcRollsTheColumn():
		return fmt.Sprintf("%d%%, the column over effect %d's %v%%", trigger.ProcChance, trigger.ProcChanceEffect, trigger.StatedChance()*100)
	case trigger.ProcChanceSource == spelldata.ProcChanceEffectN:
		return fmt.Sprintf("effect %d's chance", trigger.ProcChanceEffect)
	case trigger.ProcChanceSource == spelldata.ProcChanceAlways:
		return "every time"
	case trigger.ProcChanceSource == spelldata.ProcChancePPM:
		return "no stated rate"
	default:
		return fmt.Sprintf("%d%%", trigger.ProcChance)
	}
}

// Group holds a category of effects.
type Group struct {
	Name    string
	Entries []*Entry
}

type MissingItemEffect struct {
	ItemID  int32
	Name    string
	Effects []Variant
}

var missingEffectsMap = map[string]map[int32]MissingItemEffect{
	"EnchantEffects": {},
	"ItemEffects":    {},
}

type EffectParseResult byte

const (
	EffectParseResultInvalid     EffectParseResult = iota // Returned when the effect is invalid for the current parameters
	EffectParseResultUnsupported                          // Returned when the effect could be parsed but is not supported for effect generation
	EffectParseResultSuccess                              // Returned when the effect was parsed successfuly
	EffectParseResultRefused                              // Returned when the effect was parsed, is not supported, and said so in an entry of its own
)

func GenerateEffectsFile(groups []*Group, outFile string, templateString string) error {
	if _, err := os.Stat(outFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to check file %s: %w", outFile, err)
	}

	// Ensure groups and entries are sorted
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})

	for _, grp := range groups {
		sort.Slice(grp.Entries, func(i, j int) bool {
			if grp.Entries[i].Supported != grp.Entries[j].Supported {
				return !grp.Entries[i].Supported
			}

			return entryOrder(grp.Entries[i], grp.Entries[j])
		})
	}

	funcMap := map[string]any{
		"asCoreCallback": asCoreCallback,
		"asCoreProcMask": asCoreProcMask,
		"asCoreOutcome":  asCoreOutcome,
		"formatStrings":  formatStrings,
		"formatProcHint": formatProcHint,
	}
	tmpl := template.Must(template.New("effects").Funcs(funcMap).Parse(templateString))

	// An empty generated file must not import anything: gen_db links sim/common, so unused
	// imports in a file it just wrote break the very build the next run needs.
	hasEntries := false
	for _, grp := range groups {
		if len(grp.Entries) > 0 {
			hasEntries = true
			break
		}
	}

	hasStacking := false
	// A registration resolved from the client's rows names no core constant unless it carries a hint,
	// so a file whose live entries are all of that shape must not import core: gen_db links the sim,
	// and an unused import in a file it just wrote breaks the build the next run needs.
	usesCore := false
	// And nothing at all is imported by a file whose every entry is commented out.
	hasLive := false
	for _, grp := range groups {
		for _, entry := range grp.Entries {
			if entry.StackingOnUse != nil {
				hasStacking = true
			}
			if entry.Skipped || !entry.Supported {
				continue
			}
			hasLive = true
			if entry.Proc == nil || entry.Proc.ProcHint != 0 {
				usesCore = true
			}
		}
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, map[string]interface{}{"Groups": groups, "HasEntries": hasEntries, "HasStacking": hasStacking, "UsesCore": usesCore, "HasLive": hasLive}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// The template cannot indent commented-out blocks or blank lines the way gofmt wants, so
	// format the result. Otherwise every regeneration reverts whatever formatted the file last
	// and the diff is hundreds of whitespace-only lines.
	out := rendered.Bytes()
	if formatted, err := format.Source(out); err != nil {
		fmt.Printf("WARN: generated %s is not valid Go, writing unformatted: %v\n", outFile, err)
	} else {
		out = formatted
	}

	if err := os.WriteFile(outFile, out, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", outFile, err)
	}

	return nil
}

// Whether two entries are resolved from the same rows, which is what lets them be emitted as one
// call with a variant list.
func sameProcRows(a *Entry, b *Entry) bool {
	if a.Proc == nil || b.Proc == nil {
		return a.Proc == b.Proc
	}

	return a.Proc.TriggerSpellID == b.Proc.TriggerSpellID && a.Proc.BuffSpellID == b.Proc.BuffSpellID
}

// A total order over entries. Sorting on the item or enchant ID alone is not one: an item with
// two effects yields two entries sharing that ID, and sort.Slice is not stable, so their order
// in the generated file flipped between runs.
func entryOrder(a *Entry, b *Entry) bool {
	if a.Variants[0].ID != b.Variants[0].ID {
		return a.Variants[0].ID < b.Variants[0].ID
	}
	if a.Variants[0].SpellID != b.Variants[0].SpellID || a.Proc == nil || b.Proc == nil {
		return a.Variants[0].SpellID < b.Variants[0].SpellID
	}
	return a.Proc.TriggerSpellID < b.Proc.TriggerSpellID
}

// Escapes a rendered tooltip for use inside a double-quoted TypeScript string. Tooltips
// routinely span several lines, which would otherwise produce a file that does not parse.
func jsString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	return s
}

const missingEffectsFileName = "ui/sim/constants/missing_effects_auto_gen.ts"

func GenerateMissingEffectsFile() error {
	if _, err := os.Stat(missingEffectsFileName); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to check file %s: %w", missingEffectsFileName, err)
	}

	funcMap := map[string]any{
		"asCoreCallback": asCoreCallback,
		"asCoreProcMask": asCoreProcMask,
		"asCoreOutcome":  asCoreOutcome,
		"formatStrings":  formatStrings,
		"jsString":       jsString,
	}
	tmpl := template.Must(template.New("missingEffects").Funcs(funcMap).Parse(TmplStrMissingEffects))
	f, err := os.Create(missingEffectsFileName)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", missingEffectsFileName, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, missingEffectsMap); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func GenerateEnchantEffects(instance *dbc.DBC, db *WowDatabase) {
	groupMapProc := map[string]Group{}
	enchantSpellEffects := enchantGrantEffects(instance.SpellEffectsById)

	// The raw rows repeat an enchant once per recipe name, and an enchant registers once.
	generated := map[int32]bool{}
	for _, enchant := range instance.Enchants {
		parsed := enchant.ToProto()
		if _, ok := db.Enchants[EnchantToDBKey(parsed)]; !ok || generated[parsed.EffectId] {
			continue
		}
		generated[parsed.EffectId] = true

		TryParseEnchantEffect(parsed, enchant.ProcSlots(), groupMapProc, instance, enchantSpellEffects)
		storeUnmappedSpeedEnchant(instance, enchant, parsed, enchantSpellEffects)
	}

	var procGroups []*Group
	for _, grp := range groupMapProc {
		procGroups = append(procGroups, &grp)
	}

	GenerateEffectsFile(procGroups, "sim/common/forever/enchants_auto_gen.go", TmplStrEnchant)
}

var speedAuras = []dbcenums.EffectAuraType{
	dbcenums.A_MOD_ATTACKSPEED, dbcenums.A_MOD_MELEE_HASTE, dbcenums.A_MOD_MELEE_RANGED_HASTE,
	dbcenums.A_HASTE_SPELLS, dbcenums.A_MOD_MELEE_HASTE_2, dbcenums.A_MOD_RANGED_HASTE_2,
}

// These attack and cast speed auras reach no enchant field, so an enchant whose equip spell states
// one is listed as missing its effect unless it is implemented by hand.
func storeUnmappedSpeedEnchant(instance *dbc.DBC, enchant dbc.Enchant, parsed *proto.UIEnchant, enchantSpellEffects map[int]*dbc.SpellEffect) {
	grant, ok := enchantSpellEffects[int(parsed.EffectId)]
	if !ok || core.HasEnchantEffect(parsed.EffectId) {
		return
	}

	for idx, effect := range enchant.Effects {
		if effect != dbc.ITEM_ENCHANTMENT_EQUIP_SPELL || idx >= len(enchant.EffectArgs) {
			continue
		}
		spellID := enchant.EffectArgs[idx]
		for _, spellEffect := range instance.SpellEffectsInOrder(spellID) {
			if spellEffect.EffectType == dbcenums.E_APPLY_AURA && slices.Contains(speedAuras, spellEffect.EffectAura) {
				StoreMissingEffect("EnchantEffects", parsed.Name, Variant{
					ID:      int(parsed.EffectId),
					Name:    renderSpellTooltip(instance, grant.SpellID) + " (an attack or cast speed aura, which no enchant stat carries)",
					SpellID: spellID,
				})
				return
			}
		}
	}
}

// The E_ENCHANT_ITEM effect that grants each enchant, by enchant ID. Only that effect names an
// enchant in its first misc value; any other effect's misc value is a school, a stat or a power.
//
// Several spells can grant the same enchant -- an enchant that was re-taught
// by a later expansion's recipe has one spell per version, up to five here.
// Map iteration order is randomized, so keep the lowest spell ID rather than
// letting whichever one is visited last win and churn the generated file.
func enchantGrantEffects(effects map[int]dbc.SpellEffect) map[int]*dbc.SpellEffect {
	grants := map[int]*dbc.SpellEffect{}
	for _, effect := range effects {
		if effect.EffectType != dbcenums.E_ENCHANT_ITEM {
			continue
		}
		enchantID := effect.EffectMiscValues[0]
		if existing, ok := grants[enchantID]; ok && existing.SpellID <= effect.SpellID {
			continue
		}
		grants[enchantID] = &effect
	}
	return grants
}

// Names the ignore-list rule that excluded an effect, for the comment emitted in the generated
// file. Returns "" when nothing excludes it.
func ignoredEffectReason(instance *dbc.DBC, effectID int) string {
	for _, effect := range instance.SpellEffectsInOrder(effectID) {
		if params, ok := IgnoreSpellEffectByAuraType[effect.EffectAura]; ok {
			if len(params) == 0 || slices.Contains(params, effect.EffectMiscValues[0]) {
				return fmt.Sprintf("ignored aura type %d", effect.EffectAura)
			}
		}

		if params, ok := IgnoreSpellEffectBySpellEffectType[effect.EffectType]; ok {
			if len(params) == 0 || slices.Contains(params, effect.EffectMiscValues[0]) {
				return fmt.Sprintf("ignored effect type %d", effect.EffectType)
			}
		}
	}

	return ""
}

// Records an effect excluded by an ignore list so the generated file documents it. Kept in its
// own group: variant merging is per-group, so these cannot affect whether a real effect's
// variant set is emitted live or commented.
func storeSkippedEffect(id int32, name string, buffID int32, instance *dbc.DBC, groupMap map[string]Group) {
	grp, exists := groupMap["Skipped"]
	if !exists {
		grp = Group{Name: "Skipped"}
	}

	buffName := instance.Spells[int(buffID)].NameLang
	grp.Entries = append(grp.Entries, &Entry{
		Skipped:  true,
		Variants: []*Variant{{ID: int(id), Name: name, SpellID: int(buffID)}},
		Tooltip: []string{fmt.Sprintf("%s: %q (%d) - %s",
			name, buffName, buffID, ignoredEffectReason(instance, int(buffID)))},
	})
	groupMap["Skipped"] = grp
}

func ItemEffectIsSupported(instance *dbc.DBC, effectID int) bool {
	supported := true
	if effects, ok := instance.SpellEffects[effectID]; ok {
		for _, effect := range effects {
			if params, ok := IgnoreSpellEffectByAuraType[effect.EffectAura]; ok {
				if len(params) == 0 {
					supported = false
					break
				} else {
					if slices.Contains(params, effect.EffectMiscValues[0]) {
						supported = false
					}
				}
			}

			if params, ok := IgnoreSpellEffectBySpellEffectType[effect.EffectType]; ok {
				if len(params) == 0 {
					supported = false
					break
				} else {
					if slices.Contains(params, effect.EffectMiscValues[0]) {
						supported = false
					}
				}
			}
		}
	}
	return supported
}

func GenerateItemEffects(instance *dbc.DBC, db *WowDatabase, itemSources map[int][]*proto.DropSource) {
	groupMapOnUse := map[string]Group{}
	groupMapProc := map[string]Group{}

	// Example loop over your items
	for _, parsed := range db.Items {
		parsed.ItemEffects = dbc.MergeItemEffectsForAllStates(parsed)

		for _, itemEffect := range parsed.ItemEffects {
			if !ItemEffectIsSupported(instance, int(itemEffect.BuffId)) {
				// Commented into the generated file rather than dropped. These are deliberately
				// out of scope - summons, teleports, created items - but an item whose only
				// effect is skipped otherwise vanished with no trace, while a sibling marker
				// aura on the same item got reported as missing instead.
				skippedGroup := groupMapProc
				if itemEffect.GetOnUse() != nil {
					skippedGroup = groupMapOnUse
				}
				storeSkippedEffect(parsed.Id, parsed.Name, itemEffect.BuffId, instance, skippedGroup)
				continue
			}

			switch TryParseOnUseEffect(parsed, itemEffect, instance, groupMapOnUse) {
			case EffectParseResultSuccess, EffectParseResultRefused:
				continue
			}

			switch TryParseProcEffect(parsed, itemEffect, instance, groupMapProc) {
			case EffectParseResultSuccess, EffectParseResultRefused:
			default:
				ParseTooltipForMissingEffect(parsed, itemEffect, instance, groupMapProc, "Procs")
			}
		}
	}

	// Sorting done in GenerateEffectsFile
	var onUseGroups []*Group
	for _, grp := range groupMapOnUse {
		onUseGroups = append(onUseGroups, &grp)
	}

	// Merge variants
	var procGroups []*Group
	needsStatPostfix := map[string]bool{}
	for _, grp := range groupMapProc {
		newEntries := []*Entry{}
		entryGroupings := map[string]*Entry{}

		// sort entries first to make tooltip generation consistent for variants
		sort.Slice(grp.Entries, func(i, j int) bool {
			return entryOrder(grp.Entries[i], grp.Entries[j])
		})

		for _, entry := range grp.Entries {
			var idx int64 = 0
			added := false

			// Make sure to only group by name and proc mask, each proc mask will create it's own sub group
			// A variant set is emitted as one call, so its members also have to name the same rows:
			// the reissued PvP shields share a name and a buff and carry different triggers.
			for _, group := range entryGroupings {
				if group.Variants[0].Name == entry.Variants[0].Name {
					idx++
					if group.ProcInfo.ProcMask == entry.ProcInfo.ProcMask && sameProcRows(group, entry) {
						group.AddVariant(entry.Variants[0])
						added = true
						break
					}
				}
			}

			if !added {
				groupName := entry.Variants[0].Name
				if idx > 0 {
					needsStatPostfix[groupName] = true
					groupName += "(" + strconv.FormatInt(idx, 10) + ")"
				}

				newEntries = append(newEntries, entry)
				entryGroupings[entry.Variants[0].Name] = entry
			}
		}

		grp.Entries = newEntries
		procGroups = append(procGroups, &grp)
	}

	updateNames := func(entries []*Entry) {
		for _, entry := range entries {
			for _, variant := range entry.Variants {
				if _, ok := needsStatPostfix[variant.Name]; ok {
					item := db.Items[int32(variant.ID)]
					for _, itemEffect := range item.ItemEffects {
						variant.Name += " - " + GetEffectStatString(itemEffect)
					}
				}

				variant.Name += BuildItemDifficultyPostfix(itemSources, variant.ID, instance)
			}
		}
	}

	// Update Item names
	for _, grp := range onUseGroups {
		updateNames(grp.Entries)
	}

	for _, grp := range procGroups {
		updateNames(grp.Entries)
	}

	GenerateEffectsFile(onUseGroups, "sim/common/forever/stat_bonus_cds_auto_gen.go", TmplStrOnUse)
	GenerateEffectsFile(procGroups, "sim/common/forever/stat_bonus_procs_auto_gen.go", TmplStrProc)
}

func GenerateItemEffectRandomPropPoints(instance *dbc.DBC, db *WowDatabase) {
	for id, allocMap := range instance.RandomPropertiesByIlvl {
		ilvl := int32(id)
		if ilvl < core.MinIlvl || ilvl > core.MaxIlvl {
			continue
		}
		db.ItemEffectRandPropPoints[ilvl] = &proto.ItemEffectRandPropPoints{
			Ilvl:           ilvl,
			RandPropPoints: allocMap[proto.ItemQuality_ItemQualityEpic][0],
		}
	}
}

func BuildItemDifficultyPostfix(itemSources map[int][]*proto.DropSource, itemId int, instance *dbc.DBC) string {
	difficultyPostfix := ""
	if sources, ok := itemSources[itemId]; ok {
		name := DifficultyToShortName(sources[0].Difficulty)
		if len(name) > 0 {
			difficultyPostfix += " " + name
		}
	}

	if item, ok := instance.Items[itemId]; ok {
		if len(item.NameDescription) > 0 && item.NameDescription != "Heroic" {
			difficultyPostfix += " (" + item.NameDescription + ")"
		}

		if item.Flags1.Has(dbc.HORDE_SPECIFIC) {
			difficultyPostfix += " (Horde)"
		}

		if item.Flags1.Has(dbc.ALLIANCE_SPECIFIC) {
			difficultyPostfix += " (Alliance)"
		}
	}

	return difficultyPostfix
}

func TryParseProcEffect(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC, groupMapProc map[string]Group) EffectParseResult {
	if itemEffect.GetProc() != nil && parsed.ScalingOptions[0].Ilvl >= MIN_EFFECT_ILVL {
		// Effect was already manually implemented
		if core.HasItemEffect(parsed.Id) {
			return EffectParseResultSuccess
		}

		tooltipString, id := dbc.GetItemEffectSpellTooltip(int(parsed.Id), int(itemEffect.BuffId))
		tooltip, _ := tooltip.ParseTooltip(tooltipString, tooltip.DBCTooltipDataProvider{DBC: instance}, int64(id))

		grp, exists := groupMapProc["Procs"]
		if !exists {
			grp = Group{Name: "Procs"}
		}

		if tooltip != nil {
			renderedTooltip := tooltip.String()
			entry := Entry{Tooltip: strings.Split(renderedTooltip, "\n"), Variants: []*Variant{{ID: int(parsed.Id), Name: parsed.Name, SpellID: int(itemEffect.BuffId)}}}
			entry.ProcInfo, entry.Supported = BuildProcInfo(parsed, int(itemEffect.BuffId), instance, renderedTooltip)

			entry.StackProcInfo = buildStackProcInfo(itemEffect, instance, renderedTooltip)

			// entry.Supported speaks only for the trigger that opens the window, so the stack side
			// has to be refused separately. Without this the template omits StackCallback,
			// attachStackTrigger early-returns on the empty callback and the stat aura activates at
			// zero stacks with no duration of its own - a trinket whose window opens and never gains
			// a stack, worth nothing and reported as implemented. The on-use path refuses the same
			// shape further down.
			if itemEffect.StackingAura != nil && itemEffect.GetStackProc() != nil && entry.StackProcInfo == nil {
				entry.Supported = false
			}

			// A stat-buff proc carries its two spell ids and nothing else: what it hears, how often
			// and for how long are the rows' to say, at run time, through the same decision this
			// reads here - and that decision is the whole of it, rather than the tooltip reading
			// BuildProcInfo does for the shapes below. The two that need more than the rows state
			// stay where they are: a window accumulating a second aura, and an effect with no stats.
			if itemEffect.StackingAura == nil && len(dbc.EffectStats(itemEffect)) > 0 {
				entry.Proc = routeItemProc(parsed, itemEffect, renderedTooltip)
				if entry.Proc != nil {
					entry.Proc.requireABuffDuration()
					entry.Supported = entry.Proc.Supported()
				}
			}

			// An effect that resolves no stats may still deal flat damage, which is a shape of its
			// own rather than a reason to refuse: there is no buff, so the proc casts the spell the
			// client hangs below its trigger, read from that spell's own row.
			if len(dbc.EffectStats(itemEffect)) == 0 {
				if damage := dbc.ResolveDamageEffect(int(itemEffect.BuffId)); damage != nil {
					entry.Proc = routeItemProc(parsed, itemEffect, renderedTooltip)
					if entry.Proc != nil {
						entry.Proc.asDamage(int32(damage.SpellID))
						entry.Supported = entry.Proc.Supported()
						entry.DealsDamage = true
					}
				}
			}

			// The same for an effect that heals the wearer.
			if len(dbc.EffectStats(itemEffect)) == 0 && !entry.DealsDamage {
				if heal := procHealSpell(itemEffect.BuffId); heal != 0 {
					entry.Proc = routeItemProc(parsed, itemEffect, renderedTooltip)
					if entry.Proc != nil {
						entry.Proc.asHeal(heal)
						entry.Supported = entry.Proc.Supported()
						entry.Heals = true
					}
				}
			}

			// The same for an effect that shields the wearer.
			if len(dbc.EffectStats(itemEffect)) == 0 && !entry.DealsDamage && !entry.Heals {
				if absorb := procAbsorbSpell(itemEffect.BuffId); absorb != 0 {
					entry.Proc = routeItemProc(parsed, itemEffect, renderedTooltip)
					if entry.Proc != nil {
						entry.Proc.asAbsorb(absorb)
						entry.Supported = entry.Proc.Supported()
						entry.Absorbs = true
					}
				}
			}

			if (len(dbc.EffectStats(itemEffect)) == 0 && !entry.DealsDamage && !entry.Heals && !entry.Absorbs) || !entry.Supported {
				StoreMissingEffect("ItemEffects", parsed.Name, Variant{
					ID:      int(parsed.Id),
					Name:    renderedTooltip,
					SpellID: int(itemEffect.BuffId),
				})

				// A proc the rows themselves refuse is emitted with the reason they gave, so the
				// generated file says why rather than leaving a shapeless commented block behind.
				if entry.Proc != nil && !entry.Proc.Supported() {
					grp.Entries = append(grp.Entries, &entry)
					groupMapProc["Procs"] = grp
					return EffectParseResultRefused
				}

				return EffectParseResultUnsupported
			}

			grp.Entries = append(grp.Entries, &entry)
			groupMapProc["Procs"] = grp

			return EffectParseResultSuccess
		} else {
			return EffectParseResultUnsupported
		}
	}

	// check if the item has any kind of proc as we only support stat proc parsing right now
	if effects, ok := instance.ItemEffectsByParentID[int(parsed.Id)]; ok && parsed.ScalingOptions[0].Ilvl >= MIN_EFFECT_ILVL {
		for _, effect := range effects {
			if SpellHasTriggerEffect(effect.SpellID, instance) {
				return EffectParseResultUnsupported
			}
		}
	}

	return EffectParseResultInvalid
}

// The rows an item effect names: the client's ItemEffect row carries the spell with the proc on it,
// and the shipped entry carries the spell that applies the stats. A 100 beside "sometimes" on the
// tooltip is a rate the rows do not carry, as it is on an enchant: Darkmoon Card: Heroism 23689.
func routeItemProc(parsed *proto.UIItem, itemEffect *proto.ItemEffect, tooltip string) *ProcRouting {
	effect := dbc.GetItemEffectForBuffID(int(parsed.Id), int(itemEffect.BuffId))
	if effect == nil {
		return nil
	}

	routing := routeProc(effect.SpellID, int(itemEffect.BuffId), effect.TriggerType == dbc.ITEM_SPELLTRIGGER_CHANCE_ON_HIT)
	trigger := spelldata.Find(int32(effect.SpellID))
	if trigger.ProcChanceSource == spelldata.ProcChanceAlways && trigger.RPPM == 0 &&
		enchantTooltipStatesAnUnknownRate(tooltip) && !statesNoRate(routing) {
		routing.Unsupported = append(routing.Unsupported, spelldata.ReasonStatesNoRate)
	}

	return routing
}

func TryParseOnUseEffect(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC, groupMap map[string]Group) EffectParseResult {
	// Effect was already manually implemented
	if core.HasItemEffect(parsed.Id) {
		return EffectParseResultSuccess
	}

	if itemEffect.GetOnUse() != nil && parsed.ScalingOptions[0].Ilvl >= MIN_EFFECT_ILVL {
		if itemEffect.GetOnUse().CooldownMs < 0 && itemEffect.GetOnUse().CategoryCooldownMs < 0 {
			return EffectParseResultUnsupported
		}

		if len(dbc.EffectStats(itemEffect)) == 0 {
			return parseOnUseSpell(parsed, itemEffect, instance, groupMap)
		}

		groupName := GetEffectStatString(itemEffect)
		grp, exists := groupMap[groupName]
		if !exists {
			grp = Group{Name: groupName}
		}

		entry := &Entry{Variants: []*Variant{{ID: int(parsed.Id), Name: parsed.Name, SpellID: int(itemEffect.BuffId)}}, Supported: true}
		grp.Entries = append(grp.Entries, entry)
		groupMap[groupName] = grp

		// A stacking on-use keeps its stats on the accumulating aura, so the flat check below
		// would call it unsupported, and the flat helper would grant nothing.
		stacking := itemEffect.StackingAura
		if stacking != nil && len(stacking.GetScalingOptions()[0].GetStats()) > 0 {
			entry.StackProcInfo = buildStackProcInfo(itemEffect, instance, "")
			if entry.StackProcInfo == nil {
				entry.Supported = false
				return EffectParseResultUnsupported
			}
			entry.StackingOnUse = &StackingOnUse{
				Name:       parsed.Name,
				CooldownMs: itemEffect.GetOnUse().CooldownMs,
			}
			return EffectParseResultSuccess
		}

		if trigger := buffProcTrigger(itemEffect.BuffId); trigger != 0 {
			entry.NotSimulated = fmt.Sprintf("the proc the buff carries, %d (%s)", trigger, spellEffectKinds(instance, int(trigger)))
			StoreMissingEffect("ItemEffects", parsed.Name, Variant{
				ID:      int(parsed.Id),
				Name:    itemEffectTooltip(parsed, itemEffect, instance),
				SpellID: int(itemEffect.BuffId),
			})
		}

		return EffectParseResultSuccess
	}

	return EffectParseResultInvalid
}

// The spell an A_PROC_TRIGGER_SPELL on the buff casts, or 0. The flat on-use grants the buff's stats
// and nothing it procs: Aegis of Preservation 23780's heal on every hit taken, 23781.
func buffProcTrigger(buffID int32) int32 {
	for _, e := range spelldata.Find(buffID).Effects {
		if e.Type == dbcenums.E_APPLY_AURA && e.Aura == dbcenums.A_PROC_TRIGGER_SPELL {
			return e.TriggerID
		}
	}
	return 0
}

// An on-use with no stats to grant, resolved from the spell it casts. One the sim cannot build from
// that is refused in an entry of its own, with the reason, and listed as missing.
func parseOnUseSpell(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC, groupMap map[string]Group) EffectParseResult {
	routing := routeOnUse(parsed, itemEffect, instance)
	entry := &Entry{
		Variants:    []*Variant{{ID: int(parsed.Id), Name: parsed.Name, SpellID: int(itemEffect.BuffId)}},
		Proc:        routing,
		Supported:   routing.Supported(),
		DealsDamage: routing.Damage,
		Heals:       routing.Heal,
		Absorbs:     routing.Absorb,
	}

	groupName := ""
	switch {
	case !entry.Supported:
	case routing.Damage:
		groupName = "Damage"
	case routing.Heal:
		groupName = "Heals"
	case routing.Absorb:
		groupName = "Absorbs"
	}
	grp := groupMap[groupName]
	grp.Name = groupName
	grp.Entries = append(grp.Entries, entry)
	groupMap[groupName] = grp

	if entry.Supported {
		return EffectParseResultSuccess
	}

	if _, ignored := IgnoreMissingEffectBySpellID[int(itemEffect.BuffId)]; !ignored {
		StoreMissingEffect("ItemEffects", parsed.Name, Variant{
			ID:      int(parsed.Id),
			Name:    itemEffectTooltip(parsed, itemEffect, instance),
			SpellID: int(itemEffect.BuffId),
		})
	}
	return EffectParseResultRefused
}

// The spell an on-use casts, read from the store the way the sim reads it: damage on the enemy it is
// used on, or a heal or an absorb on the wearer. What else the row does - a root, a stun - is left
// out, and the summary names it.
func routeOnUse(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC) *ProcRouting {
	return routeOnUseSpell(int(parsed.Id), itemEffect.BuffId, instance)
}

func routeOnUseSpell(itemID int, buffID int32, instance *dbc.DBC) *ProcRouting {
	spellID := int(buffID)
	routing := &ProcRouting{TriggerSpellID: spellID}
	s := spelldata.Find(buffID)
	direct, periodic, heal, absorb := s.DamageEffect(), s.PeriodicDamageEffect(), s.ProcHealEffect(), s.AbsorbEffect()

	switch {
	case !castsOnUse(itemID, spellID, instance):
		routing.Unsupported = append(routing.Unsupported, fmt.Sprintf("the item's on-use row does not cast %d itself", spellID))
	case s == spelldata.Nil:
		routing.Unsupported = append(routing.Unsupported, "the spell has no row in the store")
	case direct != spelldata.NilEffect || periodic != spelldata.NilEffect:
		routing.Damage = true
		routing.Summary = onUseSummary(s, direct, periodic)
		if direct != spelldata.NilEffect && direct.Type != dbcenums.E_SCHOOL_DAMAGE {
			routing.Unsupported = append(routing.Unsupported, fmt.Sprintf("the direct damage is %v rather than an amount", direct.Type))
		}
		for _, e := range []*spelldata.Effect{direct, periodic} {
			if e != spelldata.NilEffect && !e.HitsAnEnemy() {
				routing.Unsupported = append(routing.Unsupported, fmt.Sprintf("the damage lands on implicit target %d, not an enemy", e.Target[0]))
			}
		}
		if periodic != spelldata.NilEffect && (periodic.PeriodMs <= 0 || s.DurationMs <= 0) {
			routing.Unsupported = append(routing.Unsupported, "the damage over time states no period or no duration to tick over")
		}
	case heal != spelldata.NilEffect:
		routing.Heal = true
		routing.Summary = onUseSummary(s, heal)
		routing.Unsupported = append(routing.Unsupported, healUnsupported(s)...)
	case absorb != spelldata.NilEffect:
		routing.Absorb = true
		routing.Summary = onUseSummary(s, absorb)
		routing.Unsupported = append(routing.Unsupported, absorbUnsupported(s)...)
	default:
		routing.Unsupported = append(routing.Unsupported,
			fmt.Sprintf("%d deals no damage and heals no one (%s)", spellID, spellEffectKinds(instance, spellID)))
	}

	return routing
}

// Whether the item's own on-use row names the spell, rather than a spell that reaches it through a
// trigger: the sim casts the spell the effect names.
func castsOnUse(itemID int, spellID int, instance *dbc.DBC) bool {
	return slices.ContainsFunc(instance.ItemEffectsByParentID[itemID], func(e dbc.ItemEffect) bool {
		return e.TriggerType == dbc.ITEM_SPELLTRIGGER_ON_USE && e.SpellID == spellID
	})
}

func onUseSummary(s *spelldata.Spell, modelled ...*spelldata.Effect) string {
	var cast, left []string
	for i := range s.Effects {
		e := &s.Effects[i]
		kind := e.Type.String()
		if e.Type == dbcenums.E_APPLY_AURA {
			kind = e.Aura.String()
		}
		if slices.Contains(modelled, e) {
			cast = append(cast, kind)
		} else {
			left = append(left, kind)
		}
	}

	summary := fmt.Sprintf("on use: %d (%s)", s.ID, strings.Join(cast, ", "))
	if len(left) > 0 {
		summary += fmt.Sprintf("; not simulated: %s", strings.Join(left, ", "))
	}
	return summary
}

// The item effect's tooltip as the missing-effects list shows it, or the spell's name where the
// tooltip does not render.
func itemEffectTooltip(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC) string {
	tooltipString, id := dbc.GetItemEffectSpellTooltip(int(parsed.Id), int(itemEffect.BuffId))
	if rendered, _ := tooltip.ParseTooltip(tooltipString, tooltip.DBCTooltipDataProvider{DBC: instance}, int64(id)); rendered != nil {
		return rendered.String()
	}
	return instance.Spells[int(itemEffect.BuffId)].NameLang
}

// The constructor an on-use routed from its spell registers through. A refused one that is neither
// damage nor a heal names the stat constructor in its commented call.
func (r *ProcRouting) OnUseConstructor() string {
	switch {
	case r.Damage:
		return "NewSpellDataDamageOnUse"
	case r.Heal:
		return "NewSpellDataHealOnUse"
	case r.Absorb:
		return "NewSpellDataAbsorbOnUse"
	default:
		return "NewSimpleStatActive"
	}
}

// The constructor an item or enchant proc registers through.
func (r *ProcRouting) ProcConstructor() string {
	switch {
	case r.Damage:
		return "NewSpellDataDamageProc"
	case r.Heal:
		return "NewSpellDataHealProc"
	case r.Absorb:
		return "NewSpellDataAbsorbProc"
	default:
		return "NewSpellDataProc"
	}
}

func TryParseEnchantEffect(enchant *proto.UIEnchant, slots []dbc.EnchantProcSlot, groupMapProc map[string]Group, instance *dbc.DBC, enchantSpellEffects map[int]*dbc.SpellEffect) EffectParseResult {
	if len(slots) == 0 || !isGeneratableEnchant(enchant.EffectId) {
		return EffectParseResultInvalid
	}

	// Effect was already manually implemented
	if core.HasEnchantEffect(enchant.EffectId) {
		return EffectParseResultSuccess
	}

	enchantingSpell, ok := enchantSpellEffects[int(enchant.EffectId)]
	if !ok {
		return EffectParseResultInvalid
	}

	renderedTooltip := renderSpellTooltip(instance, enchantingSpell.SpellID)

	grp, exists := groupMapProc["Enchants"]
	if !exists {
		grp = Group{Name: "Enchants"}
	}

	result := EffectParseResultSuccess
	for _, routing := range routeEnchantProcs(slots, instance, renderedTooltip) {
		grp.Entries = append(grp.Entries, &Entry{
			Tooltip:   strings.Split(renderedTooltip, "\n"),
			Variants:  []*Variant{{ID: int(enchant.EffectId), Name: enchant.Name, SpellID: int(enchantingSpell.SpellID)}},
			Proc:      routing,
			Supported: routing.Supported(),
		})

		if !routing.Supported() {
			StoreMissingEffect("EnchantEffects", enchant.Name, Variant{
				ID:      int(enchant.EffectId),
				Name:    renderedTooltip,
				SpellID: int(enchant.SpellId),
			})
			result = EffectParseResultUnsupported
		}
	}
	groupMapProc["Enchants"] = grp

	return result
}

// The rows an enchant's procs are resolved from, one per slot. A combat spell and an equip aura that
// apply the same spell are one proc stated twice (Crusader 1900), so one of the two is kept: the
// combat spell, unless only the aura states a rate. An enchant registers once, so where two slots
// could, the first does.
func routeEnchantProcs(slots []dbc.EnchantProcSlot, instance *dbc.DBC, grantTooltip string) []*ProcRouting {
	routings := make([]*ProcRouting, len(slots))
	for i, slot := range slots {
		routings[i] = routeEnchantSlot(slot, instance, grantTooltip)
	}

	for i, combat := range slots {
		if !combat.IsCombatSpell {
			continue
		}
		for j, aura := range slots {
			if aura.IsCombatSpell || routings[j] == nil || aura.AppliesSpellID != combat.AppliesSpellID {
				continue
			}

			kept, dropped := i, j
			if statesNoRate(routings[i]) && !statesNoRate(routings[j]) {
				kept, dropped = j, i
			}
			routings[kept].Summary += fmt.Sprintf("; slot spell %d applies the same spell and is not registered", slots[dropped].SpellID)
			routings[dropped] = nil
			break
		}
	}

	var kept []*ProcRouting
	registers := false
	for _, routing := range routings {
		if routing == nil {
			continue
		}
		if routing.Supported() && registers {
			routing.Unsupported = append(routing.Unsupported, "another slot of the enchant registers")
		}
		registers = registers || routing.Supported()
		kept = append(kept, routing)
	}

	return kept
}

func renderSpellTooltip(instance *dbc.DBC, spellID int) string {
	tooltip, _ := tooltip.ParseTooltip(instance.Spells[spellID].Description, tooltip.DBCTooltipDataProvider{DBC: instance}, int64(spellID))
	return tooltip.String()
}

func statesNoRate(routing *ProcRouting) bool {
	return slices.Contains(routing.Unsupported, spelldata.ReasonStatesNoRate)
}

// One slot's proc. The trigger is the slot's spell; the buff is what the shipped entry says it
// applies, where it resolves stats, the spell the slot applies where that multiplies stats, and
// otherwise the spell it deals damage through.
func routeEnchantSlot(slot dbc.EnchantProcSlot, instance *dbc.DBC, grantTooltip string) *ProcRouting {
	applied := slot.AppliesSpellID
	if applied == 0 {
		applied = slot.SpellID
	}

	// A buff stated as a percentage of a stat resolves no flat stats: the sim reads the multipliers
	// off the buff's own row.
	effect, hasStats := dbc.EnchantSlotEffect(slot.SpellID)
	multipliesStats := !hasStats && len(spelldata.PercentStats(spelldata.Find(int32(applied)), 0)) > 0

	buffSpellID := slot.SpellID
	switch {
	case hasStats:
		buffSpellID = int(effect.BuffId)
	case multipliesStats:
		buffSpellID = applied
	}

	routing := routeProc(slot.SpellID, buffSpellID, slot.IsCombatSpell)
	if slot.IsCombatSpell {
		routing.ProcChancePct = slot.ChancePct
		routing.Unsupported = spelldata.CombatEnchantUnsupported(spelldata.Find(int32(slot.SpellID)), slot.ChancePct > 0)
	} else {
		routing.Unsupported = spelldata.EnchantAuraUnsupported(spelldata.Find(int32(slot.SpellID)))
		routing.readEnchantTooltip(grantTooltip)
	}

	damage := dbc.ResolveDamageEffect(slot.SpellID)
	heal := procHealSpell(int32(slot.SpellID))
	absorb := procAbsorbSpell(int32(slot.SpellID))
	switch {
	case hasStats, multipliesStats:
		routing.requireABuffDuration()
	case damage != nil:
		routing.asDamage(int32(damage.SpellID))
		if mask := spelldata.Find(int32(damage.SpellID)).TargetCreatureType; mask != 0 {
			routing.Unsupported = append(routing.Unsupported,
				fmt.Sprintf("the damage spell hits %s (TargetCreatureType %d) only", creatureTypeNames(mask), mask))
		}
	case heal != 0:
		routing.asHeal(heal)
	case absorb != 0:
		routing.asAbsorb(absorb)
	default:
		routing.Unsupported = append(routing.Unsupported,
			fmt.Sprintf("the enchant's effect entry resolves no stats from %d (%s)", applied, spellEffectKinds(instance, applied)))
	}

	if slot.ChancePct > 0 {
		routing.Summary += fmt.Sprintf("; the enchantment states %d%%", slot.ChancePct)
	}

	if routing.ProcHint != 0 {
		trigger := spelldata.Find(int32(slot.SpellID))
		decoded := core.DecodeProcTypeMask(trigger.ProcFlags, trigger.ProcHint|routing.ProcHint)
		routing.Summary += fmt.Sprintf("; the enchant's tooltip restricts it to %s", asCoreOutcome(decoded.Outcome))
	}

	return routing
}

// The client's CreatureType ids, which TargetCreatureType names as 1 << (id - 1). The damage proc does
// not read the restriction.
var creatureTypes = []string{1: "beasts", 2: "dragonkin", 3: "demons", 4: "elementals", 5: "giants",
	6: "undead", 7: "humanoids", 8: "critters", 9: "mechanicals"}

func creatureTypeNames(mask int32) string {
	var names []string
	for id := 1; id <= 32; id++ {
		if mask&(1<<(id-1)) == 0 {
			continue
		}
		if id < len(creatureTypes) {
			names = append(names, creatureTypes[id])
		} else {
			names = append(names, fmt.Sprintf("creature type %d", id))
		}
	}
	return strings.Join(names, " and ")
}

func spellEffectKinds(instance *dbc.DBC, spellID int) string {
	var kinds []string
	for _, effect := range instance.SpellEffectsInOrder(spellID) {
		if effect.EffectType == dbcenums.E_APPLY_AURA {
			kinds = append(kinds, effect.EffectAura.String())
		} else {
			kinds = append(kinds, effect.EffectType.String())
		}
	}
	return strings.Join(kinds, ", ")
}

// An equip aura's trigger and rate as the enchant's tooltip states them. The store reads both off the
// aura's own description, which the client leaves empty on these auras: the wording is on the spell
// that grants the enchant.
func (r *ProcRouting) readEnchantTooltip(tooltip string) {
	trigger := spelldata.Find(int32(r.TriggerSpellID))
	hints := procTooltipHints(tooltip)

	if hints.Matches(core.ProcHintNamedAbility) && !trigger.ProcHint.Matches(core.ProcHintNamedAbility) {
		r.Unsupported = append(r.Unsupported, "named ability")
	}
	if hints.Matches(core.ProcHintOutcomeTaken) && !trigger.ProcHint.Matches(core.ProcHintOutcomeTaken) {
		r.Unsupported = append(r.Unsupported, "an outcome the proc mask has no bit for")
	}
	if trigger.ProcChanceSource == spelldata.ProcChanceAlways && trigger.RPPM == 0 &&
		enchantTooltipStatesAnUnknownRate(tooltip) && !statesNoRate(r) {
		r.Unsupported = append(r.Unsupported, spelldata.ReasonStatesNoRate)
	}

	r.ProcHint = hints & (core.ProcHintAttackDodged | core.ProcHintAttackParried) &^ trigger.ProcHint
}

// "Often", "sometimes" and "occasionally" are how an enchant's tooltip says its proc has a rate the
// rows do not carry. "No more often than" states a cooldown instead.
var enchantRateWordMatcher = regexp.MustCompile(`(?i)\b(often|sometimes|occasionally)\b`)
var cooldownWordingMatcher = regexp.MustCompile(`(?i)more often than`)

func enchantTooltipStatesAnUnknownRate(tooltip string) bool {
	return tooltipStatesAnUnknownRate(tooltip) ||
		enchantRateWordMatcher.MatchString(cooldownWordingMatcher.ReplaceAllString(tooltip, ""))
}

func ParseTooltipForMissingEffect(parsed *proto.UIItem, itemEffect *proto.ItemEffect, instance *dbc.DBC, groupMap map[string]Group, groupMapName string) {
	if parsed.ScalingOptions[0].Ilvl >= MIN_EFFECT_ILVL {
		// Effect was already manually implemented
		if core.HasItemEffect(parsed.Id) {
			return
		}

		tooltipString, id := dbc.GetItemEffectSpellTooltip(int(parsed.Id), int(itemEffect.BuffId))
		tooltip, _ := tooltip.ParseTooltip(tooltipString, tooltip.DBCTooltipDataProvider{DBC: instance}, int64(id))

		grp, exists := groupMap[groupMapName]
		if !exists {
			grp = Group{Name: groupMapName}
		}

		if tooltip != nil {
			renderedTooltip := tooltip.String()
			entry := Entry{
				Tooltip:   strings.Split(renderedTooltip, "\n"),
				Supported: false,
				Variants: []*Variant{
					{
						ID:      int(parsed.Id),
						Name:    parsed.Name,
						SpellID: int(itemEffect.BuffId),
					},
				},
			}

			grp.Entries = append(grp.Entries, &entry)
			groupMap[groupMapName] = grp

			// Flavour auras carry no mechanic worth implementing and only add noise to the
			// report. Suppressing the report only, never the group entry: those
			// Supported: false entries take part in the variant grouping that decides whether
			// a whole variant set is emitted live or commented, so dropping one can flip a
			// real registration.
			if _, ignored := IgnoreMissingEffectBySpellID[int(itemEffect.BuffId)]; ignored {
				return
			}

			if len(dbc.EffectStats(itemEffect)) == 0 || !entry.Supported {
				StoreMissingEffect("ItemEffects", parsed.Name, Variant{
					ID:      int(parsed.Id),
					Name:    renderedTooltip,
					SpellID: int(itemEffect.BuffId),
				})
			}
		}
	}
}

// A critical strike named as the trigger. The first clause reads the wording that puts something
// after the crit ("critical strikes have a", "critical hits grant"), which is why the character
// after it may not open "by", "for" or "chance" - those three are how a tooltip states a magnitude
// instead. The second reads the trigger stated from the caster's side, "your critical strikes",
// where the plural is what separates it from the magnitudes: those all read "critical strike
// chance", "critical strike damage" or "critical strike rating", in the singular.
var critMatcher = regexp.MustCompile(`critical ([^\s]+|damage,?)( chance)? [^fbc]|[Yy]our [a-z ]{0,20}critical strikes`)
var pureHealMatcher = regexp.MustCompile(`healing spells`)
var hasHealMatcher = regexp.MustCompile(`heal(ing)?[^,]`)
var hasGenericMatcher = regexp.MustCompile(`a spell`)

// A trigger condition stated as an attack outcome. Deliberately matches the condition clause rather
// than the words themselves: "increases your dodge rating" is a stat on hundreds of items, while
// "when one of your spells is resisted" is a trigger. "After a block" is the third clause the
// client uses, on the Battlegear of Wrath parry (23548); "after you dodge" is not in here because
// the only rows wording it that way are Counterattack's, where it gates the button rather than a
// proc. A miss is not in here either: the one row naming one, 456394, means its own attack missing
// rather than an attack on it.
var outcomeConditionMatcher = regexp.MustCompile(`(?i)when .{0,60}?(is|are) resisted|((each|every) time|when|whenever) you (block|dodge|parry)|after a (block|dodge|parry)`)

// The wearer's own attack dodged or parried, named as the trigger: Recovery's "when you are Parried
// or Dodged". The same words also state a magnitude - "reduces the chance for your attacks to be
// dodged or parried" - so only the condition clause counts.
var attackAvoidedMatcher = regexp.MustCompile(`(?i)((each|every) time|when|whenever) (you are|your (melee )?attacks? (is|are)) (dodged|parried)((,? or|,) (dodged|parried))*`)
var attackDodgedMatcher = regexp.MustCompile(`(?i)dodged`)
var attackParriedMatcher = regexp.MustCompile(`(?i)parried`)

func attackAvoidedHints(tooltip string) core.ProcHint {
	var hints core.ProcHint
	for _, clause := range attackAvoidedMatcher.FindAllString(tooltip, -1) {
		if attackDodgedMatcher.MatchString(clause) {
			hints |= core.ProcHintAttackDodged
		}
		if attackParriedMatcher.MatchString(clause) {
			hints |= core.ProcHintAttackParried
		}
	}
	return hints
}

// A tooltip stating that the effect only happens sometimes. Where the data pairs that with a 100%
// rate, the real rate is the one thing the data does not carry.
var statedChanceMatcher = regexp.MustCompile(`(?i)chance (to|of|when)|has a chance`)

// What a sentence names as feeding the proc. A clause stating a chance says nothing about a rate
// unless it also says what the chance is rolled on or what it does: Force Reactive Disk's "This also
// has a chance of damaging the shield" is the shield's durability, in a sentence of its own, and the
// -ing is what keeps it apart from the damage a proc deals.
var procTriggerClauseMatcher = regexp.MustCompile(`(?i)melee|ranged|attack|swing|strike|cast|spell|whenever|each time|on hit|block|weapon|damage\b`)

// A chance that is the magnitude rather than the rate: "increases the critical effect chance of your
// Lesser Healing Wave", "grants increased chance to Block", "Chance to trigger Overload increased by
// an additional 5%". The word the effect modifies sits within a clause of the chance either way.
var increasedChanceMatcher = regexp.MustCompile(`(?i)increase[sd]?[^.]{0,40}?chance|chance[^.]{0,30}?increase[sd]?`)

// Where one statement ends and the next begins. Crude on purpose: an abbreviation or a decimal
// splits a sentence in two, and neither changes which half carries the trigger.
var sentenceBreak = regexp.MustCompile(`[.!?](?:\s|$)|\r?\n`)

// Whether the tooltip says the effect only happens sometimes and leaves the rate unsaid. Read from
// the sentence that states the trigger rather than from the whole text: "increases your critical
// strike chance" is a magnitude on hundreds of spells, and a chance in a sentence that names nothing
// the proc fires on is about something else entirely.
func tooltipStatesAnUnknownRate(description string) bool {
	for _, sentence := range sentenceBreak.Split(description, -1) {
		if !statedChanceMatcher.MatchString(sentence) {
			continue
		}
		if !procTriggerClauseMatcher.MatchString(sentence) {
			continue
		}
		if increasedChanceMatcher.MatchString(sentence) {
			continue
		}

		return true
	}

	return false
}

// Wording that names the cast itself as the trigger rather than the spell landing:
// "each time you cast a spell", "chance on successful spellcast", "chance on spell cast".
var castTriggerMatcher = regexp.MustCompile(`(?i)you cast|on spell ?cast|spellcast`)

// A trigger clause restricted to one named ability: "Your Shock spells", "Your Moonfire ability",
// "Your casts of Greater Heal", "Your Shadow Bolt has", "When you cast Flash of Light". The capital
// is what carries the meaning - an unrestricted trigger reads "your spell critical strikes" or
// "each time you cast a spell", with nothing capitalized to name.
// The client writes a conditional list where the item shows one name - Eternal Power's "Your casts
// of $?s2060[Greater Heal]..." - so the "casts of" clause reads the possessive alone and leaves what
// follows to the tooltip renderer. The other two clauses need the capital: "when you cast a spell"
// and "your melee attacks have" name no ability.
var namedAbilityMatcher = regexp.MustCompile(`[Yy]our [A-Z][A-Za-z']*( [A-Z][A-Za-z']*)* (spell|spells|ability|abilities|has|have)` +
	`|[Yy]our casts of` +
	`|[Ww]hen you cast [A-Z]`)

// What core.DecodeProcTypeMask cannot read off the mask: the trigger wording around it. The named
// ability and outcome condition bits are read here as well, for the store's rows, and say nothing
// new to the item generator: BuildSpellProcInfo refuses a tooltip carrying either before it decodes
// anything, so the outcome bit the decode does read never reaches an item.
func procTooltipHints(tooltip string) core.ProcHint {
	var hints core.ProcHint

	if castTriggerMatcher.MatchString(tooltip) {
		hints |= core.ProcHintCastTrigger
	}

	if critMatcher.MatchString(tooltip) {
		hints |= core.ProcHintCrit
	}

	// An unrestricted "a spell" counts as heal evidence too: next to a helpful-spell bit it is the
	// wording of a proc that fires off any spell the character casts, healing included.
	if hasHealMatcher.MatchString(tooltip) || hasGenericMatcher.MatchString(tooltip) {
		hints |= core.ProcHintHeals
	}

	if pureHealMatcher.MatchString(tooltip) {
		hints |= core.ProcHintPureHeal
	}

	if namedAbilityMatcher.MatchString(tooltip) {
		hints |= core.ProcHintNamedAbility
	}

	if outcomeConditionMatcher.MatchString(tooltip) {
		hints |= core.ProcHintOutcomeTaken
	}

	hints |= attackAvoidedHints(tooltip)

	return hints
}

// Derives what adds a stack to an accumulating aura, from the container spell rather than from
// the one that opens the window. buff_id is the container by then: the parser rebases the effect
// onto it precisely because that is where the duration and these proc flags live.
func buildStackProcInfo(itemEffect *proto.ItemEffect, instance *dbc.DBC, tooltip string) *ProcInfo {
	if itemEffect.StackingAura == nil || itemEffect.GetStackProc() == nil {
		return nil
	}

	container, ok := instance.Spells[int(itemEffect.BuffId)]
	if !ok {
		return nil
	}

	info, supported := BuildSpellProcInfo(&container, tooltip, proto.ItemType_ItemTypeUnknown)
	if !supported {
		return nil
	}
	return &info
}

func BuildProcInfo(parsed *proto.UIItem, itemEffectID int, instance *dbc.DBC, tooltip string) (ProcInfo, bool) {
	itemEffect := dbc.GetItemEffectForBuffID(int(parsed.Id), itemEffectID)
	if itemEffect == nil {
		return ProcInfo{}, false
	}

	// if we have multiple spells find the first that has a proc aura assigned
	procId := itemEffect.SpellID
	procSpell, ok := instance.Spells[int(procId)]
	if !ok {
		panic(fmt.Sprintf("Could not find proc aura %d spell for item effect %d.\n", procId, parsed.Id))
	}

	// A "Chance on hit" effect is a weapon proc: the game casts it off the hit itself and never
	// consults a ProcTypeMask, so it takes the on-hit shape whatever the proc spell says.
	isWeaponProc := itemEffect.TriggerType == dbc.ITEM_SPELLTRIGGER_CHANCE_ON_HIT

	itemType := proto.ItemType_ItemTypeUnknown
	if isWeaponProc {
		itemType = proto.ItemType_ItemTypeWeapon
	}

	procInfo, supported := BuildSpellProcInfo(&procSpell, tooltip, itemType)
	procInfo.setIsWeaponProc(isWeaponProc)

	if SpellHasDummyEffect(int(procId), instance) {
		return procInfo, false
	}

	return procInfo, supported
}

func BuildSpellProcInfo(procSpell *dbc.Spell, tooltip string, itemType proto.ItemType) (ProcInfo, bool) {
	var info = ProcInfo{
		RequireDamageDealt:  true,
		MaxCumulativeStacks: procSpell.MaxCumulativeStacks,
	}

	requiresOutcome := true
	onHitProc := false

	// On hit proc
	if itemType == proto.ItemType_ItemTypeWeapon {
		onHitProc = true
		info.Callback |= core.CallbackOnSpellHitDealt
		info.ProcMask |= core.ProcMaskUnknown
	}

	if itemType == proto.ItemType_ItemTypeRanged {
		info.Callback |= core.CallbackOnSpellHitDealt
		info.ProcMask |= core.ProcMaskRanged
	}

	if procSpell.OnlyProcsFromClassAbilities() {
		info.ClassSpellsOnly = true
	}

	// A spell-family filter the generated proc cannot reproduce. Two sources of evidence for the
	// same thing: the spell names a family in SpellClassMask, or its tooltip names one. On item
	// procs TBC stores the mask as either nil or all-zero and keeps the real filter server-side,
	// so for those the tooltip is the only evidence there is - "Your Shock spells", "Your Moonfire
	// ability". Generating one anyway would proc it off every spell instead of that one.
	if slices.ContainsFunc(procSpell.SpellClassMask, func(mask int) bool { return mask != 0 }) ||
		namedAbilityMatcher.MatchString(tooltip) {
		return info, false
	}

	// An outcome the client does not record. ProcTypeMask has no dodge, parry, block or resist bit -
	// TrinityCore keeps those in a HitMask of its own - so a tooltip stating one as the trigger
	// condition is the only evidence there is. Eye of Magtheridon fires on a resisted spell, and
	// without this it generates as a 100%-per-cast buff. Crit is deliberately not in here: it is the
	// one outcome the generated shape can express, and critMatcher below assigns it.
	if outcomeConditionMatcher.MatchString(tooltip) {
		return info, false
	}

	// A buff the game spends by charges rather than by time: World Breaker's 900 crit rating is
	// consumed by the next two melee hits, which is ProcCharges 2 on 36111. The generated shape
	// knows only durations, so it would hold the buff for the whole window instead. On-use effects
	// are unaffected - their charges are uses of the item, and they do not come through here.
	if procSpell.ProcCharges > 0 {
		return info, false
	}

	// The bit table, decoded in core so the sim reads the same mask the same way. Bits the decode
	// does not model come back in Unsupported and are dropped here rather than refused: a mask
	// naming one next to bits that do model something still generates a listener, and that
	// listener is deliberately the narrower trigger: it hears the hits the sim knows and stays
	// silent on the rest.
	//
	// One rule of the bit table cannot live in core. The heal branch's cleanup for a helpful-only
	// mask also has to clear the OnSpellHitDealt seeded above for a ranged item, which the decode
	// cannot see and this merge cannot undo. No ranged item carries a helpful-only mask with a
	// healing tooltip, and the pureHeal strip below still covers the "healing spells" wording.
	if !onHitProc && len(procSpell.ProcTypeMask) > 0 {
		decoded := core.DecodeProcTypeMask(
			[2]uint32{uint32(procSpell.ProcTypeMask[0]), uint32(procSpell.ProcTypeMask[1])},
			procTooltipHints(tooltip),
		)

		info.Callback |= decoded.Callback
		info.ProcMask |= decoded.ProcMask
		info.RequireDamageDealt = decoded.RequireDamageDealt
		info.Outcome = decoded.Outcome
		requiresOutcome = false
	}

	// Proc-ness is a flag on the listener, not a hit kind in the mask. ProcMaskSpellDamageProc is
	// deliberately never emitted: that is the weapon-imbue shape (Flametongue, rogue poisons), kept
	// a distinct hit kind so that Shiffar's Nexus-Horn, Wrath of Cenarius and Robe of the Elder
	// Scribes proc off Ignite, Elemental Overload and Hurricane's DoT rather than off imbue crits.
	info.CanProcFromProcs = procSpell.CanProcFromProcs()
	info.HonoursWeaponProcSuppression = procSpell.IsWeaponProcAura()

	if requiresOutcome {
		if critMatcher.MatchString(tooltip) {
			info.Outcome = core.OutcomeCrit
		} else {
			info.Outcome = core.OutcomeLanded
		}
	}

	// check for pure healing spell
	if pureHealMatcher.MatchString(tooltip) {
		info.Callback &= ^core.CallbackOnSpellHitDealt
		info.Callback &= ^core.CallbackOnPeriodicDamageDealt
	}

	// A trigger with no callback never fires, and factory_ProcStatBonusEffect returns early on
	// exactly that, so such an effect cannot be generated. The test used to also require an
	// empty Outcome and ProcMask, which it can never have: when requiresOutcome is true the
	// Outcome is set a few lines up, and when it is false the whole term is false, so nothing
	// was ever refused here and empty-callback effects were emitted as live registrations that
	// silently did nothing.
	return info, info.Callback != core.CallbackEmpty
}

func StoreMissingEffect(effectType string, name string, variant Variant) {
	if missingEffectsMap[effectType] == nil {
		missingEffectsMap[effectType] = map[int32]MissingItemEffect{}
	}
	id := int32(variant.ID)
	if missingEffectsMap[effectType][id].Effects == nil {
		missingEffectsMap[effectType][id] = MissingItemEffect{
			ItemID:  id,
			Name:    name,
			Effects: []Variant{},
		}
	}
	itemEntry := missingEffectsMap[effectType][id]
	haveEffect := false
	for _, effect := range itemEntry.Effects {
		if effect.SpellID == variant.SpellID {
			haveEffect = true
			break
		}
	}
	if haveEffect {
		return
	}

	itemEntry.Effects = append(
		itemEntry.Effects,
		variant,
	)
	missingEffectsMap[effectType][id] = itemEntry
}

func asCoreCallback(callback core.AuraCallback) string {
	callbacks := []string{}
	for i := range 32 {
		callbackFlag := core.AuraCallback(1 << i)
		if callbackFlag >= core.CallbackLast {
			break
		}

		if callback.Matches(callbackFlag) {
			callbacks = append(callbacks, "core."+callbackFlag.String())
		}
	}

	if len(callbacks) == 0 {
		return "core.CallbackEmpty"
	}

	return strings.Join(callbacks, " | ")
}

func asCoreProcMask(procMask core.ProcMask) string {
	procs := []string{}
	for i := range 32 {
		procFlag := core.ProcMask(1 << i)
		if procFlag >= core.ProcMaskLast {
			break
		}

		if procMask.Matches(procFlag) {
			procs = append(procs, "core."+procFlag.String())
		}
	}

	if len(procs) == 0 {
		return "core.ProcMaskUnknown"
	}
	return strings.Join(procs, " | ")
}

func asCoreOutcome(outcome core.HitOutcome) string {
	if outcome == core.OutcomeCrit {
		return "core.OutcomeCrit"
	}

	if outcome.Matches(core.OutcomeLanded) {
		return "core.OutcomeLanded"
	}

	var avoided []string
	if outcome.Matches(core.OutcomeDodge) {
		avoided = append(avoided, "core.OutcomeDodge")
	}
	if outcome.Matches(core.OutcomeParry) {
		avoided = append(avoided, "core.OutcomeParry")
	}
	if len(avoided) > 0 {
		return strings.Join(avoided, " | ")
	}

	return "core.OutcomeEmpty"
}

func (entry *Entry) AddVariant(variant *Variant) {
	entry.Variants = append(entry.Variants, variant)
	sort.Slice(entry.Variants, func(i, j int) bool {
		return entry.Variants[i].ID < entry.Variants[j].ID
	})
}

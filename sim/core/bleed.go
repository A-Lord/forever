package core

// The tag every bleed aura carries. spelldata.AuraConfig sets it on the auras of rows with the bleed
// mechanic, so a bleed is found without naming its spell.
const BleedTag = "Bleed"

// Whether the target is bleeding, as this unit sees it: a bleed aura is up on the target, or the unit
// assumes one for the raid members the sim does not run.
func (unit *Unit) TargetIsBleeding(target *Unit) bool {
	return unit.PseudoStats.AssumeBleedActive || target.HasActiveAuraWithTag(BleedTag)
}

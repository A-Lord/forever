package core

import (
	"fmt"

	"github.com/wowsims/forever/sim/core/proto"
)

// True while the target bleeds, from any caster's bleed, or while the player assumes a bleed for the
// raid members the sim does not run.
type APLValueBleedIsActive struct {
	DefaultAPLValueImpl
	unit   *Unit
	target UnitReference
}

func (rot *APLRotation) newValueBleedIsActive(config *proto.APLValueBleedIsActive, _ *proto.UUID) APLValue {
	target := rot.GetTargetUnit(config.TargetUnit)
	if target.Get() == nil {
		return nil
	}
	return &APLValueBleedIsActive{
		unit:   rot.unit,
		target: target,
	}
}
func (value *APLValueBleedIsActive) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueBleedIsActive) GetBool(_ *Simulation) bool {
	target := value.target.Get()
	return target != nil && value.unit.TargetIsBleeding(target)
}
func (value *APLValueBleedIsActive) String() string {
	return fmt.Sprintf("Bleed Is Active(%s)", value.target.Get().Label)
}

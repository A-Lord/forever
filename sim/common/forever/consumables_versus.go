package forever

import "github.com/wowsims/forever/sim/common/shared"

func init() {
	shared.NewSpellDataVersusPotion(9224)                   // Potion of Demonslaying
	shared.NewSpellDataVersusTemporaryEnchant(23122, 28893) // Consecrated Sharpening Stone, enchant 2684's equip spell
	shared.NewSpellDataVersusTemporaryEnchant(23123, 28890) // Blessed Wizard Oil, enchant 2685's equip spell
}

import type { UIEnchant as Enchant } from '@generated/proto/ui';
import { useActionIdWowheadDataset } from '@sim/hooks/useActionIdWowheadDataset';
import { ActionId } from '@sim/proto/action_id';
import { getEnchantDescription } from '@sim/proto/enchants';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import { useEffect, useMemo, useState } from 'react';

export interface EnchantLabelProps {
	className: string;
	testId?: string;
	enchant?: Enchant | null;
	onActivate?: () => void;
	tooltipId?: string;
}

export const EnchantLabel = ({ className, testId, enchant, onActivate, tooltipId }: EnchantLabelProps) => {
	const [description, setDescription] = useState('');

	const actionId = useMemo(
		() => (enchant ? (enchant.spellId ? ActionId.fromSpellId(enchant.spellId) : ActionId.fromItemId(enchant.itemId)) : undefined),
		[enchant],
	);

	useEffect(() => {
		if (!enchant) {
			setDescription('');
			return;
		}
		let live = true;
		getEnchantDescription(enchant)
			.then(text => {
				if (live) setDescription(text);
			})
			.catch(() => {
				if (live) setDescription(enchant.name);
			});
		return () => {
			live = false;
		};
	}, [enchant]);

	const wowheadProps = useActionIdWowheadDataset(actionId);

	const href = actionId ? (actionId.spellId ? ActionId.makeSpellUrl(actionId.spellId) : ActionId.makeItemUrl(actionId.itemId)) : undefined;

	if (!enchant) return null;

	return (
		<WowheadAnchor className={className} testId={testId} href={href} onActivate={onActivate} data-tooltip-id={tooltipId} {...wowheadProps}>
			{description}
		</WowheadAnchor>
	);
};

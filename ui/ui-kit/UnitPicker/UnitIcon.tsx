import { Icon, ICON_NAMES, type IconName } from '@ui-kit/Icon';

export interface UnitIconProps {
	/** A `UnitValue.iconUrl`: `fa-<IconName>` or an image url. */
	iconUrl: string;
}

const isIconName = (name: string): name is IconName => (ICON_NAMES as readonly string[]).includes(name);

/** The two shapes a `UnitValue` icon takes, as one element. A `fa-` name outside `IconName` has no glyph and renders nothing. */
export const UnitIcon = ({ iconUrl }: UnitIconProps) => {
	if (iconUrl.startsWith('fa-')) {
		const name = iconUrl.slice('fa-'.length);
		return isIconName(name) ? <Icon name={name} style="base" className="ui-unit-icon" data-testid="unit-picker-item-icon" /> : null;
	}
	return <img className="ui-unit-icon" data-testid="unit-picker-item-icon" src={iconUrl} alt="" />;
};

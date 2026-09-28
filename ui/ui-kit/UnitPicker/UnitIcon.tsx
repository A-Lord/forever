import { Icon, type IconName } from '@ui-kit/Icon';

export interface UnitIconProps {
	/** A `UnitValue.iconUrl`: a Font Awesome glyph name or an image url. */
	iconUrl: string;
}

/** The two shapes a `UnitValue` icon takes, as one element. */
export const UnitIcon = ({ iconUrl }: UnitIconProps) => {
	if (iconUrl.startsWith('fa-')) {
		// `iconUrl` is a free string, so a glyph outside `IconName` (e.g. `users`) reaches `Icon` as is.
		const name = iconUrl.slice('fa-'.length) as IconName;
		return <Icon name={name} style="base" className="ui-unit-icon" data-testid="unit-picker-item-icon" />;
	}
	return <img className="ui-unit-icon" data-testid="unit-picker-item-icon" src={iconUrl} alt="" />;
};

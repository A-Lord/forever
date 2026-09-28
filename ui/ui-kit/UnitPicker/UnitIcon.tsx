import clsx from 'clsx';

export interface UnitIconProps {
	/** A `UnitValue.iconUrl`: a Font Awesome glyph name or an image url. */
	iconUrl: string;
}

/** The two shapes a `UnitValue` icon takes, as one element. */
export const UnitIcon = ({ iconUrl }: UnitIconProps) => {
	if (iconUrl.startsWith('fa-')) return <i className={clsx('fa', iconUrl, 'ui-unit-icon')} data-testid="unit-picker-item-icon" />;
	return <img className="ui-unit-icon" data-testid="unit-picker-item-icon" src={iconUrl} alt="" />;
};

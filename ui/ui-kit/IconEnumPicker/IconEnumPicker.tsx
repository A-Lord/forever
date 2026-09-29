import { Field } from '@ui-kit/Field';
import type { ControlledProps, FieldChromeProps, ResolvedIcon } from '@ui-kit/Field/types';
import { Menu } from '@ui-kit/Menu';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId } from 'react';

import { resolvedIconStyle } from '../utils/icons';
import { IconEnumOptionItem, IconText } from './IconEnumOption';
import { type IconEnumOption, IconEnumPickerDirection } from './types';
import { iconStyleOf } from './utils';

export interface IconEnumPickerProps<T> extends FieldChromeProps, ControlledProps<T> {
	options: Array<IconEnumOption<T>>;
	equals: (a: T, b: T) => boolean;
	/** The value that reads as inactive. */
	zeroValue: T;
	/** The icon for a value no option carries. */
	backupIcon?: ResolvedIcon;
	/** Shown when hovering the button. */
	tooltip?: string;
	/** The side the menu opens on. */
	direction?: IconEnumPickerDirection;
	numColumns?: number;
	onOpenChange?: (open: boolean) => void;
}

export const IconEnumPicker = <T,>({
	options,
	equals,
	zeroValue,
	backupIcon,
	tooltip,
	direction,
	numColumns,
	onOpenChange,
	value,
	onChange,
	disabled,
	className,
	testId,
	...chrome
}: IconEnumPickerProps<T>) => {
	const tooltipId = useId();

	const selected = options.find(option => equals(option.value, value));
	const selectedHidden = !!selected?.hidden;
	const href = selectedHidden ? undefined : selected ? selected.icon?.href : backupIcon?.href;

	const active = !disabled && !equals(value, zeroValue);

	const disabledAttribute = (disabled ? { disabled: true } : {}) as Record<string, boolean>;

	const horizontal = direction === IconEnumPickerDirection.Horizontal;
	const tooltips = !!tooltip || options.some(option => !!option.tooltip);

	return (
		<Field
			{...chrome}
			className={['ui-icon-field relative', className]}
			testId={testId ?? 'icon-enum-picker-root'}
			disabled={disabled}
			leading={
				<Menu
					onOpenChange={onOpenChange}
					surface="none"
					// The value's text sits on the swatch, the way IconPicker's counter does: with one icon
					// shared by every rank, the text is what tells the ranks apart.
					trigger={selected?.text !== undefined && !selectedHidden ? <IconText text={selected.text} /> : undefined}
					// No href when nothing is selected: React refuses javascript:void(0), and nativeButton={false} keeps the anchor focusable.
					triggerRender={<a href={href || undefined} {...disabledAttribute} />}
					triggerProps={{
						nativeButton: false,
						openOnHover: true,
						delay: 0,
						className: 'ui-icon-picker-swatch transition-none',
						'data-testid': 'icon-enum-picker-button',
						'data-active': active ? '' : undefined,
						style: selectedHidden ? undefined : selected ? iconStyleOf(selected) : resolvedIconStyle(backupIcon),
						'data-whtticon': 'false',
						'data-disable-wowhead-touch-tooltip': 'true',
						...tooltipAnchorProps(tooltip ? tooltipId : undefined, tooltip),
					}}
					side={horizontal ? 'right' : 'bottom'}
					align="start"
					sideOffset={-1}
					// `positionMethod="fixed"`, not the default `absolute`: Base UI renders the positioner `position: fixed` until it has a position, so Floating UI measures it against the viewport, and the switch to `absolute` then reads those viewport coordinates against whatever ancestor is `position: relative`.
					positionMethod="fixed"
					portalProps={{ className: 'contents', 'data-testid': 'icon-enum-picker-portal' }}
					positionerProps={{ 'data-testid': 'icon-enum-picker-positioner' }}
					className="grid border-0 bg-grey"
					popupProps={{
						'data-testid': 'icon-enum-picker-menu',
						style: {
							gridTemplateColumns: numColumns ? `repeat(${numColumns}, 1fr)` : undefined,
							gridAutoFlow: horizontal ? 'column' : undefined,
						},
					}}>
					{options.map((option, index) => (
						<IconEnumOptionItem key={index} option={option} tooltipId={tooltipId} onSelect={() => onChange(option.value)} />
					))}
				</Menu>
			}>
			{tooltips && <Tooltip id={tooltipId} />}
		</Field>
	);
};

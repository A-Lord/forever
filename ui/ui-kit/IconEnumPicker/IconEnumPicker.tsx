import { Field } from '@ui-kit/Field';
import type { ControlledProps, FieldChromeProps, ResolvedIcon } from '@ui-kit/Field/types';
import { IconMenu, IconSwatch } from '@ui-kit/IconSwatch';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { useId } from 'react';

import { resolvedIconStyle } from '../utils/icons';
import { IconEnumOptionItem } from './IconEnumOption';
import { type IconEnumOption, IconEnumPickerDirection } from './types';
import { iconStyleOf } from './utils';

export interface IconEnumPickerProps<T> extends Omit<FieldChromeProps, 'layout'>, ControlledProps<T> {
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

/** Without a label or description the trigger renders bare; with one it sits in a `Field` laid out as `icon`, whatever layout the caller passes. */
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
	const labelled = !!chrome.label || !!chrome.description;

	const selected = options.find(option => equals(option.value, value));
	const selectedHidden = !!selected?.hidden;
	const href = selectedHidden ? undefined : selected ? selected.icon?.href : backupIcon?.href;

	const active = !disabled && !equals(value, zeroValue);

	const horizontal = direction === IconEnumPickerDirection.Horizontal;
	const tooltips = !!tooltip || options.some(option => !!option.tooltip);

	const menu = (
		<IconMenu
			onOpenChange={onOpenChange}
			trigger={
				<IconSwatch
					href={href}
					style={selectedHidden ? undefined : selected ? iconStyleOf(selected) : resolvedIconStyle(backupIcon)}
					active={active}
					disabled={disabled}
					// The value's text sits on the swatch, the way IconPicker's counter does: with one icon
					// shared by every rank, the text is what tells the ranks apart.
					text={selectedHidden ? undefined : selected?.text}
					className={clsx('transition-none', !labelled && className)}
					testId={testId ?? 'icon-enum-picker-root'}
					{...tooltipAnchorProps(tooltip ? tooltipId : undefined, tooltip)}
				/>
			}
			side={horizontal ? 'right' : 'bottom'}
			portalProps={{ 'data-testid': 'icon-enum-picker-portal' }}
			positionerProps={{ 'data-testid': 'icon-enum-picker-positioner' }}
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
		</IconMenu>
	);
	const optionTooltip = tooltips && <Tooltip id={tooltipId} />;

	if (!labelled)
		return (
			<>
				{menu}
				{optionTooltip}
			</>
		);

	return (
		<Field {...chrome} layout="icon" className={className} disabled={disabled} leading={menu}>
			{optionTooltip}
		</Field>
	);
};

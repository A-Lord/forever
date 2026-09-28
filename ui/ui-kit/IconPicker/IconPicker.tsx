import { externalRel } from '@sim/utils/links';
import { Field } from '@ui-kit/Field';
import type { ControlledProps, FieldChromeProps, ResolvedIcon } from '@ui-kit/Field/types';
import { isRightClick } from '@ui-kit/utils/dom';
import clsx from 'clsx';
import type { MouseEvent } from 'react';

import { wowheadAnchorProps } from '../utils/wowhead';
import { ImprovedAnchor } from './ImprovedAnchor';

export interface IconPickerProps extends FieldChromeProps, ControlledProps<number> {
	/** The number of states the icon cycles through: 2 is on/off, 0 is unlimited. */
	states: number;
	icon?: ResolvedIcon;
	/** Shown from states 3 up. */
	improvedIcon?: ResolvedIcon;
	/** Shown from states 4 up. */
	improvedIcon2?: ResolvedIcon;
}

export const IconPicker = ({ states, icon, improvedIcon, improvedIcon2, value, onChange, disabled, className, testId, ...chrome }: IconPickerProps) => {
	const useImprovedIcons = !!improvedIcon;
	const fillImproved1 = states >= 3 && !!improvedIcon;
	const fillImproved2 = states >= 4 && !!improvedIcon2;
	const showCounterText = !improvedIcon && (states > 3 || states === 0);
	const href = icon?.href;

	const handleLeftClick = () => {
		// `disabled` only reaches an <a>, where HTML ignores it, so the state change has to be stopped here.
		if (disabled) return;

		if (states === 0 || value + 1 < states) {
			onChange(value + 1);
		} else if (value > 0) {
			onChange(0);
		}
	};

	const handleRightClick = () => {
		if (disabled) return;

		if (value > 0) {
			onChange(value - 1);
		} else {
			onChange(states === 0 ? 1 : states - 1);
		}
	};

	// `disabled` is not in React's anchor prop types because HTML has no such attribute on <a>.
	const disabledAttribute = (disabled ? { disabled: true } : {}) as Record<string, boolean>;

	// The level container overlays the anchor rather than sitting inside it, so a click on the improved icons reaches these only by carrying them on both.
	const stateEvents = {
		onClick: (event: MouseEvent) => {
			event.preventDefault();
			handleLeftClick();
		},
		onContextMenu: (event: MouseEvent) => event.preventDefault(),
		onMouseDown: (event: MouseEvent) => {
			if (isRightClick(event.nativeEvent)) {
				event.preventDefault();
				handleRightClick();
			}
		},
	};

	const main = (
		<div className="relative size-10 shrink-0">
			<a
				className={clsx('ui-icon-picker-swatch', value > 0 ? 'filter-none' : 'grayscale')}
				data-testid="icon-picker-button"
				data-active={value > 0 ? '' : undefined}
				data-use-counter={!useImprovedIcons && states > 2 ? '' : undefined}
				{...wowheadAnchorProps()}
				// The glyph is a background image and the counter is a sibling, so without this the anchor announces nothing.
				aria-label={icon?.name || undefined}
				target="_blank"
				href={href || undefined}
				rel={externalRel(href, undefined)}
				style={icon ? { backgroundImage: `url('${icon.iconUrl}')` } : undefined}
				{...disabledAttribute}
				{...stateEvents}
			/>
			{states > 2 && (
				<div
					className={clsx('pointer-events-none absolute top-0.5 left-0.5 size-9.5', value > 0 ? 'filter-none' : 'grayscale')}
					data-testid="icon-input-level-container"
					{...stateEvents}>
					{/* Both badges stay visible in opposite corners: at a state below its own, a quadstate
					    icon's grey second badge is the only affordance that a fourth state exists. */}
					{fillImproved1 && improvedIcon && <ImprovedAnchor icon={improvedIcon} testId="icon-input-improved1" active={value > 1} corner="right" />}
					{fillImproved2 && improvedIcon2 && <ImprovedAnchor icon={improvedIcon2} testId="icon-input-improved2" active={value > 2} corner="left" />}
					<span
						className="absolute inset-x-0 bottom-0 bg-scrim text-center text-2xs font-bold whitespace-nowrap text-success"
						data-testid="icon-picker-label"
						data-active={value > 0 ? '' : undefined}>
						{showCounterText ? String(value) : null}
					</span>
				</div>
			)}
		</div>
	);

	return <Field {...chrome} className={['ui-icon-field', className]} testId={testId ?? 'icon-picker-root'} iconField disabled={disabled} leading={main} />;
};

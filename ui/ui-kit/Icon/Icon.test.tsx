import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { iconDefinition } from './definitions';
import { Icon } from './Icon';
import { ICON_NAMES } from './types';

const el = (c: HTMLElement) => c.querySelector('svg')!;
const classes = (c: HTMLElement) => [...el(c).classList];

describe('Icon', () => {
	it('renders an svg carrying its glyph class', () => {
		const { container } = render(<Icon name="circle-question" />);
		expect(el(container).tagName.toLowerCase()).toBe('svg');
		expect(classes(container)).toEqual(['svg-inline--fa', 'fa-circle-question', 'ui-icon']);
	});

	it('draws the glyph of the style it is given', () => {
		const path = (c: HTMLElement) => el(c).querySelector('path')!.getAttribute('d');
		const solid = path(render(<Icon name="circle-question" />).container);
		const regular = path(render(<Icon name="circle-question" style="regular" />).container);
		expect(regular).not.toBe(solid);
		expect(path(render(<Icon name="circle-question" style="base" />).container)).toBe(solid);
		expect(el(render(<Icon name="github" style="brands" />).container).dataset.prefix).toBe('fab');
	});

	it('falls back to solid for a name FontAwesome Free has no regular glyph for', () => {
		expect(iconDefinition('ban', 'regular')).toBe(iconDefinition('ban', 'solid'));
		expect(iconDefinition('star', 'regular')).not.toBe(iconDefinition('star', 'solid'));
	});

	it('resolves FontAwesome 5 aliases to the canonical name', () => {
		expect(classes(render(<Icon name="exclamation-triangle" />).container)).toContain('fa-triangle-exclamation');
		expect(classes(render(<Icon name="triangle-exclamation" />).container)).toContain('fa-triangle-exclamation');
		expect(classes(render(<Icon name="question-circle" />).container)).toContain('fa-circle-question');
		// `rotate-left` is its own glyph (\f2ea), not a spelling of `arrow-rotate-left` (\f0e2).
		expect(iconDefinition('rotate-left', 'solid').icon[3]).toBe('f2ea');
		expect(iconDefinition('arrow-rotate-left', 'solid').icon[3]).toBe('f0e2');
	});

	it('keeps its own name as a class when FontAwesome names the glyph differently', () => {
		expect(classes(render(<Icon name="times" />).container)).toEqual(['svg-inline--fa', 'fa-xmark', 'fa-times', 'ui-icon']);
	});

	it.each(ICON_NAMES)('renders %s from a definition, as its fa- class', name => {
		expect(iconDefinition(name, 'solid')).toBeDefined();
		expect(classes(render(<Icon name={name} />).container)).toContain(`fa-${name}`);
	});

	it('adds size and spin as classes, and the size as data-size', () => {
		const { container } = render(<Icon name="spinner" size="2xl" spin />);
		expect(classes(container)).toEqual(['svg-inline--fa', 'fa-spinner', 'ui-icon', 'fa-spin', 'fa-2xl']);
		expect(el(container).dataset.size).toBe('2xl');
		expect(el(render(<Icon name="spinner" />).container).hasAttribute('data-size')).toBe(false);
	});

	it('adds fa-fw for a fixed-width icon, and nothing without it', () => {
		expect(classes(render(<Icon name="circle-question" size="xl" fixedWidth />).container)).toContain('fa-fw');
		expect(classes(render(<Icon name="circle-question" />).container)).not.toContain('fa-fw');
	});

	it('takes its test id from testId', () => {
		const { getByTestId } = render(<Icon name="cog" testId="settings-icon" />);
		expect(getByTestId('settings-icon').tagName.toLowerCase()).toBe('svg');
	});

	it('keeps a data-testid passed as an attribute when it has no testId', () => {
		const { getByTestId } = render(<Icon name="cog" data-testid="gem-marker" />);
		expect(getByTestId('gem-marker').tagName.toLowerCase()).toBe('svg');
	});

	it('is hidden from assistive tech unless it carries a title', () => {
		expect(el(render(<Icon name="times" />).container).getAttribute('aria-hidden')).toBe('true');
		const titled = el(render(<Icon name="times" title="Close" />).container);
		expect(titled.getAttribute('aria-hidden')).toBeNull();
		const title = titled.querySelector('title')!;
		expect(title.textContent).toBe('Close');
		expect(titled.getAttribute('aria-labelledby')).toBe(title.id);
	});

	// character_stats.tsx anchors a tooltip on the icon itself; a dropped attribute fails silently.
	it('passes an unknown attribute through, so it can anchor its own tooltip', () => {
		const { container } = render(<Icon name="circle-question" data-tooltip-id="bonus-str" />);
		expect(el(container).dataset.tooltipId).toBe('bonus-str');
	});
});

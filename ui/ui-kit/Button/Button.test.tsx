import { render, screen, within } from '@testing-library/react';
import { Icon } from '@ui-kit/Icon';
import { describe, expect, it } from 'vitest';

import { Button } from './Button';

describe('Button', () => {
	it('is a non-submitting button by default', () => {
		render(<Button>Simulate</Button>);
		const button = screen.getByRole('button', { name: 'Simulate' });
		// The tree has several <button>s inside forms with no type, which submit on click.
		expect(button.getAttribute('type')).toBe('button');
		expect(button.className).toBe('ui-button ui-button-primary ui-button-md');
	});

	it('renders an anchor for an href, keeping the button classes', () => {
		render(
			<Button href="https://wowhead.com/forever/item=1" variant="link">
				Item
			</Button>,
		);
		const link = screen.getByRole('link', { name: 'Item' });
		expect(link.tagName).toBe('A');
		expect(link.getAttribute('href')).toBe('https://wowhead.com/forever/item=1');
		expect(link.className).toBe('ui-button-link');
		expect(link.getAttribute('type')).toBeNull();
		expect(screen.queryByRole('button')).toBeNull();
	});

	it('keeps the variant and size classes on an anchor', () => {
		render(
			<Button href="https://github.com/wowsims/forever/releases" variant="outline-light" target="_blank">
				Download
			</Button>,
		);
		const link = screen.getByRole('link', { name: 'Download' });
		expect(link.className).toBe('ui-button ui-button-outline-light ui-button-md');
		expect(link.getAttribute('target')).toBe('_blank');
	});

	it('accepts an anchor only through href, and an icon-only button only with an aria-label', () => {
		// @ts-expect-error `as="a"` is not a Button prop: `href` renders the anchor.
		const byAs = <Button as="a">Item</Button>;
		// @ts-expect-error An icon-only button has no text to announce, so `aria-label` is required.
		const unnamed = <Button iconOnly>×</Button>;
		const named = (
			<Button iconOnly aria-label="Close">
				×
			</Button>
		);
		expect([byAs, unnamed, named]).toHaveLength(3);
	});

	it('renders a label when asked, for a peer-checked radio', () => {
		render(
			<Button as="label" htmlFor="chart-view-rotation" variant="outline-primary">
				Rotation
			</Button>,
		);
		const label = screen.getByText('Rotation');
		expect(label.tagName).toBe('LABEL');
		expect(label.getAttribute('for')).toBe('chart-view-rotation');
		expect(label.className).toBe('ui-button ui-button-outline-primary ui-button-md');
	});

	it('adds the size and caller classes without dropping the variant', () => {
		render(
			<Button variant="outline-primary" size="sm" className="reforge-action">
				Optimize
			</Button>,
		);
		expect(screen.getByRole('button').className).toBe('ui-button ui-button-outline-primary ui-button-sm reforge-action');
	});

	it('applies no padding for size="none", leaving the caller to supply its own', () => {
		render(
			<Button variant="outline-primary" size="none" className="reforge-action">
				Optimize
			</Button>,
		);
		expect(screen.getByRole('button').className).toBe('ui-button ui-button-outline-primary reforge-action');
	});

	it('adds rel to a cross-origin link without being asked', () => {
		render(
			<>
				<Button href="https://www.wowhead.com/forever/item=1" target="_blank">
					Item
				</Button>
				<Button href="//discord.gg/p3DgvmnDCS">Discord</Button>
			</>,
		);
		screen.getAllByRole('link').forEach(link => expect(link.getAttribute('rel')).toBe('noopener noreferrer'));
	});

	it('merges an explicit rel rather than replacing it, and does not duplicate', () => {
		render(
			<Button href="https://www.wowhead.com/forever/item=1" rel="nofollow noopener">
				Item
			</Button>,
		);
		expect(screen.getByRole('link').getAttribute('rel')).toBe('nofollow noopener noreferrer');
	});

	// Relative links stay in this document, and the non-http schemes open no browsing context.
	it('leaves relative and non-http hrefs alone', () => {
		render(
			<>
				<Button href="/forever/warrior/arms/">Arms</Button>
				<Button href="#gear-tab">Gear</Button>
				<Button href="mailto:someone@example.com">Mail</Button>
				<Button href="javascript:void(0)">Inert</Button>
			</>,
		);
		screen.getAllByRole('link').forEach(link => expect(link.getAttribute('rel')).toBeNull());
	});

	it('emits the bare base bundle when variant is null, with no colour of its own', () => {
		render(
			<Button variant={null} className="talent-tree-reset link-danger text-link-danger">
				Reset
			</Button>,
		);
		expect(Array.from(screen.getByRole('button').classList).sort()).toEqual(
			['ui-button', 'ui-button-md', 'talent-tree-reset', 'link-danger', 'text-link-danger'].sort(),
		);
	});

	it('emits nothing of its own for the unstyled variant', () => {
		render(
			<Button variant="unstyled" className="delete-cooldown link-danger">
				Delete
			</Button>,
		);
		expect(screen.getByRole('button').className).toBe('delete-cooldown link-danger');
	});

	it('emits the .btn-reset look for link-danger', () => {
		render(<Button variant="link-danger">Reset</Button>);
		expect(screen.getByRole('button').className).toBe('ui-button-link ui-button-link-danger');
	});

	it('ignores size for a link variant', () => {
		render(
			<Button variant="link-danger" size="sm">
				Reset
			</Button>,
		);
		expect(screen.getByRole('button').className).toBe('ui-button-link ui-button-link-danger');
	});

	it('is chromeless for iconOnly, with no variant of its own', () => {
		render(
			<Button iconOnly aria-label="Close">
				×
			</Button>,
		);
		const button = screen.getByRole('button', { name: 'Close' });
		expect(button.className).toBe('ui-button-icon');
	});

	it('applies a colour variant to an iconOnly button, still chromeless', () => {
		render(
			<Button iconOnly variant="link-danger" aria-label="Delete">
				×
			</Button>,
		);
		expect(screen.getByRole('button', { name: 'Delete' }).className).toBe('ui-button-icon ui-button-link-danger');
	});

	it('keeps the link colour on an iconOnly link button', () => {
		render(
			<Button iconOnly variant="link" aria-label="Help">
				?
			</Button>,
		);
		expect(screen.getByRole('button', { name: 'Help' }).className).toBe('ui-button-icon ui-button-link');
	});

	it('renders a small spinner before the children and marks itself busy while loading', () => {
		render(
			<Button loading disabled>
				Stat Weights
			</Button>,
		);
		const button = screen.getByRole('button', { name: 'Stat Weights' });
		expect(button.getAttribute('aria-busy')).toBe('true');
		const spinner = within(button).getByTestId('loader');
		expect(button.firstElementChild).toBe(spinner);
		expect(spinner.getAttribute('data-size')).toBe('xs');
	});

	it('lays the spinner over an iconOnly button, keeping the icon in place but invisible, and marks itself busy', () => {
		render(
			<Button iconOnly loading aria-label="Settings">
				<Icon name="gear" testId="gear" />
			</Button>,
		);
		const button = screen.getByRole('button', { name: 'Settings' });
		expect(button.getAttribute('aria-busy')).toBe('true');
		expect(button.className).toBe('ui-button-icon');

		const icon = within(button).getByTestId('gear');
		expect(icon.parentElement!.parentElement).toBe(button);
		expect(icon.parentElement!.className).toBe('invisible');

		const spinner = within(button).getByTestId('loader');
		expect(spinner.previousElementSibling).toBe(icon.parentElement);
		expect(spinner.classList).toContain('ui-button-icon-spinner');
		expect(spinner.classList).not.toContain('ui-button-spinner');
	});

	it('renders no spinner and no aria-busy when not loading', () => {
		render(<Button>Simulate</Button>);
		const button = screen.getByRole('button', { name: 'Simulate' });
		expect(button.hasAttribute('aria-busy')).toBe(false);
		expect(within(button).queryByTestId('loader')).toBeNull();
	});
});

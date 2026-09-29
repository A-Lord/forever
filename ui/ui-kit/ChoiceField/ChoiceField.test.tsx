import { act, fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { type Ref, useImperativeHandle, useState } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ChoiceField } from './ChoiceField';
import type { ChoiceEntry, ChoiceFieldProps, ChoiceOption } from './types';

interface Unit {
	id: number;
	name: string;
}

const UNITS: Array<ChoiceOption<Unit>> = [
	{ value: { id: 0, name: 'All' }, label: 'All Targets' },
	{ value: { id: 1, name: 'One' }, label: 'Target 1', icon: 'bullseye', className: 'text-class-warrior' },
	{ value: { id: 2, name: 'Two' }, label: 'Target 2', icon: { iconUrl: 'https://example.test/two.jpg', name: 'Two' } },
];

const sameId = (a: Unit | undefined, b: Unit | undefined) => a?.id === b?.id;

const MODES: Array<ChoiceOption<number>> = [
	{ value: 0, label: 'Passive', tooltip: 'Do nothing' },
	{ value: 1, label: 'Aggressive', tooltip: '<p>Attack</p> everything' },
	{ value: 2, label: 'Defensive' },
];

interface HarnessHandle<V> {
	set: (next: V) => void;
}

type HarnessProps<V> = Omit<ChoiceFieldProps<V>, 'value' | 'onChange'> & {
	initial: V;
	accept?: (next: V, current: V) => V;
	handle?: Ref<HarnessHandle<V>>;
};

const picks = vi.fn();

// The parent owns the value; `handle.set` is a change landing from outside the field.
const Harness = <V,>({ initial, accept = next => next, handle, ...props }: HarnessProps<V>) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(handle, () => ({ set: setValue }), []);
	return (
		<ChoiceField<V>
			{...props}
			value={value}
			onChange={next => {
				picks(next);
				setValue(accept(next, value));
			}}
		/>
	);
};

const trigger = () => screen.getByRole('combobox');
const options = () => screen.queryAllByRole('option');
const open = () => userEvent.click(trigger());
const focused = () => document.activeElement!.textContent;

beforeEach(() => {
	picks.mockClear();
});

describe('ChoiceField', () => {
	describe('shell', () => {
		it('renders the bare trigger when it has no label, description or layout', () => {
			const { container } = render(<Harness id="mode" options={MODES} initial={0} className="w-full" testId="mode-field" />);

			// No wrapper: the trigger, and beside it the hidden input Base UI's Select keeps for forms.
			expect([...container.children].map(child => child.tagName)).toEqual(['BUTTON', 'INPUT']);
			expect(container.firstElementChild).toBe(trigger());
			expect(trigger().tagName).toBe('BUTTON');
			expect(trigger().id).toBe('mode');
			expect(trigger().getAttribute('data-testid')).toBe('mode-field');
			expect(trigger().getAttribute('class')).toContain('w-full');
		});

		it('renders one Field root around the trigger for a label, and the label names the trigger', () => {
			const { container } = render(
				<Harness id="mode" label="Pet Mode" description="Controls the pet" layout="inline" options={MODES} initial={0} testId="mode-field" />,
			);

			const root = container.firstElementChild!;
			expect(root.getAttribute('data-testid')).toBe('mode-field');
			expect(root.getAttribute('data-layout')).toBe('inline');
			expect(trigger().hasAttribute('data-testid')).toBe(false);
			expect(
				within(root as HTMLElement)
					.getByText('Pet Mode')
					.getAttribute('for'),
			).toBe('mode');
			expect(screen.getByRole('combobox', { name: 'Pet Mode' })).toBe(trigger());
			expect(within(root as HTMLElement).getByText('Controls the pet')).not.toBeNull();
		});

		it('names an unlabelled trigger through ariaLabel', () => {
			render(<Harness options={MODES} initial={0} ariaLabel="Stats type" />);

			expect(screen.getByRole('combobox', { name: 'Stats type' })).toBe(trigger());
		});

		it('renders the variant on the trigger, the bordered select by default', () => {
			const { unmount } = render(<Harness options={MODES} initial={0} />);
			expect(trigger().getAttribute('data-variant')).toBe('select');
			unmount();

			render(<Harness options={MODES} initial={0} variant="link" label="Mode" />);
			expect(trigger().getAttribute('data-variant')).toBe('link');
		});

		it('disables the trigger, and marks the root when there is one', () => {
			const bare = render(<Harness options={MODES} initial={0} disabled />);
			expect((trigger() as HTMLButtonElement).disabled).toBe(true);
			expect(bare.container.firstElementChild).toBe(trigger());
			bare.unmount();

			const labelled = render(<Harness options={MODES} initial={0} disabled label="Mode" />);
			expect((trigger() as HTMLButtonElement).disabled).toBe(true);
			expect(labelled.container.firstElementChild!.hasAttribute('data-disabled')).toBe(true);
		});
	});

	describe('value', () => {
		it("shows the selected option's icon and label, and carries its class onto the trigger", () => {
			render(<Harness options={UNITS} equals={sameId} initial={{ id: 1, name: 'One' }} variant="link" />);

			expect(trigger().textContent).toBe('Target 1');
			expect(within(trigger()).getByTestId('unit-picker-item-icon').tagName).toBe('svg');
			expect(trigger().getAttribute('class')).toContain('text-class-warrior');
		});

		it('shows the placeholder while no option matches the value', () => {
			render(<Harness options={UNITS} equals={sameId} initial={{ id: 9, name: 'Missing' }} placeholder="Unit" />);

			expect(trigger().textContent).toBe('Unit');
		});

		it('shows the first option of a select with no placeholder while nothing matches, and writes nothing', async () => {
			render(<Harness options={MODES} initial={7} />);

			expect(trigger().textContent).toBe('Passive');
			expect(picks).not.toHaveBeenCalled();
			await open();
			expect(options().map(option => option.getAttribute('aria-selected'))).toEqual(['false', 'false', 'false']);
		});

		it('shows what renderValue draws for the selection', () => {
			render(<Harness options={MODES} initial={1} renderValue={option => `[${String(option.label)}]`} />);

			expect(trigger().textContent).toBe('[Aggressive]');
		});

		it('reads the selection through `equals`, not by identity, and reports the option value itself', async () => {
			render(<Harness options={UNITS} equals={sameId} initial={{ id: 1, name: 'a different object' }} />);
			expect(trigger().textContent).toBe('Target 1');

			await open();
			expect(options().map(option => option.getAttribute('aria-selected'))).toEqual(['false', 'true', 'false']);

			await userEvent.click(options()[2]);
			expect(picks).toHaveBeenLastCalledWith(UNITS[2].value);
			expect(trigger().textContent).toBe('Target 2');
			expect(options()).toHaveLength(0);
		});

		it('treats an `undefined` value as an option like any other', async () => {
			const kinds: Array<ChoiceOption<string | undefined>> = [
				{ value: undefined, label: 'None' },
				{ value: 'and', label: 'All of' },
			];
			render(<Harness options={kinds} initial={'and' as string | undefined} />);

			await open();
			await userEvent.click(options()[0]);

			expect(picks).toHaveBeenLastCalledWith(undefined);
			expect(trigger().textContent).toBe('None');
		});

		it('shows a value the parent changes underneath it', () => {
			const handle = { current: null as HarnessHandle<number> | null };
			render(<Harness options={MODES} initial={0} handle={handle} />);

			act(() => handle.current!.set(2));

			expect(trigger().textContent).toBe('Defensive');
		});

		it('snaps back to the parent value when a pick leaves it where it was', async () => {
			render(<Harness options={MODES} initial={0} accept={(next, current) => (next === 2 ? current : next)} />);

			await open();
			await userEvent.click(options()[2]);

			expect(picks).toHaveBeenLastCalledWith(2);
			expect(trigger().textContent).toBe('Passive');
		});
	});

	describe('keyboard, select path', () => {
		it('opens from the trigger and reaches every option with the arrow keys', async () => {
			const user = userEvent.setup();
			render(<Harness options={MODES} initial={0} />);

			trigger().focus();
			await user.keyboard('{ArrowDown}');
			expect(options()).toHaveLength(3);

			const reached: Array<string | null> = [focused()];
			for (let step = 1; step < MODES.length; step++) {
				await user.keyboard('{ArrowDown}');
				reached.push(focused());
			}
			expect(reached).toEqual(['Passive', 'Aggressive', 'Defensive']);
			await user.keyboard('{ArrowUp}');
			expect(focused()).toBe('Aggressive');

			await user.keyboard('{Enter}');
			expect(picks).toHaveBeenLastCalledWith(1);
			expect(trigger().textContent).toBe('Aggressive');
			expect(options()).toHaveLength(0);
			expect(document.activeElement).toBe(trigger());
		});

		it('jumps with Home and End, finds an option by typing, and closes on Escape back onto the trigger without a pick', async () => {
			const user = userEvent.setup();
			render(<Harness options={MODES} initial={1} />);

			trigger().focus();
			await user.keyboard('{ArrowDown}');
			expect(focused()).toBe('Aggressive');
			await user.keyboard('{End}');
			expect(focused()).toBe('Defensive');
			await user.keyboard('{Home}');
			expect(focused()).toBe('Passive');
			await user.keyboard('d');
			expect(focused()).toBe('Defensive');

			await user.keyboard('{Escape}');
			expect(options()).toHaveLength(0);
			expect(document.activeElement).toBe(trigger());
			expect(picks).not.toHaveBeenCalled();
			expect(trigger().textContent).toBe('Aggressive');
		});
	});

	describe('per-option tooltips', () => {
		it('anchors only the options that carry one on the one tooltip the list shares', async () => {
			render(<Harness options={MODES} initial={0} />);
			await open();

			const [passive, aggressive, defensive] = options();
			expect(passive.getAttribute('data-tooltip-content')).toBe('Do nothing');
			expect(aggressive.getAttribute('data-tooltip-content')).toBe('<p>Attack</p> everything');
			expect(passive.getAttribute('data-tooltip-id')).toBe(aggressive.getAttribute('data-tooltip-id'));
			expect(defensive.hasAttribute('data-tooltip-id')).toBe(false);
		});

		it('anchors nothing when no option carries one', async () => {
			render(<Harness options={UNITS} equals={sameId} initial={UNITS[0].value} />);
			await open();

			expect(options().every(option => !option.hasAttribute('data-tooltip-id'))).toBe(true);
		});
	});

	// A picker sitting in an overflow-clipped drawer at the bottom of the page opens upwards and out of flow.
	describe('side and positionMethod', () => {
		const positioner = () => screen.getByTestId('choice-field-positioner');

		it('opens below the trigger, positioned in flow, by default', async () => {
			render(<Harness options={MODES} initial={0} />);
			await open();

			expect(positioner().getAttribute('data-side')).toBe('bottom');
			expect(positioner().style.position).toBe('absolute');
		});

		it('opens above the trigger and out of flow when asked', async () => {
			render(<Harness options={MODES} initial={0} side="top" positionMethod="fixed" />);
			await open();

			expect(positioner().getAttribute('data-side')).toBe('top');
			expect(positioner().style.position).toBe('fixed');
		});
	});

	describe('submenus', () => {
		const KINDS: Array<ChoiceEntry<string>> = [
			{ value: 'none', label: 'None' },
			{
				label: 'Logic »',
				submenu: [
					{ value: 'and', label: 'All of' },
					{ value: 'or', label: 'Any of', tooltip: 'Either side' },
				],
			},
			{ label: 'Resources »', submenu: [{ label: 'Mana »', submenu: [{ value: 'mana', label: 'Current Mana' }] }] },
		];

		const menuTrigger = () => screen.getByTestId('kind-field');
		const openMenu = () => act(() => void fireEvent.click(menuTrigger()));
		const rows = () => [...screen.getByRole('menu').querySelectorAll(':scope > ul > li')] as Array<HTMLElement>;

		it('opens as a menu with the submenus at the level they are filed at', async () => {
			render(<Harness options={KINDS} initial="mana" variant="link" testId="kind-field" />);
			expect(menuTrigger().textContent).toBe('Current Mana');
			expect(menuTrigger().getAttribute('data-variant')).toBe('link');

			await openMenu();

			expect(rows().map(row => row.textContent)).toEqual(['None', 'Logic »', 'Resources »']);
			expect(within(rows()[0]).getByRole('menuitemradio').getAttribute('aria-checked')).toBe('false');
		});

		it('renders a submenu once it is opened, and nests a two-level path', async () => {
			render(<Harness options={KINDS} initial="none" testId="kind-field" />);
			await openMenu();

			await act(() => void fireEvent.click(within(rows()[1]).getByRole('menuitem')));
			const logic = screen.getByTestId('choice-field-submenu');
			expect(
				within(logic)
					.getAllByRole('menuitemradio')
					.map(item => item.textContent),
			).toEqual(['All of', 'Any of']);
			expect(within(logic).getAllByRole('menuitemradio')[1].getAttribute('data-tooltip-content')).toBe('Either side');
		});

		it('reports the option picked inside a submenu', async () => {
			render(<Harness options={KINDS} initial="none" testId="kind-field" />);
			await openMenu();
			await act(() => void fireEvent.click(within(rows()[1]).getByRole('menuitem')));

			await act(() => void fireEvent.click(within(screen.getByTestId('choice-field-submenu')).getByText('Any of')));

			expect(picks).toHaveBeenLastCalledWith('or');
		});

		it('keeps an option that opens a submenu selectable', async () => {
			const units: Array<ChoiceEntry<string>> = [{ value: 'self', label: 'Self', submenu: [{ value: 'pet', label: 'Pet' }] }];
			render(<Harness options={units} initial="pet" testId="kind-field" />);
			await openMenu();

			const owner = within(rows()[0]).getByRole('menuitem');
			expect(owner.textContent).toBe('Self');
			await act(() => void fireEvent.click(owner));

			expect(picks).toHaveBeenLastCalledWith('self');
		});

		it('lets a keyboard user into a submenu with ArrowRight, back out with ArrowLeft, and pick inside it', async () => {
			const user = userEvent.setup();
			render(<Harness options={KINDS} initial="none" testId="kind-field" />);
			menuTrigger().focus();
			await user.keyboard('{ArrowDown}');
			expect(focused()).toBe('None');

			await user.keyboard('{ArrowDown}');
			expect(focused()).toBe('Logic »');
			await user.keyboard('{ArrowRight}');
			expect(focused()).toBe('All of');
			await user.keyboard('{ArrowDown}');
			expect(focused()).toBe('Any of');
			await user.keyboard('{ArrowUp}');
			expect(focused()).toBe('All of');

			await user.keyboard('{ArrowLeft}');
			expect(focused()).toBe('Logic »');
			expect(screen.queryByTestId('choice-field-submenu')).toBeNull();

			await user.keyboard('{ArrowRight}');
			await user.keyboard('{ArrowDown}');
			await user.keyboard('{Enter}');
			expect(picks).toHaveBeenLastCalledWith('or');
			expect(screen.queryByRole('menu')).toBeNull();
			expect(document.activeElement).toBe(menuTrigger());
		});

		it('jumps with Home and End, finds a row by typing, and closes on Escape back onto the trigger without a pick', async () => {
			const user = userEvent.setup();
			render(<Harness options={KINDS} initial="none" testId="kind-field" />);
			menuTrigger().focus();
			await user.keyboard('{ArrowDown}');

			await user.keyboard('{End}');
			expect(focused()).toBe('Resources »');
			await user.keyboard('{Home}');
			expect(focused()).toBe('None');
			await user.keyboard('l');
			expect(focused()).toBe('Logic »');

			await user.keyboard('{Escape}');
			expect(screen.queryByRole('menu')).toBeNull();
			expect(document.activeElement).toBe(menuTrigger());
			expect(picks).not.toHaveBeenCalled();
		});
	});
});

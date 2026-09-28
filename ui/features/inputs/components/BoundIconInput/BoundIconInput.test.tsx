import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { BoundIconInput } from './BoundIconInput';

vi.mock('../BoundIconPicker', () => ({
	BoundIconPicker: ({ testId }: { testId?: string }) => <div data-testid="bound-icon-picker" data-forwarded={testId} />,
}));
vi.mock('../BoundIconEnumPicker', () => ({
	BoundIconEnumPicker: ({ testId }: { testId?: string }) => <div data-testid="bound-icon-enum-picker" data-forwarded={testId} />,
}));

describe('BoundIconInput', () => {
	it('renders the icon picker for an icon config', () => {
		render(<BoundIconInput modObject={{}} config={{ type: 'icon' } as never} testId="buff" />);
		expect(screen.getByTestId('bound-icon-picker').getAttribute('data-forwarded')).toBe('buff');
		expect(screen.queryByTestId('bound-icon-enum-picker')).toBeNull();
	});

	it('renders the icon enum picker for an iconEnum config', () => {
		render(<BoundIconInput modObject={{}} config={{ type: 'iconEnum' } as never} />);
		expect(screen.getByTestId('bound-icon-enum-picker')).toBeTruthy();
		expect(screen.queryByTestId('bound-icon-picker')).toBeNull();
	});
});

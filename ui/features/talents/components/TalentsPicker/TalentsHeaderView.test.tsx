import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { TalentsHeaderView } from './TalentsHeaderView';

describe('TalentsHeaderView', () => {
	it('shows the points left to spend', () => {
		render(<TalentsHeaderView pointsRemaining={36} getTalentsString={() => ''} />);
		expect(document.querySelector('label')!.textContent!.match(/\d+$/)?.[0]).toBe('36');
	});
});

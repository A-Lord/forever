import type { ReactNode } from 'react';

export interface ReforgeFrozenSlotsViewProps {
	hidden: boolean;
	pickers: ReadonlyArray<ReactNode>;
}

export const ReforgeFrozenSlotsView = ({ hidden, pickers }: ReforgeFrozenSlotsViewProps) => (
	<div className="mb-2 grid grid-cols-2" data-testid="reforge-frozen-slots" hidden={hidden}>
		{pickers}
	</div>
);

import { SimHostProvider } from '@sim/context/SimHostContext';
import { fakeHost, type FakeHostParts } from '@sim/testing';
import { render, type RenderOptions } from '@testing-library/react';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { type ReactElement, type ReactNode, useState } from 'react';

// The portal container sits inside the rendered tree, as the shell's root element does, so an
// overlay opened by the component under test lands next to it.
const HostFrame = ({ parts, children }: { parts: FakeHostParts; children: ReactNode }) => {
	const [host] = useState(() => fakeHost(parts));
	const [portal, setPortal] = useState<HTMLDivElement | null>(null);
	return (
		<PortalContainerContext value={portal}>
			<SimHostProvider host={host}>{children}</SimHostProvider>
			<div ref={setPortal} data-testid="test-portal" />
		</PortalContainerContext>
	);
};

/** `render` inside `<SimHostProvider host={fakeHost(parts)}>` and a portal container. */
export const renderWithHost = (ui: ReactElement, parts: FakeHostParts = {}, options?: Omit<RenderOptions, 'wrapper'>) =>
	render(ui, { ...options, wrapper: ({ children }) => <HostFrame parts={parts}>{children}</HostFrame> });

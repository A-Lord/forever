import { useSimHost } from '@sim/context/SimHostContext';

import { AplListToolbarView, type AplListToolbarViewProps } from './AplListToolbarView';

export type AplListToolbarProps = Omit<AplListToolbarViewProps, 'onReset'>;

export const AplListToolbar = (props: AplListToolbarProps) => {
	const host = useSimHost();
	return <AplListToolbarView {...props} onReset={() => host.applyEmptyAplRotation()} />;
};

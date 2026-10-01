import i18n from '@i18n/config';
import { Button, ButtonProps } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import clsx from 'clsx';
import { useState } from 'react';
import { flushSync } from 'react-dom';

export interface AbortButtonProps extends Pick<ButtonProps, 'className'> {
	onAbort: (event: MouseEvent) => void;
	testId?: string;
}

export const AbortButton = ({ onAbort, className, testId }: AbortButtonProps) => {
	const [stopping, setStopping] = useState(false);
	return (
		<Button
			variant="unstyled"
			className={clsx('focus-visible:focus-ring', className)}
			disabled={stopping}
			data-testid={testId}
			onClick={event => {
				// Synchronous: the run action reads the button back in the click's own task.
				flushSync(() => setStopping(true));
				onAbort(event.nativeEvent);
			}}>
			<Icon name="xmark" style="base" size="lg" className="mr-1" />
			{stopping ? i18n.t('sidebar.results.stopping') : i18n.t('sidebar.results.stop')}
		</Button>
	);
};

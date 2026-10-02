import clsx from 'clsx';
import type { MouseEvent } from 'react';

export interface FakeLinkProps {
	onActivate: (event: MouseEvent<HTMLSpanElement>) => void;
	className?: string;
	testId?: string;
}

/**
 * A click layer over its positioned parent: placed first, it takes the clicks on the parent's empty
 * space and plain text, while a later sibling lifted with `relative z-1` keeps its own. It is hidden
 * from assistive tech and out of the tab order, so the parent's real control stays the one keyboard
 * and screen-reader path.
 */
export const FakeLink = ({ onActivate, className, testId = 'fake-link' }: FakeLinkProps) => (
	<span aria-hidden="true" className={clsx('absolute inset-0 cursor-pointer', className)} data-testid={testId} onClick={onActivate} />
);

import type { AnchorHTMLAttributes, MouseEvent, Ref } from 'react';

export interface WowheadAnchorProps extends Omit<AnchorHTMLAttributes<HTMLAnchorElement>, 'href'> {
	href?: string;
	/** `false` leaves `data-whtticon` off a linked anchor. */
	icon?: boolean;
	/**
	 * Cancels the click's navigation, then acts. It runs before `onClick`, which still receives the click. The anchor carries
	 * `data-disable-wowhead-touch-tooltip`, so a tap acts rather than opening wowhead's touch tooltip. Without an `href` it is a focusable
	 * `role="button"` that Enter and Space activate.
	 */
	onActivate?: (event: MouseEvent<HTMLAnchorElement>) => void;
	ref?: Ref<HTMLAnchorElement>;
	testId?: string;
}

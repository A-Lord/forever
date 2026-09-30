import type { AnchorHTMLAttributes, MouseEvent, Ref } from 'react';

export interface WowheadAnchorProps extends Omit<AnchorHTMLAttributes<HTMLAnchorElement>, 'href'> {
	href?: string;
	/** `false` leaves `data-whtticon` off a linked anchor. */
	icon?: boolean;
	/** `false` cancels the click and stops there: no `onActivate`, no role, no tab index, no touch opt-out. Default `true`. */
	navigate?: boolean;
	/**
	 * Cancels the click's navigation, then acts. It runs before `onClick`, which still receives the click. The anchor carries
	 * `data-disable-wowhead-touch-tooltip`, so a tap acts rather than opening wowhead's touch tooltip. Without an `href` it is a focusable
	 * `role="button"` that Enter and Space activate.
	 */
	onActivate?: (event: MouseEvent<HTMLAnchorElement>) => void;
	/** Marks the anchor `aria-disabled`, keeps it out of the tab order where it would otherwise be a button, and never calls `onActivate`. The click is still cancelled. */
	disabled?: boolean;
	ref?: Ref<HTMLAnchorElement>;
	testId?: string;
}

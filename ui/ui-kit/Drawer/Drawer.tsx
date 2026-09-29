import { Drawer as BaseDrawer } from '@base-ui/react/drawer';
import { usePortalContainer } from '@ui-kit/hooks/usePortalContainer';
import clsx from 'clsx';
import type { CSSProperties, ReactElement, ReactNode } from 'react';

export interface DrawerProps {
	trigger: ReactElement;
	open?: boolean;
	onOpenChange?: (open: boolean) => void;
	ignoreOutsidePress?: (event: Event) => boolean;
	modal?: boolean;
	title?: ReactNode;
	/** Content beside the title, inside the header bar. */
	headerChildren?: ReactNode;
	/** On the popup, which carries position and fade but no paint of its own: the caller's class places and paints it. */
	className?: string;
	/** Positions the portalled popup, since it does not share a containing block with its anchor. */
	style?: CSSProperties;
	children?: ReactNode;
	testId?: string;
}

export const Drawer = ({
	trigger,
	open,
	onOpenChange,
	ignoreOutsidePress,
	modal = true,
	title,
	headerChildren,
	className,
	style,
	children,
	testId,
}: DrawerProps) => {
	const portalContainer = usePortalContainer();
	return (
		<BaseDrawer.Root
			open={open}
			modal={modal}
			onOpenChange={(nextOpen, eventDetails) => {
				if (!nextOpen && eventDetails.reason === 'outside-press' && ignoreOutsidePress?.(eventDetails.event)) {
					eventDetails.cancel();
					return;
				}
				onOpenChange?.(nextOpen);
			}}>
			<BaseDrawer.Trigger render={trigger} />
			<BaseDrawer.Portal container={portalContainer ?? undefined}>
				<BaseDrawer.Viewport className="contents">
					<BaseDrawer.Popup className={clsx('ui-drawer', className)} style={style} data-testid={testId ?? 'sim-drawer-popup'}>
						{(title != null || headerChildren != null) && (
							<div className="ui-drawer-header" data-testid="sim-drawer-header">
								{title != null && (
									<BaseDrawer.Title className="m-0 text-lg leading-normal" data-testid="sim-drawer-title">
										{title}
									</BaseDrawer.Title>
								)}
								{headerChildren}
							</div>
						)}
						{children}
					</BaseDrawer.Popup>
				</BaseDrawer.Viewport>
			</BaseDrawer.Portal>
		</BaseDrawer.Root>
	);
};

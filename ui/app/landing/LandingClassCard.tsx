import { Class } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translatePlayerClass, translatePlayerSpec, translateStatus } from '@i18n/localization';
import { LaunchStatus } from '@sim/constants/other';
import { PlayerClasses } from '@sim/player/classes/index';
import type { PlayerClass } from '@sim/player/player_class';
import type { PlayerSpec } from '@sim/player/player_spec';
import { Button } from '@ui-kit/Button';
import type { ReactNode } from 'react';

import { classLaunchStatus } from './landing_classes';

export interface LandingClassCardProps {
	playerClass: PlayerClass<Class>;
}

// A launched sim needs no badge; every other status is shown on the card.
const statusBadge = (status: LaunchStatus): string | undefined => (status === LaunchStatus.Launched ? undefined : translateStatus(status));

interface CardContentProps {
	iconPath: string;
	title: string;
	status?: string;
	children?: ReactNode;
}

const CardContent = ({ iconPath, title, status, children }: CardContentProps) => (
	<>
		<span className="ui-landing-class-card-cover" aria-hidden="true" />
		<span className="ui-landing-class-card-body">
			{status && (
				<span className="ui-landing-class-card-badge" data-testid="launch-status-label">
					<span className="ui-landing-class-card-badge-dot" aria-hidden="true" />
					{status}
				</span>
			)}
			<span className="flex flex-col items-start gap-2">
				<img className="ui-landing-class-card-logo" src={iconPath} alt="" />
				<span className="ui-landing-class-card-title" data-testid="class-card-title">
					{title}
				</span>
			</span>
			{children}
		</span>
		<span className="ui-landing-class-card-sheen" aria-hidden="true" />
	</>
);

export const LandingClassCard = ({ playerClass }: LandingClassCardProps) => {
	const className = translatePlayerClass(playerClass);
	const scheme = PlayerClasses.getCssScheme(playerClass);
	const specs = Object.values(playerClass.specs) as Array<PlayerSpec<any>>;
	// Priest's class-level icon (`getIcon`) is a spell icon (spell_shadow_shadowwordpain), not the
	// generic class_*.jpg every other class returns, so the landing page hardcodes class_priest.jpg here.
	const classIconPath =
		playerClass.classID === Class.ClassPriest ? 'https://wow.zamimg.com/images/wow/icons/large/class_priest.jpg' : playerClass.getIcon('large');

	// A class with one sim for all its specs gets a single "All Specs" link, under the class icon.
	const links =
		specs.length === 1
			? [{ href: specs[0].simLink, iconPath: classIconPath, label: i18n.t('landing.home.allSpecs') }]
			: specs.map(spec => ({ href: spec.simLink, iconPath: spec.getIcon('large'), label: translatePlayerSpec(spec) }));

	return (
		<div className="ui-landing-class-card" data-class={scheme} data-testid="class-card">
			<CardContent iconPath={classIconPath} title={className} status={statusBadge(classLaunchStatus(playerClass))}>
				<span className="flex flex-wrap gap-2">
					{links.map(link => (
						<Button key={link.href} href={link.href} variant="unstyled" className="ui-landing-spec-pill" data-testid="spec-link">
							<img className="size-6 rounded-full" src={link.iconPath} alt="" />
							{link.label}
						</Button>
					))}
				</span>
			</CardContent>
		</div>
	);
};

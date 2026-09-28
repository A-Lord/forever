import { Unstyled } from '@storybook/addon-docs/blocks';
import clsx from 'clsx';

import { SPEC_THEMES, type SpecTheme } from '../tokens/spec_themes';

const SpecThemeCard = ({ theme }: { theme: SpecTheme }) => (
	<div className={clsx('sim-ui', theme.className, 'flex min-h-40 flex-col gap-2 bg-sim bg-cover bg-center p-3 text-foreground')}>
		<strong>{theme.name}</strong>
		<span className="self-start bg-primary px-2 py-1 text-primary-foreground">bg-primary text-primary-foreground</span>
		<code className="text-ui">.{theme.className}</code>
		<code className="mt-auto truncate text-ui text-link" title={theme.backgroundImage}>
			{theme.backgroundImage}
		</code>
	</div>
);

export const SpecThemeCards = () => (
	<Unstyled>
		<div className="grid grid-cols-auto-fit-220 gap-3">
			{SPEC_THEMES.map(theme => (
				<SpecThemeCard key={theme.key} theme={theme} />
			))}
		</div>
	</Unstyled>
);

import { getLang, setLang, supportedLanguages } from '@i18n/locale_service';
import { Icon } from '@ui-kit/Icon';
import { Menu, MenuItem } from '@ui-kit/Menu';

export const LandingLanguageMenu = () => {
	const currentLang = getLang();

	const selectLang = (lang: string) => {
		setLang(lang);
		window.location.reload();
	};

	return (
		<div className="relative">
			<Menu
				surface="none"
				trigger={<Icon name="globe" size="2x" />}
				triggerProps={{
					id: 'languageDropdown',
					className: 'ui-landing-language-caret flex items-center px-0 py-4 text-sm whitespace-nowrap text-white/55 max-md:pt-2 md:px-2',
					'aria-label': supportedLanguages[currentLang],
				}}
				align="end"
				sideOffset={0}
				keepMounted
				className="ui-landing-language-popup">
				{Object.entries(supportedLanguages).map(([code, name]) => (
					<MenuItem
						key={code}
						layout="block"
						className="clear-both rounded-none no-underline"
						data-active={code === currentLang ? '' : undefined}
						data-lang={code}
						data-testid="dropdown-item"
						onClick={() => selectLang(code)}>
						{name}
					</MenuItem>
				))}
			</Menu>
		</div>
	);
};

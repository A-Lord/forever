import type { APLValidation } from '@generated/proto/api';
import { LogLevel } from '@generated/proto/common';
import i18n from '@i18n/config';
import { ListItemAction } from '@ui-kit/ListPicker';
import { NOTICE_COLOR, NOTICE_ICON_NAME, type NoticeLevel } from '@ui-kit/Notice';
import { Tooltip } from '@ui-kit/Tooltip';
import { Fragment, useId } from 'react';

export interface AplValidationsViewProps {
	/** The item's validations, every spell id already resolved to its name. Empty hides the button. */
	validations: ReadonlyArray<APLValidation>;
}

const LEVEL = new Map<LogLevel, NoticeLevel>([
	[LogLevel.Information, 'info'],
	[LogLevel.Warning, 'warning'],
	[LogLevel.Error, 'error'],
]);

const HEADER = new Map<LogLevel, string>([
	[LogLevel.Information, i18n.t('common.list_picker.additional_information')],
	[LogLevel.Warning, i18n.t('common.list_picker.action_has_warnings')],
	[LogLevel.Error, i18n.t('common.list_picker.action_has_errors')],
]);

/** Written out rather than derived: `LogLevel` is a numeric enum, so iterating it yields the reverse mappings too. */
const LEVEL_CLASS = new Map<LogLevel, string>([
	[LogLevel.Undefined, 'apl-validation-undefined'],
	[LogLevel.Information, `apl-validation-information ${NOTICE_COLOR.info}`],
	[LogLevel.Warning, `apl-validation-warning ${NOTICE_COLOR.warning}`],
	[LogLevel.Error, `apl-validation-error ${NOTICE_COLOR.error}`],
]);

const groupByLevel = (validations: ReadonlyArray<APLValidation>) => {
	const grouped = new Map<LogLevel, Array<string>>();
	let maxLogLevel = LogLevel.Undefined;
	for (const entry of validations) {
		maxLogLevel = Math.max(entry.logLevel, maxLogLevel);
		const messages = grouped.get(entry.logLevel);
		if (messages) messages.push(entry.validation);
		else grouped.set(entry.logLevel, [entry.validation]);
	}
	return { maxLogLevel, groups: Array.from(grouped, ([logLevel, messages]) => ({ header: HEADER.get(logLevel), messages })) };
};

/**
 * The warning triangle on an APL list item's header, and the grouped list its tooltip shows.
 *
 * The button is present but `display: none` while there is nothing to say, which keeps the header's
 * element count stable. A message is text, never markup: the sim formats rotation-supplied group and
 * variable names straight into it, so it reaches the tooltip as a text child and not as HTML.
 */
export const AplValidationsView = ({ validations }: AplValidationsViewProps) => {
	const tooltipId = useId();
	const shown = validations.length > 0;
	const { maxLogLevel, groups } = groupByLevel(validations);
	const icon = NOTICE_ICON_NAME[(shown && LEVEL.get(maxLogLevel)) || 'warning'];

	return (
		<>
			<ListItemAction
				icon={icon}
				className={shown ? LEVEL_CLASS.get(maxLogLevel) : undefined}
				testId="apl-validations"
				hidden={!shown}
				tooltip={i18n.t('common.list_picker.warnings')}
				tooltipId={tooltipId}
			/>
			<Tooltip
				id={tooltipId}
				// `render` wins over the anchor's own `data-tooltip-content`, which is what the fallback branch hands back.
				render={({ content }) =>
					shown
						? groups.map((entry, index) => (
								<Fragment key={index}>
									<p>{entry.header}</p>
									<ul>
										{entry.messages.map((message, messageIndex) => (
											<li key={messageIndex}>{message}</li>
										))}
									</ul>
								</Fragment>
							))
						: content
				}
			/>
		</>
	);
};

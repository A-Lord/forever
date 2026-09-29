import { AplNameDialog } from '@features/apl/components/AplNameDialog';
import { AplValidations } from '@features/apl/components/ListItemHeader';
import { uuidValidations } from '@features/apl/components/ListItemHeader/utils';
import { ValuePicker } from '@features/apl/components/ValuePicker';
import { useVariableExtraction } from '@features/apl/hooks/useVariableExtraction';
import { type ListItemsConfig, useListItems } from '@features/inputs/hooks/useListItems';
import type { InputConfig } from '@features/inputs/input';
import { keyFor } from '@features/inputs/list_ids';
import { APLValue } from '@generated/proto/apl';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { randomUUID } from '@sim/utils/misc';
import { Tree, type TreeSortable } from '@ui-kit/Tree';
import clsx from 'clsx';
import { useMemo, useRef } from 'react';

export interface ValueListFieldProps {
	player: Player<any>;
	config: InputConfig<Player<any>, Array<APLValue | undefined>> & { id: string };
}

const newValue = () => APLValue.create({ uuid: { value: randomUUID() } });

/**
 * The operand list of `and`, `or`, `min` and `max`.
 *
 * Each row is a whole `ValuePicker`, so this is one of the two edges the value tree recurses
 * across. The row's own validations ride in the item title, and a row's id is its value's uuid.
 */
export const ValueListField = ({ player, config }: ValueListFieldProps) => {
	const values = () => config.getValue(player) || [];
	const extraction = useVariableExtraction(
		player,
		index => values()[index],
		(index, reference) => {
			const list = values();
			list[index] = reference;
			config.setValue(player, list);
		},
	);

	const listConfig: ListItemsConfig<Player<any>, APLValue | undefined> = {
		...config,
		// An empty row is stored as a value with a uuid, never as a hole.
		setValue: (subject: Player<any>, next: Array<APLValue | undefined>) =>
			config.setValue(
				subject,
				next.map(value => value || newValue()),
			),
		newItem: newValue,
		copyItem: (oldValue: APLValue | undefined) => (oldValue ? APLValue.clone(oldValue) : oldValue),
	};
	const { items, itemConfig, create, remove, copy, move, hidden, disabled } = useListItems(player, listConfig);
	const moveRef = useRef(move);
	moveRef.current = move;
	const sortable = useMemo((): TreeSortable => ({ onMove: ({ from, to }) => moveRef.current(from.index, to.index) }), []);
	if (hidden) return null;

	const itemLabel = 'Value';
	const { extraAction } = extraction;
	return (
		<>
			<Tree.Root
				group="apl-values"
				layout="inline"
				density="compact"
				sortable={sortable}
				disabled={disabled}
				className={clsx(config.extraClassNames)}
				testId="apl-value-list">
				<Tree.Items>
					{items.map((value, index) => {
						const key = String(keyFor(value, index));
						const valueConfig = itemConfig(value, index);
						return (
							<Tree.Item key={key} id={value?.uuid?.value || key} testId="apl-value-list-item">
								<Tree.Item.Header>
									<Tree.Item.Handle label={i18n.t('common.list_picker.move_drag_drop')} testId="apl-value-list-handle" />
									<Tree.Item.Title>
										<AplValidations getValidations={subject => uuidValidations(subject, valueConfig.getValue(subject)?.uuid?.value)} />
									</Tree.Item.Title>
									<Tree.Item.Actions label={i18n.t('common.list_picker.item_actions', { itemLabel })}>
										<Tree.Item.Action
											label={i18n.t('common.list_picker.delete_item', { itemLabel })}
											icon="times"
											tone="danger"
											testId="apl-value-list-delete"
											onClick={() => remove(index)}
										/>
										{(extraAction.shouldShow?.(index) ?? true) && (
											<Tree.Item.Action
												label={extraAction.tooltip}
												icon={extraAction.icon}
												testId="apl-value-list-extract"
												onClick={() => extraAction.onClick(index)}
											/>
										)}
										<Tree.Item.Action
											label={i18n.t('common.list_picker.copy_to_new', { itemLabel })}
											icon="copy"
											testId="apl-value-list-copy"
											onClick={() => copy(index)}
										/>
									</Tree.Item.Actions>
								</Tree.Item.Header>
								<Tree.Item.Body>
									<ValuePicker player={player} config={valueConfig} />
								</Tree.Item.Body>
							</Tree.Item>
						);
					})}
				</Tree.Items>
				<Tree.AddButton label={i18n.t('common.list_picker.new_item', { itemLabel })} icon="plus" onClick={create} />
			</Tree.Root>
			<AplNameDialog {...extraction.dialog} />
		</>
	);
};

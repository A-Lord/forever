import { ActionPicker } from '@features/apl/components/ActionPicker';
import { useApl } from '@features/apl/context/AplContext';
import { freshUuids } from '@features/apl/model/fresh_uuids';
import { type ListItemsConfig, useListItems } from '@features/inputs/hooks/useListItems';
import type { InputConfig } from '@features/inputs/input';
import { keyFor } from '@features/inputs/list_ids';
import { APLAction } from '@generated/proto/apl';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { Tree, type TreeSortable } from '@ui-kit/Tree';
import clsx from 'clsx';
import { useMemo, useRef } from 'react';

export interface ActionListFieldProps {
	player: Player<any>;
	config: InputConfig<Player<any>, Array<APLAction>> & { id: string };
}

/**
 * The action list of a sequence or strict sequence — the action tree's own recursion edge.
 *
 * Its rows sit in the group the enclosing `ActionPicker` names after its sequence (`actionListGroup`),
 * which is not the priority list's, so a sequence's rows never mix with a priority list's. A row's
 * id is its list key: an `APLAction` carries no uuid.
 */
export const ActionListField = ({ player, config }: ActionListFieldProps) => {
	const listConfig: ListItemsConfig<Player<any>, APLAction> = {
		...config,
		newItem: APLAction.create,
		// Called through, not passed: `clone` reads `this`, so the bare method copies nothing.
		copyItem: (oldItem: APLAction) => freshUuids(APLAction.clone(oldItem)),
	};
	const group = useApl().listGroup ?? 'apl-actions';
	const { items, itemConfig, create, remove, copy, move, hidden, disabled } = useListItems(player, listConfig);
	const latest = useRef({ move, group });
	latest.current = { move, group };
	// Outside an `AplList` root this list owns the drag, and the lists nested in its rows report to it too.
	const sortable = useMemo(
		(): TreeSortable => ({
			onMove: ({ from, to }) => {
				if (from.group === latest.current.group && to.group === latest.current.group) latest.current.move(from.index, to.index);
			},
		}),
		[],
	);
	if (hidden) return null;

	const itemLabel = i18n.t('rotation_tab.apl.priorityList.name');
	return (
		<Tree.Root
			group={group}
			layout="inline"
			density="compact"
			sortable={sortable}
			disabled={disabled}
			className={clsx(config.extraClassNames)}
			testId="apl-action-list">
			<Tree.Items>
				{items.map((action, index) => {
					const key = String(keyFor(action, index));
					return (
						<Tree.Item key={key} id={key} testId="apl-action-list-item">
							<Tree.Item.Header>
								<Tree.Item.Handle label={i18n.t('common.list_picker.move_drag_drop')} testId="apl-action-list-handle" />
								<Tree.Item.Actions label={i18n.t('common.list_picker.item_actions', { itemLabel })}>
									<Tree.Item.Action
										label={i18n.t('common.list_picker.delete_item', { itemLabel })}
										icon="times"
										tone="danger"
										testId="apl-action-list-delete"
										onClick={() => remove(index)}
									/>
									<Tree.Item.Action
										label={i18n.t('common.list_picker.copy_to_new', { itemLabel })}
										icon="copy"
										testId="apl-action-list-copy"
										onClick={() => copy(index)}
									/>
								</Tree.Item.Actions>
							</Tree.Item.Header>
							<Tree.Item.Body>
								<ActionPicker player={player} config={itemConfig(action, index)} />
							</Tree.Item.Body>
						</Tree.Item>
					);
				})}
			</Tree.Items>
			<Tree.AddButton label={i18n.t('common.list_picker.new_item', { itemLabel })} icon="plus" onClick={create} />
		</Tree.Root>
	);
};

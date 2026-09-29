import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { AplNameDialog, type AplNameDialogProps } from './AplNameDialog';

const AplNameDialogStory = ({ open: initialOpen, onClose, ...props }: AplNameDialogProps) => {
	const [open, setOpen] = useState(initialOpen);
	return (
		<>
			<Button onClick={() => setOpen(true)}>{props.title}</Button>
			<AplNameDialog
				{...props}
				open={open}
				onClose={() => {
					setOpen(false);
					onClose();
				}}
			/>
		</>
	);
};

const meta = {
	title: 'Features/Apl/AplNameDialog',
	component: AplNameDialog,
	tags: ['autodocs'],
	args: {
		open: false,
		title: 'New Group',
		inputLabel: 'Name',
		existingNames: ['Execute', 'Opener'],
		onSubmit: fn(),
		onCancel: fn(),
		onClose: fn(),
	},
	render: args => <AplNameDialogStory {...args} />,
} satisfies Meta<typeof AplNameDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Creating: the confirm button waits for a name. */
export const Create: Story = {};

/** A taken name blocks the confirm button and says why. */
export const NameConflict: Story = { args: { defaultValue: 'Execute' } };

export const Rename: Story = {
	args: { title: 'Rename Group', confirmLabel: 'Rename', defaultValue: 'Opener', existingNames: ['Execute'] },
};

/** Extracting a condition to a new value variable. */
export const Extract: Story = {
	args: { title: 'Extract to Variable', confirmLabel: 'Extract', placeholder: 'rage_pooled', existingNames: [] },
};

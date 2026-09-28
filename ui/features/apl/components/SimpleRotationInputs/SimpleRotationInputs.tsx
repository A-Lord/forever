import { BoundInput } from '@features/inputs/components/BoundInput';
import { useSpecConfig } from '@sim/context/SimHostContext';

import { RotationIconGroup } from './RotationIconGroup';

/**
 * The simple-rotation block: the spec's icon row, then its inputs.
 *
 * `inline` on `BoundInput` lays each input out on a copy of its config, so the spec object is
 * never written.
 */
export const SimpleRotationInputs = () => {
	const config = useSpecConfig();

	return (
		<>
			<RotationIconGroup inputs={config.rotationIconInputs ?? []} />
			{(config.rotationInputs?.inputs ?? []).map((input, index) => (
				<BoundInput key={index} config={input} inline />
			))}
		</>
	);
};

import { createContext, useContext } from 'react';

/** Opens the Batch Sim tab. Unset where the panel sits in that tab already, or the page has none. */
export const ReforgeOpenBatchSimContext = createContext<(() => void) | undefined>(undefined);

export const useReforgeOpenBatchSim = () => useContext(ReforgeOpenBatchSimContext);

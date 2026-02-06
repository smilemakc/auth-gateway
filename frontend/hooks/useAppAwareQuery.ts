import { useApplication } from '../services/appContext';

export function useCurrentAppId(): string | null {
  const { currentApplicationId } = useApplication();
  return currentApplicationId;
}

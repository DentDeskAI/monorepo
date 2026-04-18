import { useAuth } from './useAuth'

/**
 * Returns the current tenant's clinic_id from the auth store.
 * Use this whenever building tenant-scoped API queries.
 */
export function useTenant() {
  const user = useAuth((s) => s.user)
  return {
    clinicId: user?.clinic_id ?? null,
    isReady: !!user?.clinic_id,
  }
}

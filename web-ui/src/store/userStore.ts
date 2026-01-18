import { create } from 'zustand'
import { getCurrentUser } from '../api/client'

interface User {
  user_id: string
  org_id: string
  user_name: string
  email: string
  avatar_url: string
  role: string
  user_type: string
  is_admin: boolean
  permissions: string[]
  balance: number
}

interface UserStore {
  user: User | null
  loading: boolean
  error: string | null
  fetchUser: () => Promise<void>
  clearUser: () => void
}

export const useUserStore = create<UserStore>((set) => ({
  user: null,
  loading: true,
  error: null,
  fetchUser: async () => {
    set({ loading: true, error: null })
    try {
      const data = await getCurrentUser()
      if (data && data.user_id) {
        set({ user: data, loading: false })
      } else {
        set({ user: null, loading: false, error: 'Invalid user data' })
      }
    } catch {
      set({ user: null, loading: false, error: 'Not authenticated' })
    }
  },
  clearUser: () => set({ user: null, loading: false, error: null }),
}))

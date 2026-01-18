import { getCurrentUser } from './client'

export interface UserInfo {
  org_id: string
  user_id: string
  org_name: string
  email: string
  user_name: string
  avatar_url: string
  role: string
  user_type?: 'admin' | 'regular'
  permissions: string[]
}

export const authApi = {
  getCurrentUser: (): Promise<UserInfo> => getCurrentUser(),
}

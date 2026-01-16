import { UserOutlined } from '@ant-design/icons'
import { useUserStore } from '../store/userStore'

interface HeaderProps {
  title?: string
}

export default function Header({ title = 'Dashboard' }: HeaderProps) {
  const { user, loading } = useUserStore()

  return (
    <header className="header">
      <h1 className="page-title">{title}</h1>
      <div className="user-info">
        {loading ? (
          <span style={{ color: 'var(--text-secondary)' }}>Loading...</span>
        ) : user ? (
          <>
            <div className="user-avatar">
              {user.avatar_url ? (
                <img src={user.avatar_url} alt={user.user_name} style={{ width: 32, height: 32, borderRadius: '50%' }} />
              ) : (
                <UserOutlined />
              )}
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end' }}>
              <span>{user.user_name || user.email}</span>
              <span style={{ fontSize: 12, color: 'var(--success)' }}>${user.balance?.toFixed(2) || '0.00'}</span>
            </div>
          </>
        ) : (
          <a href="https://wavespeed.ai/login" style={{ color: 'var(--primary)' }}>Login</a>
        )}
      </div>
    </header>
  )
}

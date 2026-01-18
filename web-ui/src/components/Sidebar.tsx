import { NavLink, useLocation } from 'react-router-dom'
import {
  RocketOutlined,
  CloudServerOutlined,
  UnorderedListOutlined,
  WalletOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  CloudOutlined,
  ClusterOutlined,
  AppstoreOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import { useUserStore } from '../store/userStore'

interface SidebarProps {
  collapsed: boolean
  onCollapse: (collapsed: boolean) => void
}

const menuItems = [
  { path: '/serverless', icon: <RocketOutlined />, label: 'Serverless', adminOnly: false },
  { path: '/endpoints', icon: <CloudServerOutlined />, label: 'Endpoints', adminOnly: false },
  { path: '/tasks', icon: <UnorderedListOutlined />, label: 'Tasks', adminOnly: false },
  { path: '/billing', icon: <WalletOutlined />, label: 'Billing', adminOnly: false },
  { path: '/settings', icon: <SettingOutlined />, label: 'Settings', adminOnly: false },
  { path: '/clusters', icon: <ClusterOutlined />, label: 'Clusters', adminOnly: true },
  { path: '/specs', icon: <AppstoreOutlined />, label: 'Specs', adminOnly: true },
]

export default function Sidebar({ collapsed, onCollapse }: SidebarProps) {
  const location = useLocation()
  const user = useUserStore((state) => state.user)
  const isAdmin = user?.is_admin || false

  const visibleItems = menuItems.filter(item => !item.adminOnly || isAdmin)

  return (
    <div className={`sidebar ${collapsed ? 'collapsed' : ''}`}>
      <div className="logo">
        <CloudOutlined style={{ fontSize: 20, color: '#1da1f2' }} />
        {!collapsed && <span>Portal</span>}
      </div>
      <nav className="nav-menu">
        {visibleItems.map((item) => (
          <NavLink
            key={item.path}
            to={item.path}
            className={({ isActive }) =>
              `nav-item ${isActive || location.pathname.startsWith(item.path) ? 'active' : ''}`
            }
          >
            {item.icon}
            {!collapsed && <span>{item.label}</span>}
          </NavLink>
        ))}
      </nav>
      <div
        className="nav-item"
        onClick={() => onCollapse(!collapsed)}
        style={{ position: 'absolute', bottom: 12, width: '100%' }}
      >
        {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
        {!collapsed && <span>Collapse</span>}
      </div>
    </div>
  )
}

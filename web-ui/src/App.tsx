import { useState, useEffect } from 'react'
import { Routes, Route, Navigate, useLocation } from 'react-router-dom'
import Sidebar from './components/Sidebar'
import Header from './components/Header'
import Serverless from './pages/Serverless'
import Endpoints from './pages/Endpoints'
import EndpointDetail from './pages/EndpointDetail'
import Tasks from './pages/Tasks'
import Billing from './pages/Billing'
import Clusters from './pages/Clusters'
import ClusterDetail from './pages/ClusterDetail'
import SpecsAdmin from './pages/SpecsAdmin'
import Settings from './pages/Settings'
import Login from './pages/Login'
import { useUserStore } from './store/userStore'
import { ThemeProvider } from './hooks/useTheme'
import './index.css'

const pageTitles: Record<string, string> = {
  '/serverless': 'Serverless',
  '/endpoints': 'Endpoints',
  '/tasks': 'Tasks',
  '/billing': 'Billing',
  '/clusters': 'Clusters',
  '/specs': 'Specs',
  '/settings': 'Settings',
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useUserStore()
  const location = useLocation()

  if (loading) {
    return <div className="loading"><div className="spinner"></div></div>
  }

  if (!user) {
    sessionStorage.setItem('redirectAfterLogin', location.pathname)
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

function AppLayout() {
  const location = useLocation()
  const [collapsed, setCollapsed] = useState(false)

  const getPageTitle = () => {
    for (const [path, title] of Object.entries(pageTitles)) {
      if (location.pathname.startsWith(path)) return title
    }
    return 'Portal'
  }

  return (
    <div className="app">
      <Sidebar collapsed={collapsed} onCollapse={setCollapsed} />
      <div className="main-content" style={{ marginLeft: collapsed ? 64 : 200 }}>
        <Header title={getPageTitle()} />
        <div className="content">
          <Routes>
            <Route path="/serverless" element={<Serverless />} />
            <Route path="/endpoints" element={<Endpoints />} />
            <Route path="/endpoints/:name" element={<EndpointDetail />} />
            <Route path="/tasks" element={<Tasks />} />
            <Route path="/billing" element={<Billing />} />
            <Route path="/clusters" element={<Clusters />} />
            <Route path="/clusters/:id" element={<ClusterDetail />} />
            <Route path="/specs" element={<SpecsAdmin />} />
            <Route path="/settings" element={<Settings />} />
            <Route path="*" element={<Navigate to="/endpoints" replace />} />
          </Routes>
        </div>
      </div>
    </div>
  )
}

function App() {
  const fetchUser = useUserStore((state) => state.fetchUser)

  useEffect(() => {
    fetchUser()
  }, [])

  return (
    <ThemeProvider>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/*" element={
          <ProtectedRoute>
            <AppLayout />
          </ProtectedRoute>
        } />
        <Route path="/" element={<Navigate to="/endpoints" replace />} />
      </Routes>
    </ThemeProvider>
  )
}

export default App

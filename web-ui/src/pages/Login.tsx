import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { CloudOutlined, LoginOutlined } from '@ant-design/icons'
import { useUserStore } from '../store/userStore'

export default function Login() {
  const navigate = useNavigate()
  const user = useUserStore((state) => state.user)

  // 登录成功后跳转
  useEffect(() => {
    if (user) {
      const redirectPath = sessionStorage.getItem('redirectAfterLogin') || '/dashboard'
      sessionStorage.removeItem('redirectAfterLogin')
      navigate(redirectPath, { replace: true })
    }
  }, [user, navigate])

  const handleLogin = () => {
    const isDev = import.meta.env.DEV
    const mainSiteURL = isDev ? 'https://tropical.wavespeed.ai' : 'https://wavespeed.ai'
    window.location.href = mainSiteURL + '/sign-in?redirect=' + encodeURIComponent(window.location.href)
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <div className="login-logo">
          <CloudOutlined style={{ fontSize: 48, color: '#1da1f2' }} />
        </div>
        <h1>Portal</h1>
        <p>Sign in to manage your serverless endpoints</p>
        <button className="btn btn-primary login-btn" onClick={handleLogin}>
          <LoginOutlined /> Sign in with WaveSpeed
        </button>
        <div className="login-footer">
          © {new Date().getFullYear()} WaveSpeedAI
        </div>
      </div>
    </div>
  )
}

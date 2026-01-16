import { useEffect, useState } from 'react'
import { CloudServerOutlined, DollarOutlined, ThunderboltOutlined, ClockCircleOutlined } from '@ant-design/icons'
import { getEndpoints, getBalance, getUsage } from '../api/client'

export default function Dashboard() {
  const [data, setData] = useState({ endpoints: 0, balance: 0, totalCost: 0, totalGPUHours: 0 })
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [endpointsRes, balanceRes, usageRes] = await Promise.all([
          getEndpoints().catch(() => ({ endpoints: [] })),
          getBalance().catch(() => ({ balance: 0 })),
          getUsage().catch(() => ({ total_amount: 0, total_gpu_hours: 0 })),
        ])
        setData({
          endpoints: endpointsRes.endpoints?.length || 0,
          balance: balanceRes.balance || 0,
          totalCost: usageRes.total_amount || 0,
          totalGPUHours: usageRes.total_gpu_hours || 0,
        })
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [])

  if (loading) return <div className="loading"><div className="spinner"></div></div>

  return (
    <div>
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-label"><CloudServerOutlined /> Active Endpoints</div>
          <div className="stat-value">{data.endpoints}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label"><DollarOutlined /> Balance</div>
          <div className="stat-value green">${data.balance.toFixed(2)}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label"><ClockCircleOutlined /> Total Cost (30d)</div>
          <div className="stat-value">${data.totalCost.toFixed(2)}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label"><ThunderboltOutlined /> GPU Hours (30d)</div>
          <div className="stat-value purple">{data.totalGPUHours.toFixed(1)}</div>
        </div>
      </div>

      <div className="card">
        <div className="card-header"><h3>Quick Start</h3></div>
        <div className="card-body">
          <p style={{ color: 'var(--text-secondary)', marginBottom: 16 }}>
            Deploy your first serverless endpoint in minutes.
          </p>
          <a href="/endpoints" className="btn btn-primary">Create Endpoint</a>
        </div>
      </div>
    </div>
  )
}

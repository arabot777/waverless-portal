import { useEffect, useState } from 'react'
import { DollarOutlined, ThunderboltOutlined, CreditCardOutlined, ClockCircleOutlined } from '@ant-design/icons'
import { getBalance, getUsage, getWorkerRecords } from '../api/client'
import dayjs from 'dayjs'

interface BillingRecord {
  id: number
  worker_id: string
  gpu_type: string
  gpu_count: number
  duration_seconds: number
  amount: number
  status: string
  created_at: string
}

export default function Billing() {
  const [balance, setBalance] = useState({ balance: 0, credit_limit: 0 })
  const [usage, setUsage] = useState({ total_amount: 0, total_gpu_hours: 0 })
  const [records, setRecords] = useState<BillingRecord[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [b, u, r] = await Promise.all([
          getBalance().catch(() => ({ balance: 0, credit_limit: 0 })),
          getUsage().catch(() => ({ total_amount: 0, total_gpu_hours: 0 })),
          getWorkerRecords(20, 0).catch(() => ({ records: [] })),
        ])
        setBalance(b)
        setUsage(u)
        setRecords(r.records || [])
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [])

  if (loading) return <div className="loading"><div className="spinner"></div></div>

  return (
    <div>
      <div className="stats-grid mb-5">
        <div className="stat-card">
          <div className="stat-label"><DollarOutlined /> Balance</div>
          <div className="stat-value green">${balance.balance.toFixed(2)}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label"><CreditCardOutlined /> Credit Limit</div>
          <div className="stat-value">${balance.credit_limit.toFixed(2)}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label"><ClockCircleOutlined /> Total Cost (30d)</div>
          <div className="stat-value">${usage.total_amount.toFixed(2)}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label"><ThunderboltOutlined /> GPU Hours (30d)</div>
          <div className="stat-value purple">{usage.total_gpu_hours.toFixed(1)}</div>
        </div>
      </div>

      <div className="card">
        <div className="card-header"><h3>Recent Transactions</h3></div>
        <table>
          <thead>
            <tr>
              <th>Time</th>
              <th>Worker</th>
              <th>GPU</th>
              <th>Duration</th>
              <th>Amount</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {records.length === 0 ? (
              <tr><td colSpan={6} className="empty-state">No transactions yet</td></tr>
            ) : (
              records.map(r => (
                <tr key={r.id}>
                  <td>{dayjs(r.created_at).format('YYYY-MM-DD HH:mm')}</td>
                  <td style={{ maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis' }}>{r.worker_id}</td>
                  <td>{r.gpu_type} x{r.gpu_count}</td>
                  <td>{Math.round(r.duration_seconds / 60)} min</td>
                  <td style={{ color: 'var(--danger)' }}>-${r.amount?.toFixed(4)}</td>
                  <td><span className={`tag ${r.status}`}>{r.status}</span></td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}

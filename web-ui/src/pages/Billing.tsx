import { useEffect, useState } from 'react'
import { DollarOutlined, ClockCircleOutlined, InfoCircleOutlined } from '@ant-design/icons'
import { Select } from 'antd'
import { getUsage, getWorkerRecords } from '../api/client'
import dayjs from 'dayjs'

interface BillingRecord {
  id: number
  worker_id: string
  endpoint_name: string
  spec_name: string
  duration_seconds: number
  amount: number
  status: string
  created_at: string
}

export default function Billing() {
  const [usage, setUsage] = useState({ total_amount: 0, total_seconds: 0 })
  const [records, setRecords] = useState<BillingRecord[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(20)
  const [loading, setLoading] = useState(true)

  const fetchRecords = async () => {
    try {
      const r = await getWorkerRecords(pageSize, page * pageSize)
      setRecords(r.records || [])
      setTotal(r.total || 0)
    } catch {}
  }

  useEffect(() => {
    const fetchData = async () => {
      try {
        const u = await getUsage().catch(() => ({ total_amount: 0, total_seconds: 0 }))
        setUsage(u)
        await fetchRecords()
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [])

  useEffect(() => {
    if (!loading) fetchRecords()
  }, [page, pageSize])

  if (loading) return <div className="loading"><div className="spinner"></div></div>

  const totalHours = (usage.total_seconds / 3600).toFixed(1)
  const totalPages = Math.ceil(total / pageSize) || 1

  return (
    <div>
      <div className="stats-grid mb-5" style={{ gridTemplateColumns: 'repeat(2, 1fr)' }}>
        <div className="stat-card">
          <div className="stat-label"><DollarOutlined /> Total Cost (7d)</div>
          <div className="stat-value">${usage.total_amount.toFixed(2)}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label"><ClockCircleOutlined /> Usage Hours (7d)</div>
          <div className="stat-value purple">{totalHours}</div>
        </div>
      </div>

      <div className="card">
        <div className="card-header">
          <h3>Recent Transactions</h3>
          <span style={{ fontSize: 12, color: 'var(--warning)', display: 'flex', alignItems: 'center', gap: 4 }}>
            <InfoCircleOutlined /> Only last 7 days data available
          </span>
        </div>
        <table>
          <thead>
            <tr>
              <th>Time</th>
              <th>Endpoint</th>
              <th>Spec</th>
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
                  <td>{r.endpoint_name || '-'}</td>
                  <td>{r.spec_name || '-'}</td>
                  <td>{Math.round(r.duration_seconds / 60)} min</td>
                  <td style={{ color: 'var(--danger)' }}>-${r.amount?.toFixed(4)}</td>
                  <td><span className={`tag ${r.status}`}>{r.status}</span></td>
                </tr>
              ))
            )}
          </tbody>
        </table>
        {/* Pagination */}
        <div className="flex justify-between items-center" style={{ padding: '12px 16px', borderTop: '1px solid var(--border-color)' }}>
          <div className="flex gap-2 items-center">
            <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>Show</span>
            <Select value={pageSize} onChange={v => { setPageSize(v); setPage(0) }} size="small" style={{ width: 70 }} options={[
              { value: 10, label: '10' }, { value: 20, label: '20' }, { value: 50, label: '50' },
            ]} />
            <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>/ page • Total: {total}</span>
          </div>
          <div className="flex gap-2 items-center">
            <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>Page {page + 1} of {totalPages}</span>
            <button className="btn btn-sm btn-outline" disabled={page === 0} onClick={() => setPage(p => p - 1)}>Prev</button>
            <button className="btn btn-sm btn-outline" disabled={page + 1 >= totalPages} onClick={() => setPage(p => p + 1)}>Next</button>
          </div>
        </div>
      </div>
    </div>
  )
}

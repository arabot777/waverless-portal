import { useState, useEffect } from 'react'
import { Input, Select, Tooltip, Modal, Popconfirm, message } from 'antd'
import { SearchOutlined, EyeOutlined, StopOutlined } from '@ant-design/icons'
import { getAllTasks, getTaskStatus, getEndpoints } from '../api/client'
import api from '../api/client'

export default function Tasks() {
  const [searchInput, setSearchInput] = useState('')
  const [statusInput, setStatusInput] = useState('all')
  const [endpointInput, setEndpointInput] = useState('all')
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')
  const [endpointFilter, setEndpointFilter] = useState('all')
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(20)
  const [tasks, setTasks] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [endpoints, setEndpoints] = useState<any[]>([])
  const [selectedTask, setSelectedTask] = useState<any>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [fullTask, setFullTask] = useState<any>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [overview, setOverview] = useState<any>({})

  useEffect(() => {
    getEndpoints().then(data => setEndpoints(data?.endpoints || []))
    api.get('/api/v1/tasks/overview').then(r => setOverview(r.data || {}))
  }, [])

  const fetchTasks = () => {
    setLoading(true)
    const params: any = { limit: pageSize, offset: page * pageSize }
    if (statusFilter !== 'all') params.status = statusFilter
    if (endpointFilter !== 'all') params.endpoint = endpointFilter
    if (search) params.task_id = search
    getAllTasks(params)
      .then(data => { setTasks(data?.tasks || []); setTotal(data?.total || 0) })
      .catch(() => { setTasks([]); setTotal(0) })
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchTasks() }, [statusFilter, endpointFilter, search, page, pageSize])

  const handleSearch = () => { setSearch(searchInput); setStatusFilter(statusInput); setEndpointFilter(endpointInput); setPage(0) }
  const handleReset = () => { setSearchInput(''); setStatusInput('all'); setEndpointInput('all'); setSearch(''); setStatusFilter('all'); setEndpointFilter('all'); setPage(0) }

  const cancelTask = async (taskId: string) => {
    try {
      await api.post(`/v1/cancel/${taskId}`)
      message.success('Task cancelled')
      fetchTasks()
    } catch (e: any) {
      message.error(e.response?.data?.error || 'Failed to cancel')
    }
  }

  const openDetail = (task: any) => {
    setSelectedTask(task)
    setDetailOpen(true)
    setDetailLoading(true)
    getTaskStatus(task.id)
      .then(data => setFullTask(data))
      .catch(() => setFullTask(null))
      .finally(() => setDetailLoading(false))
  }

  const getStatusClass = (s: string) => s === 'COMPLETED' ? 'success' : s === 'FAILED' ? 'failed' : s === 'IN_PROGRESS' ? 'running' : 'pending'

  return (
    <div>
      <h2 style={{ marginBottom: 20 }}>Tasks</h2>

      {/* Stats */}
      <div className="stats-grid mb-4" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
        <div className="stat-card">
          <div className="stat-value" style={{ color: '#48bb78' }}>{overview?.completed || 0}</div>
          <div className="stat-label">Completed</div>
        </div>
        <div className="stat-card">
          <div className="stat-value" style={{ color: '#1da1f2' }}>{overview?.in_progress || 0}</div>
          <div className="stat-label">In Progress</div>
        </div>
        <div className="stat-card">
          <div className="stat-value" style={{ color: '#f59e0b' }}>{overview?.pending || 0}</div>
          <div className="stat-label">Pending</div>
        </div>
        <div className="stat-card">
          <div className="stat-value" style={{ color: '#f56565' }}>{overview?.failed || 0}</div>
          <div className="stat-label">Failed</div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex justify-between items-center mb-4">
        <div className="flex gap-2">
          <Input placeholder="Search task ID..." prefix={<SearchOutlined />} value={searchInput} onChange={e => setSearchInput(e.target.value)} onPressEnter={handleSearch} style={{ width: 240 }} />
          <Select value={statusInput} onChange={setStatusInput} style={{ width: 140 }} options={[
            { value: 'all', label: 'All Status' },
            { value: 'PENDING', label: 'Pending' },
            { value: 'IN_PROGRESS', label: 'In Progress' },
            { value: 'COMPLETED', label: 'Completed' },
            { value: 'FAILED', label: 'Failed' },
          ]} />
          <Select value={endpointInput} onChange={setEndpointInput} style={{ width: 180 }} options={[
            { value: 'all', label: 'All Endpoints' },
            ...(endpoints?.map((ep: any) => ({ value: ep.name, label: ep.name })) || []),
          ]} />
          <button className="btn btn-primary" onClick={handleSearch}><SearchOutlined /> Search</button>
          <button className="btn btn-outline" onClick={handleReset}>Reset</button>
        </div>
        <span style={{ color: 'var(--text-secondary)', fontSize: 13 }}>Total: {total}</span>
      </div>

      {/* Table */}
      <div className="card">
        <div className="table-container">
          <table>
            <thead><tr><th>Task ID</th><th>Status</th><th>Endpoint</th><th>Worker</th><th>Created</th><th>Exec Time</th><th>Actions</th></tr></thead>
            <tbody>
              {loading ? <tr><td colSpan={7}><div className="loading"><div className="spinner"></div></div></td></tr> :
               tasks.length === 0 ? <tr><td colSpan={7} style={{ textAlign: 'center', color: 'var(--text-secondary)', padding: 40 }}>No tasks</td></tr> :
               tasks.map((t: any) => (
                 <tr key={t.id}>
                   <td><Tooltip title={t.id}><span style={{ fontFamily: 'monospace', fontSize: 11 }}>{t.id?.substring(0, 20)}...</span></Tooltip></td>
                   <td><span className={`tag ${getStatusClass(t.status)}`}>{t.status}</span></td>
                   <td style={{ fontSize: 12 }}>{t.endpoint || '-'}</td>
                   <td>{t.workerId ? <Tooltip title={t.workerId}><span style={{ fontSize: 12, color: 'var(--text-secondary)' }}>{t.workerId.substring(0, 12)}...</span></Tooltip> : '-'}</td>
                   <td style={{ fontSize: 12 }}>{t.createdAt ? new Date(t.createdAt).toLocaleString() : '-'}</td>
                   <td style={{ fontSize: 12 }}>{t.executionTime ? `${(t.executionTime/1000).toFixed(2)}s` : '-'}</td>
                   <td>
                     <div className="flex gap-2">
                       <button className="btn btn-sm btn-outline" onClick={() => openDetail(t)}><EyeOutlined /></button>
                       {(t.status === 'PENDING' || t.status === 'IN_PROGRESS') && (
                         <Popconfirm title="Cancel this task?" onConfirm={() => cancelTask(t.id)}>
                           <button className="btn btn-sm btn-outline" style={{ color: '#f56565' }}><StopOutlined /></button>
                         </Popconfirm>
                       )}
                     </div>
                   </td>
                 </tr>
               ))}
            </tbody>
          </table>
        </div>
        {/* Pagination */}
        <div className="flex justify-between items-center" style={{ padding: '12px 16px', borderTop: '1px solid var(--border)' }}>
          <div className="flex gap-2 items-center">
            <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>Show</span>
            <Select value={pageSize} onChange={v => { setPageSize(v); setPage(0) }} size="small" style={{ width: 70 }} options={[
              { value: 10, label: '10' }, { value: 20, label: '20' }, { value: 50, label: '50' }, { value: 100, label: '100' },
            ]} />
            <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>/ page</span>
          </div>
          <div className="flex gap-2 items-center">
            <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>Page {page + 1} of {Math.ceil(total / pageSize) || 1}</span>
            <button className="btn btn-sm btn-outline" disabled={page === 0} onClick={() => setPage(p => p - 1)}>Prev</button>
            <button className="btn btn-sm btn-outline" disabled={(page + 1) * pageSize >= total} onClick={() => setPage(p => p + 1)}>Next</button>
          </div>
        </div>
      </div>

      {/* Detail Modal */}
      <Modal title="Task Details" open={detailOpen} onCancel={() => { setDetailOpen(false); setFullTask(null) }} footer={null} width={800}>
        {detailLoading ? <div className="loading"><div className="spinner"></div></div> : (fullTask || selectedTask) && (() => {
          const t = fullTask || selectedTask
          return (
            <div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
                <div><label className="form-label">Task ID</label><div style={{ fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-all' }}>{t.id}</div></div>
                <div><label className="form-label">Status</label><span className={`tag ${getStatusClass(t.status)}`}>{t.status}</span></div>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginTop: 16 }}>
                <div><label className="form-label">Endpoint</label><div>{t.endpoint || '-'}</div></div>
                <div><label className="form-label">Worker</label><div style={{ fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-all' }}>{t.workerId || '-'}</div></div>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginTop: 16 }}>
                <div><label className="form-label">Created</label><div>{t.createdAt ? new Date(t.createdAt).toLocaleString() : '-'}</div></div>
                <div><label className="form-label">Execution Time</label><div>{t.executionTime ? `${(t.executionTime/1000).toFixed(2)}s` : '-'}</div></div>
              </div>
              <div style={{ marginTop: 16 }}>
                <label className="form-label">Input</label>
                <pre style={{ background: '#1e1e1e', color: '#d4d4d4', padding: 12, borderRadius: 6, fontSize: 12, overflow: 'auto', maxHeight: 200, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{t.input ? JSON.stringify(t.input, null, 2) : '-'}</pre>
              </div>
              {t.output && (
                <div style={{ marginTop: 16 }}>
                  <label className="form-label">Output</label>
                  <pre style={{ background: '#1e1e1e', color: '#d4d4d4', padding: 12, borderRadius: 6, fontSize: 12, overflow: 'auto', maxHeight: 300, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{JSON.stringify(t.output, null, 2)}</pre>
                </div>
              )}
              {t.error && (
                <div style={{ marginTop: 16 }}>
                  <label className="form-label">Error</label>
                  <pre style={{ background: '#2d1f1f', color: '#f87171', padding: 12, borderRadius: 6, fontSize: 12, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{t.error}</pre>
                </div>
              )}
            </div>
          )
        })()}
      </Modal>
    </div>
  )
}

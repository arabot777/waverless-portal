import { useState, useEffect } from 'react'
import { Input, Select, Tooltip, Modal, Popconfirm, message } from 'antd'
import { SearchOutlined, EyeOutlined, StopOutlined, CheckCircleOutlined, ClockCircleOutlined, SyncOutlined, CloseCircleOutlined, InfoCircleOutlined } from '@ant-design/icons'
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
      api.get('/api/v1/tasks/overview').then(r => setOverview(r.data || {}))
    } catch (e: any) {
      message.error(e.response?.data?.error || 'Failed to cancel')
    }
  }

  const openDetail = (task: any) => {
    setSelectedTask(task)
    setDetailOpen(true)
    setDetailLoading(true)
    const taskId = task.task_id || task.id
    getTaskStatus(taskId)
      .then(data => setFullTask(data))
      .catch(() => setFullTask(null))
      .finally(() => setDetailLoading(false))
  }

  const getStatusClass = (s: string) => s === 'COMPLETED' ? 'success' : s === 'FAILED' ? 'failed' : s === 'IN_PROGRESS' ? 'running' : 'pending'
  const getStatusIcon = (s: string) => {
    switch (s) {
      case 'COMPLETED': return <CheckCircleOutlined style={{ color: '#48bb78' }} />
      case 'IN_PROGRESS': return <SyncOutlined spin style={{ color: '#1da1f2' }} />
      case 'PENDING': return <ClockCircleOutlined style={{ color: '#f59e0b' }} />
      case 'FAILED': return <CloseCircleOutlined style={{ color: '#f56565' }} />
      default: return null
    }
  }

  // 兼容两种字段格式
  const getTaskId = (t: any) => t.task_id || t.id
  const getWorkerId = (t: any) => t.worker_id || t.workerId
  const getCreatedAt = (t: any) => t.submitted_at || t.createdAt
  const getExecTime = (t: any) => t.execution_time_ms || t.executionTime

  return (
    <div>
      <div className="flex justify-between items-center" style={{ marginBottom: 20 }}>
        <h2 style={{ margin: 0 }}>Tasks</h2>
        <span style={{ fontSize: 12, color: 'var(--warning)', display: 'flex', alignItems: 'center', gap: 4 }}>
          <InfoCircleOutlined /> Only last 7 days data available
        </span>
      </div>

      {/* Stats */}
      <div className="stats-grid mb-4" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
        <div className="stat-card">
          <div className="stat-icon" style={{ background: 'rgba(72,187,120,0.1)', color: '#48bb78' }}><CheckCircleOutlined /></div>
          <div className="stat-content"><div className="stat-label">Completed</div><div className="stat-value">{overview?.completed || 0}</div></div>
        </div>
        <div className="stat-card">
          <div className="stat-icon" style={{ background: 'rgba(29,161,242,0.1)', color: '#1da1f2' }}><SyncOutlined /></div>
          <div className="stat-content"><div className="stat-label">In Progress</div><div className="stat-value">{overview?.in_progress || 0}</div></div>
        </div>
        <div className="stat-card">
          <div className="stat-icon" style={{ background: 'rgba(245,158,11,0.1)', color: '#f59e0b' }}><ClockCircleOutlined /></div>
          <div className="stat-content"><div className="stat-label">Pending</div><div className="stat-value">{overview?.pending || 0}</div></div>
        </div>
        <div className="stat-card">
          <div className="stat-icon" style={{ background: 'rgba(245,101,101,0.1)', color: '#f56565' }}><CloseCircleOutlined /></div>
          <div className="stat-content"><div className="stat-label">Failed</div><div className="stat-value">{overview?.failed || 0}</div></div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex justify-between items-center mb-4">
        <div className="filters">
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
            <thead><tr><th>Task ID</th><th>Endpoint</th><th>Status</th><th>Worker</th><th>Created</th><th>Exec Time</th><th>Actions</th></tr></thead>
            <tbody>
              {loading ? <tr><td colSpan={7}><div className="loading"><div className="spinner"></div></div></td></tr> :
               tasks.length === 0 ? <tr><td colSpan={7} style={{ textAlign: 'center', color: 'var(--text-secondary)', padding: 40 }}>No tasks found</td></tr> :
               tasks.map((t: any) => (
                 <tr key={t.id || t.task_id}>
                   <td><Tooltip title={getTaskId(t)}><span style={{ fontFamily: 'monospace', fontSize: 11 }}>{getTaskId(t)?.substring(0, 20)}...</span></Tooltip></td>
                   <td style={{ fontSize: 12 }}>{t.endpoint || '-'}</td>
                   <td><span className={`tag ${getStatusClass(t.status)}`}>{getStatusIcon(t.status)} {t.status}</span></td>
                   <td>{getWorkerId(t) ? <Tooltip title={getWorkerId(t)}><span style={{ fontSize: 12, color: 'var(--text-secondary)' }}>{getWorkerId(t).substring(0, 12)}...</span></Tooltip> : '-'}</td>
                   <td style={{ fontSize: 12 }}>{getCreatedAt(t) ? new Date(getCreatedAt(t)).toLocaleString() : '-'}</td>
                   <td style={{ fontSize: 12 }}>{getExecTime(t) ? `${(getExecTime(t)/1000).toFixed(2)}s` : '-'}</td>
                   <td>
                     <div className="flex gap-2">
                       <button className="btn btn-sm btn-outline" onClick={() => openDetail(t)}><EyeOutlined /></button>
                       {(t.status === 'PENDING' || t.status === 'IN_PROGRESS') && (
                         <Popconfirm title="Cancel this task?" onConfirm={() => cancelTask(getTaskId(t))}>
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
        {detailLoading ? <div className="loading"><div className="spinner"></div></div> : (() => {
          const t = fullTask || selectedTask
          if (!t) return null
          return (
            <div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
                <div><label className="form-label">Task ID</label><div style={{ fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-all' }}>{getTaskId(t)}</div></div>
                <div><label className="form-label">Status</label><span className={`tag ${getStatusClass(t.status)}`}>{getStatusIcon(t.status)} {t.status}</span></div>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginTop: 16 }}>
                <div><label className="form-label">Endpoint</label><div>{t.endpoint || '-'}</div></div>
                <div><label className="form-label">Worker</label><div style={{ fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-all' }}>{getWorkerId(t) || '-'}</div></div>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginTop: 16 }}>
                <div><label className="form-label">Created</label><div>{getCreatedAt(t) ? new Date(getCreatedAt(t)).toLocaleString() : '-'}</div></div>
                <div><label className="form-label">Execution Time</label><div>{getExecTime(t) ? `${(getExecTime(t)/1000).toFixed(2)}s` : '-'}</div></div>
              </div>
              <div style={{ marginTop: 16 }}>
                <label className="form-label">Input</label>
                <pre style={{ background: '#1e1e1e', color: '#d4d4d4', padding: 12, borderRadius: 6, fontSize: 12, overflow: 'auto', maxHeight: 200, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{t.input ? (typeof t.input === 'string' ? t.input : JSON.stringify(t.input, null, 2)) : '-'}</pre>
              </div>
              {t.output && (
                <div style={{ marginTop: 16 }}>
                  <label className="form-label">Output</label>
                  <pre style={{ background: '#1e1e1e', color: '#d4d4d4', padding: 12, borderRadius: 6, fontSize: 12, overflow: 'auto', maxHeight: 300, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{typeof t.output === 'string' ? t.output : JSON.stringify(t.output, null, 2)}</pre>
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

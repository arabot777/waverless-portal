import { useEffect, useState, useRef } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { message, Popconfirm, Modal, Form, Input, InputNumber, Collapse, Drawer, Tabs, Select, Tooltip } from 'antd'
import { ArrowLeftOutlined, DeleteOutlined, ReloadOutlined, EditOutlined, PlayCircleOutlined, CopyOutlined, PlusOutlined, SyncOutlined, SearchOutlined, EyeOutlined } from '@ant-design/icons'
import * as echarts from 'echarts'
import { getEndpoint, getEndpointWorkers, getWorkerLogs, getWorkerTasks, getEndpointMetrics, getEndpointStats, updateEndpoint, updateEndpointConfig, deleteEndpoint, submitTask, getTaskStatus, getEndpointTasks } from '../api/client'
import Terminal from '../components/Terminal'

type TabKey = 'overview' | 'metrics' | 'workers' | 'tasks' | 'settings'

interface EndpointData {
  // Portal 字段
  logical_name?: string
  physical_name?: string
  cluster_id?: string
  price_per_hour?: number
  // Waverless 字段
  name?: string
  specName?: string
  spec_type?: string
  gpu_type?: string
  gpu_count?: number
  cpu_cores?: number
  ram_gb?: number
  status?: string
  replicas?: number
  minReplicas?: number
  maxReplicas?: number
  readyReplicas?: number
  image?: string
  taskTimeout?: number
  createdAt?: string
  shmSize?: string
}

interface Worker {
  id: string
  pod_name: string
  status: string
  phase: string
  current_jobs: number
  concurrency: number
  tasks_completed: number
  last_heartbeat: string
  last_task_time: string
}

interface RealtimeMetrics {
  workers?: { active: number; idle: number; total: number }
  tasks?: { pending: number; running: number; completed_last_minute: number }
  performance?: { avg_execution_ms: number; avg_queue_wait_ms: number }
}

export default function EndpointDetail() {
  const { name } = useParams<{ name: string }>()
  const navigate = useNavigate()
  const [endpoint, setEndpoint] = useState<EndpointData | null>(null)
  const [workers, setWorkers] = useState<Worker[]>([])
  const [metrics, setMetrics] = useState<RealtimeMetrics | null>(null)
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<TabKey>('overview')
  const [editModalOpen, setEditModalOpen] = useState(false)
  const [form] = Form.useForm()

  const fetchData = async () => {
    if (!name) return
    try {
      const [ep, wk, mt] = await Promise.all([
        getEndpoint(name),
        getEndpointWorkers(name).catch(() => []),
        getEndpointMetrics(name).catch(() => null),
      ])
      setEndpoint(ep)
      setWorkers(Array.isArray(wk) ? wk : wk?.workers || [])
      setMetrics(mt)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 5000)
    return () => clearInterval(interval)
  }, [name])

  const handleDelete = async () => {
    try {
      await deleteEndpoint(name!)
      message.success('Endpoint deleted')
      navigate('/endpoints')
    } catch {
      message.error('Failed to delete endpoint')
    }
  }

  const handleUpdate = async () => {
    try {
      const values = await form.validateFields()
      await updateEndpoint(name!, values)
      message.success('Updated')
      setEditModalOpen(false)
      fetchData()
    } catch {
      message.error('Failed to update')
    }
  }

  const openEdit = () => {
    form.setFieldsValue({ replicas: workers.length || 1, image: endpoint?.image })
    setEditModalOpen(true)
  }

  if (loading) return <div className="loading"><div className="spinner"></div></div>
  if (!endpoint) return <div className="empty-state"><p>Endpoint not found</p><Link to="/endpoints" className="btn btn-outline mt-4">Back</Link></div>

  const activeWorkers = workers.filter(w => (w.current_jobs || 0) > 0).length
  const pendingTasks = metrics?.tasks?.pending || 0
  const runningTasks = metrics?.tasks?.running || 0

  return (
    <>
      {/* Header */}
      <div className="flex justify-between items-center mb-4">
        <div className="flex items-center gap-3">
          <Link to="/endpoints" className="btn btn-outline btn-icon"><ArrowLeftOutlined /></Link>
          <h2 style={{ margin: 0, fontSize: 20, fontWeight: 600 }}>{name}</h2>
          <span className={`tag ${endpoint.status?.toLowerCase()}`}>{endpoint.status}</span>
        </div>
        <div className="flex gap-2">
          <button className="btn btn-outline" onClick={() => fetchData()}><ReloadOutlined /> Refresh</button>
          <button className="btn btn-outline" onClick={openEdit}><EditOutlined /> Edit</button>
          <Popconfirm title="Delete this endpoint?" onConfirm={handleDelete} okText="Yes" cancelText="No">
            <button className="btn btn-outline" style={{ color: 'var(--danger)', borderColor: 'var(--danger)' }}><DeleteOutlined /> Delete</button>
          </Popconfirm>
        </div>
      </div>

      {/* Meta */}
      <div style={{ fontSize: 13, color: 'var(--text-secondary)', marginBottom: 12 }}>
        {endpoint.specName} • {endpoint.cluster_id || '-'} • ${(endpoint.price_per_hour || 0).toFixed(2)}/hr
      </div>

      {/* Stats */}
      <div className="stats-grid mb-4" style={{ gridTemplateColumns: 'repeat(5, 1fr)' }}>
        <div className="stat-card"><div className="stat-label">Workers</div><div className="stat-value"><span style={{ fontSize: 28 }}>{activeWorkers}</span><span style={{ fontSize: 14, color: 'var(--text-secondary)' }}> / {workers.length}</span></div></div>
        <div className="stat-card"><div className="stat-label">Replicas</div><div className="stat-value" style={{ color: 'var(--success)' }}><span style={{ fontSize: 28 }}>{endpoint.readyReplicas || 0}</span><span style={{ fontSize: 14, color: 'var(--text-secondary)' }}> / {endpoint.replicas || 0}</span></div></div>
        <div className="stat-card"><div className="stat-label">Running</div><div className="stat-value" style={{ color: 'var(--primary)', fontSize: 28 }}>{runningTasks}</div></div>
        <div className="stat-card"><div className="stat-label">Pending</div><div className="stat-value" style={{ color: 'var(--warning)', fontSize: 28 }}>{pendingTasks}</div></div>
        <div className="stat-card"><div className="stat-label">Cost/hr</div><div className="stat-value" style={{ color: 'var(--warning)', fontSize: 28 }}>${(workers.length * (endpoint.price_per_hour || 0)).toFixed(2)}</div></div>
      </div>

      {/* Tabs */}
      <div className="tabs mb-4">
        {(['overview', 'metrics', 'workers', 'tasks', 'settings'] as TabKey[]).map(t => (
          <div key={t} className={`tab ${activeTab === t ? 'active' : ''}`} onClick={() => setActiveTab(t)}>
            {t.charAt(0).toUpperCase() + t.slice(1)}
          </div>
        ))}
      </div>

      {/* Tab Content */}
      {activeTab === 'overview' && <OverviewTab endpoint={endpoint} workers={workers} />}
      {activeTab === 'metrics' && <MetricsTab name={name!} />}
      {activeTab === 'workers' && <WorkersTab workers={workers} endpointName={name!} />}
      {activeTab === 'tasks' && <TasksTab endpointName={name!} />}
      {activeTab === 'settings' && <SettingsTab endpoint={endpoint} onDelete={handleDelete} onRefresh={fetchData} />}

      {/* Edit Modal */}
      <Modal title="Quick Edit" open={editModalOpen} onOk={handleUpdate} onCancel={() => setEditModalOpen(false)}>
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="replicas" label="Replicas">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="image" label="Docker Image">
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

function OverviewTab({ endpoint, workers }: { endpoint: EndpointData; workers: Worker[] }) {
  const [apiMethod, setApiMethod] = useState<'run' | 'runsync' | 'status'>('run')
  const [codeLang, setCodeLang] = useState<'curl' | 'python' | 'js'>('curl')
  const [inputJson, setInputJson] = useState('{"prompt": "hello world"}')
  const [taskId, setTaskId] = useState('')
  const [result, setResult] = useState<string>('')
  const [loading, setLoading] = useState(false)

  const apiUrl = window.location.origin
  const ep = endpoint.logical_name

  const handleSubmit = async () => {
    setLoading(true)
    setResult('')
    try {
      if (apiMethod === 'status') {
        const res = await getTaskStatus(taskId)
        setResult(JSON.stringify(res, null, 2))
      } else {
        const input = JSON.parse(inputJson)
        const res = await submitTask(ep!, input, apiMethod === 'runsync')
        setResult(JSON.stringify(res, null, 2))
        if (res.id) setTaskId(res.id)
      }
    } catch (e: any) {
      setResult(e.response?.data?.error || e.message || 'Request failed')
    } finally {
      setLoading(false)
    }
  }

  const getCodeExample = () => {
    if (codeLang === 'curl') {
      if (apiMethod === 'status') return `curl ${apiUrl}/v1/status/${taskId || 'TASK_ID'}`
      return `curl -X POST ${apiUrl}/v1/${ep}/${apiMethod} \\
  -H "Content-Type: application/json" \\
  -d '{"input": ${inputJson}}'`
    }
    if (codeLang === 'python') {
      if (apiMethod === 'status') return `import requests\nres = requests.get("${apiUrl}/v1/status/${taskId || 'TASK_ID'}")\nprint(res.json())`
      return `import requests\nres = requests.post(\n    "${apiUrl}/v1/${ep}/${apiMethod}",\n    json={"input": ${inputJson}}\n)\nprint(res.json())`
    }
    if (apiMethod === 'status') return `const res = await fetch("${apiUrl}/v1/status/${taskId || 'TASK_ID'}");\nconsole.log(await res.json());`
    return `const res = await fetch("${apiUrl}/v1/${ep}/${apiMethod}", {\n  method: "POST",\n  headers: { "Content-Type": "application/json" },\n  body: JSON.stringify({ input: ${inputJson} })\n});\nconsole.log(await res.json());`
  }

  const copyCode = () => { navigator.clipboard.writeText(getCodeExample()); message.success('Copied!') }

  return (
    <>
      {/* Quick Start */}
      <Collapse defaultActiveKey={['quickstart']} className="mb-4" items={[{
        key: 'quickstart',
        label: <span><PlayCircleOutlined style={{ color: 'var(--warning)', marginRight: 8 }} />Quick Start</span>,
        children: (
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
            <div>
              <div className="flex gap-2 mb-3">
                {(['run', 'runsync', 'status'] as const).map(m => (
                  <button key={m} className={`btn btn-sm ${apiMethod === m ? 'btn-primary' : 'btn-outline'}`} onClick={() => setApiMethod(m)}>
                    {m === 'run' ? 'Async' : m === 'runsync' ? 'Sync' : 'Status'}
                  </button>
                ))}
              </div>
              {apiMethod !== 'status' ? (
                <div className="form-group mb-3">
                  <label className="form-label">Input JSON</label>
                  <textarea className="form-input" rows={4} value={inputJson} onChange={e => setInputJson(e.target.value)} style={{ fontFamily: 'monospace', fontSize: 12 }} />
                </div>
              ) : (
                <div className="form-group mb-3">
                  <label className="form-label">Task ID</label>
                  <input className="form-input" value={taskId} onChange={e => setTaskId(e.target.value)} placeholder="Enter task ID" />
                </div>
              )}
              <button className="btn btn-primary" onClick={handleSubmit} disabled={loading}>
                  <PlayCircleOutlined /> {loading ? 'Loading...' : apiMethod === 'status' ? 'Query' : 'Submit'}
                </button>
              {result && <pre style={{ marginTop: 12, background: 'var(--bg-primary)', padding: 12, borderRadius: 6, fontSize: 11, maxHeight: 150, overflow: 'auto' }}>{result}</pre>}
            </div>
            <div>
              <div className="flex gap-2 mb-2">
                {(['curl', 'python', 'js'] as const).map(l => (
                  <span key={l} className={`code-tab ${codeLang === l ? 'active' : ''}`} onClick={() => setCodeLang(l)}>{l}</span>
                ))}
                <button className="btn btn-outline btn-sm" style={{ marginLeft: 'auto', fontSize: 11 }} onClick={copyCode}><CopyOutlined /> Copy</button>
              </div>
              <pre style={{ background: 'var(--bg-primary)', padding: 12, borderRadius: 6, fontSize: 11, overflow: 'auto', height: 180 }}>{getCodeExample()}</pre>
            </div>
          </div>
        ),
      }]} />

      {/* Basic Info */}
      <div className="card mb-4">
        <div className="card-header"><h3>Basic Information</h3></div>
        <table className="info-table">
          <tbody>
            <tr><td className="info-label">Endpoint Name</td><td>{endpoint.logical_name || endpoint.name}</td><td className="info-label">Physical Name</td><td>{endpoint.physical_name || endpoint.name}</td></tr>
            <tr><td className="info-label">Spec</td><td>{endpoint.specName}</td><td className="info-label">Status</td><td><span className={`tag ${endpoint.status?.toLowerCase()}`}>{endpoint.status}</span></td></tr>
            <tr><td className="info-label">Cluster</td><td>{endpoint.cluster_id || '-'}</td><td className="info-label">Shared Memory</td><td>{endpoint.shmSize || '-'}</td></tr>
            <tr><td className="info-label">Task Timeout</td><td>{endpoint.taskTimeout}s</td><td className="info-label">Created</td><td>{endpoint.createdAt ? new Date(endpoint.createdAt).toLocaleString() : '-'}</td></tr>
          </tbody>
        </table>
      </div>

      {/* Image */}
      <div className="card mb-4">
        <div className="card-header"><h3>Image</h3></div>
        <table className="info-table">
          <tbody>
            <tr><td className="info-label">Image</td><td colSpan={3}><code style={{ fontSize: 11, wordBreak: 'break-all' }}>{endpoint.image}</code></td></tr>
          </tbody>
        </table>
      </div>

      {/* Scaling */}
      <div className="card">
        <div className="card-header"><h3>Scaling Configuration</h3></div>
        <table className="info-table">
          <tbody>
            <tr><td className="info-label">Replicas</td><td>{endpoint.readyReplicas || 0} / {endpoint.replicas || 0}</td><td className="info-label">Min/Max</td><td>{endpoint.minReplicas} / {endpoint.maxReplicas}</td></tr>
            <tr><td className="info-label">Current Workers</td><td>{workers.length}</td><td className="info-label">Price/hr</td><td>${(endpoint.price_per_hour || 0).toFixed(2)}</td></tr>
          </tbody>
        </table>
      </div>
    </>
  )
}

function MetricsTab({ name }: { name: string }) {
  const [timeRange, setTimeRange] = useState('1h')
  const [stats, setStats] = useState<any>(null)
  const requestsRef = useRef<HTMLDivElement>(null)
  const executionRef = useRef<HTMLDivElement>(null)
  const workersRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    getEndpointStats(name, timeRange).then(res => setStats(res?.stats || [])).catch(() => setStats([]))
  }, [name, timeRange])

  useEffect(() => {
    if (!stats?.length) return
    const labels = stats.map((s: any) => new Date(s.timestamp).toLocaleTimeString('en', { hour: '2-digit', minute: '2-digit' }))
    const grid = { left: 50, right: 20, top: 20, bottom: 30 }
    const axisLabel = { fontSize: 11, color: '#6b7280' }

    const init = (ref: React.RefObject<HTMLDivElement>, opt: echarts.EChartsOption) => {
      if (!ref.current) return
      let c = echarts.getInstanceByDom(ref.current)
      if (!c) c = echarts.init(ref.current)
      c.setOption(opt, true)
    }

    init(requestsRef, {
      grid, tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: labels, axisLabel },
      yAxis: { type: 'value', axisLabel },
      series: [
        { name: 'Completed', type: 'bar', stack: 'a', data: stats.map((s: any) => s.tasks_completed || 0), itemStyle: { color: '#48bb78' }, barMaxWidth: 10 },
        { name: 'Failed', type: 'bar', stack: 'a', data: stats.map((s: any) => s.tasks_failed || 0), itemStyle: { color: '#f56565' }, barMaxWidth: 10 },
      ]
    })

    init(executionRef, {
      grid, tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: labels, axisLabel },
      yAxis: { type: 'value', axisLabel: { ...axisLabel, formatter: (v: number) => (v / 1000).toFixed(0) + 's' } },
      series: [
        { name: 'P50', type: 'line', data: stats.map((s: any) => s.p50_execution_ms || s.avg_execution_ms || 0), smooth: true, symbol: 'none', lineStyle: { color: '#3b82f6' } },
        { name: 'P95', type: 'line', data: stats.map((s: any) => s.p95_execution_ms || 0), smooth: true, symbol: 'none', lineStyle: { color: '#f56565', type: 'dashed' } },
      ]
    })

    init(workersRef, {
      grid, tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: labels, axisLabel },
      yAxis: { type: 'value', minInterval: 1, axisLabel },
      series: [
        { name: 'Active', type: 'line', stack: 'a', data: stats.map((s: any) => s.active_workers || 0), smooth: true, symbol: 'none', areaStyle: { color: '#3b82f6' } },
        { name: 'Idle', type: 'line', stack: 'a', data: stats.map((s: any) => s.idle_workers || 0), smooth: true, symbol: 'none', areaStyle: { color: '#94a3b8' } },
      ]
    })
  }, [stats])

  const hasData = stats && stats.length > 0

  return (
    <>
      <div className="flex gap-2 mb-4">
        {['1h', '6h', '24h', '7d'].map(r => (
          <button key={r} className={`btn btn-sm ${timeRange === r ? 'btn-primary' : 'btn-outline'}`} onClick={() => setTimeRange(r)}>{r}</button>
        ))}
      </div>
      <div className="charts-grid">
        <div className="chart-card">
          <div className="chart-header"><span className="chart-title">Requests</span></div>
          <div className="chart-container" ref={requestsRef}>{!hasData && <span style={{ color: 'var(--text-secondary)' }}>No data</span>}</div>
        </div>
        <div className="chart-card">
          <div className="chart-header"><span className="chart-title">Execution Time</span></div>
          <div className="chart-container" ref={executionRef}>{!hasData && <span style={{ color: 'var(--text-secondary)' }}>No data</span>}</div>
        </div>
        <div className="chart-card">
          <div className="chart-header"><span className="chart-title">Worker Count</span></div>
          <div className="chart-container" ref={workersRef}>{!hasData && <span style={{ color: 'var(--text-secondary)' }}>No data</span>}</div>
        </div>
      </div>
    </>
  )
}

function WorkersTab({ workers, endpointName }: { workers: Worker[]; endpointName: string }) {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [selectedWorker, setSelectedWorker] = useState<Worker | null>(null)
  const [drawerTab, setDrawerTab] = useState('overview')
  const [logs, setLogs] = useState<string>('')
  const [tasks, setTasks] = useState<any[]>([])
  const [loadingLogs, setLoadingLogs] = useState(false)
  const [loadingTasks, setLoadingTasks] = useState(false)
  const [selectedTask, setSelectedTask] = useState<any>(null)
  const [taskDetailOpen, setTaskDetailOpen] = useState(false)

  const sortedWorkers = [...workers].sort((a, b) => (a.id || '').localeCompare(b.id || ''))
  const active = sortedWorkers.filter(w => (w.current_jobs || 0) > 0).length

  const formatTime = (t: string) => {
    if (!t) return '-'
    const s = Math.floor((Date.now() - new Date(t).getTime()) / 1000)
    return s < 60 ? `${s}s ago` : s < 3600 ? `${Math.floor(s/60)}m ago` : new Date(t).toLocaleString()
  }

  const getIdleTag = (w: Worker) => {
    if ((w.current_jobs || 0) > 0) return <span className="tag running">Active</span>
    if (!w.last_task_time) return <span className="tag" style={{ background: 'rgba(107,114,128,0.1)', color: '#6b7280' }}>New</span>
    const m = Math.floor((Date.now() - new Date(w.last_task_time).getTime()) / 60000)
    if (m < 5) return <span className="tag success">{m}m idle</span>
    if (m < 30) return <span className="tag pending">{m}m idle</span>
    return <span className="tag failed">{m}m idle</span>
  }

  const fetchLogs = async () => {
    if (!selectedWorker?.pod_name) return
    setLoadingLogs(true)
    try {
      const data = await getWorkerLogs(endpointName, selectedWorker.pod_name)
      setLogs(data || 'No logs')
    } catch { setLogs('Failed to fetch logs') }
    finally { setLoadingLogs(false) }
  }

  useEffect(() => {
    console.log('useEffect triggered:', { drawerOpen, drawerTab, workerId: selectedWorker?.id })
    if (!drawerOpen || !selectedWorker) return
    if (drawerTab === 'logs' && selectedWorker.pod_name) {
      setLoadingLogs(true)
      getWorkerLogs(endpointName, selectedWorker.pod_name)
        .then(data => setLogs(data || 'No logs'))
        .catch(() => setLogs('Failed to fetch logs'))
        .finally(() => setLoadingLogs(false))
    }
    if (drawerTab === 'tasks') {
      setLoadingTasks(true)
      getWorkerTasks(endpointName, selectedWorker.id)
        .then(data => { console.log('tasks response:', data); setTasks(data?.tasks || []) })
        .catch(e => { console.error('tasks error:', e); setTasks([]) })
        .finally(() => setLoadingTasks(false))
    }
  }, [drawerOpen, drawerTab, selectedWorker?.id, endpointName])

  return (
    <>
      <div className="stats-grid mb-4" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
        <div className="stat-card"><div className="stat-label">Active</div><div className="stat-value" style={{ color: 'var(--success)' }}>{active}</div></div>
        <div className="stat-card"><div className="stat-label">Idle</div><div className="stat-value" style={{ color: 'var(--warning)' }}>{sortedWorkers.length - active}</div></div>
        <div className="stat-card"><div className="stat-label">Total</div><div className="stat-value">{sortedWorkers.length}</div></div>
        <div className="stat-card"><div className="stat-label">Jobs</div><div className="stat-value">{sortedWorkers.reduce((s, w) => s + (w.current_jobs || 0), 0)}</div></div>
      </div>

      {sortedWorkers.length === 0 ? (
        <div className="card"><div className="empty-state">No workers running</div></div>
      ) : (
        sortedWorkers.map(w => (
          <div key={w.id || w.pod_name} className="worker-card" onClick={() => { setSelectedWorker(w); setDrawerTab('overview'); setDrawerOpen(true) }} style={{ cursor: 'pointer' }}>
            <div className={`worker-icon ${(w.current_jobs || 0) > 0 ? '' : 'idle'}`}>🖥️</div>
            <div className="worker-info" style={{ flex: 1 }}>
              <div className="worker-name">{w.pod_name || w.id}</div>
              <div className="worker-meta">
                <span className={`tag ${w.status?.toLowerCase() === 'online' || w.phase === 'Running' ? 'running' : 'pending'}`} style={{ marginRight: 4 }}>{w.status || w.phase}</span>
                Jobs: {w.current_jobs || 0}/{w.concurrency || 1} • Heartbeat: {formatTime(w.last_heartbeat)}
              </div>
            </div>
            {getIdleTag(w)}
          </div>
        ))
      )}

      <Drawer title={`Worker: ${selectedWorker?.pod_name || selectedWorker?.id || '-'}`} open={drawerOpen} width="70%" onClose={() => setDrawerOpen(false)}>
        <Tabs activeKey={drawerTab} onChange={(key) => { console.log('tab changed:', key); setDrawerTab(key) }} items={[
          { key: 'overview', label: '📊 Overview', children: selectedWorker && (
            <div className="card">
              <table className="info-table"><tbody>
                <tr><td className="info-label">ID</td><td>{selectedWorker.id}</td></tr>
                <tr><td className="info-label">Pod Name</td><td>{selectedWorker.pod_name || '-'}</td></tr>
                <tr><td className="info-label">Status</td><td><span className={`tag ${selectedWorker.status?.toLowerCase() === 'online' ? 'running' : 'pending'}`}>{selectedWorker.status || selectedWorker.phase}</span></td></tr>
                <tr><td className="info-label">Current Jobs</td><td>{selectedWorker.current_jobs || 0} / {selectedWorker.concurrency || 1}</td></tr>
                <tr><td className="info-label">Tasks Completed</td><td>{selectedWorker.tasks_completed || 0}</td></tr>
                <tr><td className="info-label">Last Heartbeat</td><td>{formatTime(selectedWorker.last_heartbeat)}</td></tr>
                <tr><td className="info-label">Last Task</td><td>{formatTime(selectedWorker.last_task_time)}</td></tr>
              </tbody></table>
            </div>
          )},
          { key: 'tasks', label: '📋 Tasks', children: (
            <div className="table-container">
              {loadingTasks ? <div className="loading"><div className="spinner"></div></div> : (
                <table><thead><tr><th>Task ID</th><th>Status</th><th>Created</th><th>Exec Time</th><th>Actions</th></tr></thead>
                  <tbody>
                    {tasks.map((t: any) => (
                      <tr key={t.id}>
                        <td style={{ fontFamily: 'monospace', fontSize: 11 }}>{t.id?.substring(0, 20)}...</td>
                        <td><span className={`tag ${t.status === 'COMPLETED' ? 'success' : t.status === 'FAILED' ? 'failed' : t.status === 'IN_PROGRESS' ? 'running' : 'pending'}`}>{t.status}</span></td>
                        <td style={{ fontSize: 12 }}>{t.createdAt ? new Date(t.createdAt).toLocaleString() : '-'}</td>
                        <td style={{ fontSize: 12 }}>{t.executionTime ? `${(t.executionTime/1000).toFixed(2)}s` : '-'}</td>
                        <td><button className="btn btn-sm btn-outline" onClick={() => { setSelectedTask(t); setTaskDetailOpen(true) }}><EyeOutlined /></button></td>
                      </tr>
                    ))}
                    {tasks.length === 0 && <tr><td colSpan={5} style={{ textAlign: 'center', color: 'var(--text-secondary)', padding: 20 }}>No tasks</td></tr>}
                  </tbody>
                </table>
              )}
            </div>
          )},
          { key: 'logs', label: '📄 Logs', children: (
            <div style={{ height: 'calc(100vh - 200px)', display: 'flex', flexDirection: 'column', background: '#0d1117', borderRadius: 8, overflow: 'hidden' }}>
              <div style={{ padding: '6px 12px', background: '#161b22', borderBottom: '1px solid #30363d', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ fontSize: 12, color: '#8b949e' }}>{loadingLogs ? 'Loading...' : `${logs.split('\n').length} lines`}</span>
                <button onClick={fetchLogs} disabled={loadingLogs} style={{ background: 'transparent', border: '1px solid #30363d', borderRadius: 6, padding: '4px 10px', color: '#c9d1d9', cursor: 'pointer', fontSize: 12, display: 'flex', alignItems: 'center', gap: 4 }}>
                  <SyncOutlined spin={loadingLogs} /> Refresh
                </button>
              </div>
              <pre style={{ flex: 1, margin: 0, padding: 12, background: '#0d1117', color: '#c9d1d9', fontSize: 12, overflow: 'auto', fontFamily: 'Menlo, Monaco, "Courier New", monospace' }}>{logs || 'No logs'}</pre>
            </div>
          )},
          { key: 'exec', label: '💻 Exec', children: selectedWorker && (
            <div style={{ height: 'calc(100vh - 200px)' }}>
              <Terminal endpoint={endpointName} workerId={selectedWorker.pod_name || selectedWorker.id} />
            </div>
          )},
        ]} />
      </Drawer>

      <Modal title="Task Details" open={taskDetailOpen} onCancel={() => setTaskDetailOpen(false)} footer={null} width={800}>
        {selectedTask && <TaskDetailContent task={selectedTask} />}
      </Modal>
    </>
  )
}

function TasksTab({ endpointName }: { endpointName: string }) {
  const [searchInput, setSearchInput] = useState('')
  const [statusInput, setStatusInput] = useState('all')
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(20)
  const [tasks, setTasks] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [selectedTask, setSelectedTask] = useState<any>(null)
  const [detailOpen, setDetailOpen] = useState(false)

  const fetchTasks = () => {
    setLoading(true)
    const params: any = { limit: pageSize, offset: page * pageSize }
    if (statusFilter !== 'all') params.status = statusFilter
    if (search) params.task_id = search
    getEndpointTasks(endpointName, params)
      .then(data => { setTasks(data?.tasks || []); setTotal(data?.total || 0) })
      .catch(() => { setTasks([]); setTotal(0) })
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchTasks() }, [endpointName, statusFilter, search, page, pageSize])

  const handleSearch = () => { setSearch(searchInput); setStatusFilter(statusInput); setPage(0) }
  const getStatusClass = (s: string) => s === 'COMPLETED' ? 'success' : s === 'FAILED' ? 'failed' : s === 'IN_PROGRESS' ? 'running' : 'pending'

  return (
    <>
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
          <button className="btn btn-primary" onClick={handleSearch}><SearchOutlined /> Search</button>
          <button className="btn btn-outline" onClick={() => { setSearchInput(''); setStatusInput('all'); setSearch(''); setStatusFilter('all'); setPage(0) }}>Reset</button>
        </div>
        <span style={{ color: 'var(--text-secondary)', fontSize: 13 }}>Total: {total}</span>
      </div>

      {/* Table */}
      <div className="card">
        <div className="table-container">
          <table>
            <thead><tr><th>Task ID</th><th>Status</th><th>Worker</th><th>Created</th><th>Exec Time</th><th>Delay</th><th>Actions</th></tr></thead>
            <tbody>
              {loading ? <tr><td colSpan={7}><div className="loading"><div className="spinner"></div></div></td></tr> :
               tasks.length === 0 ? <tr><td colSpan={7} style={{ textAlign: 'center', color: 'var(--text-secondary)', padding: 40 }}>No tasks</td></tr> :
               tasks.map((t: any) => (
                 <tr key={t.id}>
                   <td><Tooltip title={t.id}><span style={{ fontFamily: 'monospace', fontSize: 11 }}>{t.id?.substring(0, 20)}...</span></Tooltip></td>
                   <td><span className={`tag ${getStatusClass(t.status)}`}>{t.status}</span></td>
                   <td>{t.workerId ? <Tooltip title={t.workerId}><span style={{ fontSize: 12, color: 'var(--text-secondary)' }}>{t.workerId.substring(0, 12)}...</span></Tooltip> : '-'}</td>
                   <td style={{ fontSize: 12 }}>{t.createdAt ? new Date(t.createdAt).toLocaleString() : '-'}</td>
                   <td style={{ fontSize: 12 }}>{t.executionTime ? `${(t.executionTime/1000).toFixed(2)}s` : '-'}</td>
                   <td style={{ fontSize: 12 }}>{t.delayTime ? `${(t.delayTime/1000).toFixed(2)}s` : '-'}</td>
                   <td>
                     <button className="btn btn-sm btn-outline" onClick={() => { setSelectedTask(t); setDetailOpen(true) }}><EyeOutlined /></button>
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
      <Modal title="Task Details" open={detailOpen} onCancel={() => setDetailOpen(false)} footer={null} width={800}>
        {selectedTask && <TaskDetailContent task={selectedTask} />}
      </Modal>
    </>
  )
}

function TaskDetailContent({ task }: { task: any }) {
  const [fullTask, setFullTask] = useState<any>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (task?.id) {
      setLoading(true)
      getTaskStatus(task.id)
        .then(data => { console.log('task detail:', data); setFullTask(data) })
        .catch(e => { console.error('task detail error:', e); setFullTask(null) })
        .finally(() => setLoading(false))
    }
  }, [task?.id])

  const t = fullTask || task
  const getStatusClass = (s: string) => s === 'COMPLETED' ? 'success' : s === 'FAILED' ? 'failed' : s === 'IN_PROGRESS' ? 'running' : 'pending'

  if (loading) return <div className="loading"><div className="spinner"></div></div>

  return (
    <div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
        <div><label className="form-label">Task ID</label><div style={{ fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-all' }}>{t.id}</div></div>
        <div><label className="form-label">Status</label><span className={`tag ${getStatusClass(t.status)}`}>{t.status}</span></div>
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginTop: 16 }}>
        <div><label className="form-label">Endpoint</label><div>{t.endpoint}</div></div>
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
}

function SettingsTab({ endpoint, onDelete, onRefresh }: { endpoint: EndpointData; onDelete: () => void; onRefresh: () => void }) {
  return (
    <>
      <Collapse defaultActiveKey={['scaling']} accordion items={[
        { key: 'scaling', label: '⚡ Scaling Configuration', children: <ScalingPanel endpoint={endpoint} onRefresh={onRefresh} /> },
        { key: 'env', label: '🔧 Environment Variables', children: <EnvVarsPanel endpoint={endpoint} onRefresh={onRefresh} /> },
      ]} />
      <div className="card mt-4">
        <div className="card-header"><h3 style={{ color: 'var(--danger)' }}>Danger Zone</h3></div>
        <div style={{ padding: 16 }}>
          <p style={{ color: 'var(--text-secondary)', marginBottom: 16 }}>Deleting this endpoint will stop all workers and remove all configuration.</p>
          <Popconfirm title="Are you sure?" onConfirm={onDelete} okText="Yes" cancelText="No">
            <button className="btn" style={{ background: 'var(--danger)', color: '#fff' }}><DeleteOutlined /> Delete Endpoint</button>
          </Popconfirm>
        </div>
      </div>
    </>
  )
}

function ScalingPanel({ endpoint, onRefresh }: { endpoint: EndpointData; onRefresh: () => void }) {
  const [form] = Form.useForm()
  const [saving, setSaving] = useState(false)
  useEffect(() => { form.setFieldsValue({ minReplicas: endpoint.minReplicas || 0, maxReplicas: endpoint.maxReplicas || 10, taskTimeout: endpoint.taskTimeout || 3600 }) }, [endpoint, form])
  const handleSave = async () => {
    setSaving(true)
    try {
      await updateEndpointConfig(endpoint.logical_name!, form.getFieldsValue())
      message.success('Saved')
      onRefresh()
    } catch { message.error('Failed to save') }
    finally { setSaving(false) }
  }
  return (
    <Form form={form} layout="vertical">
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
        <Form.Item name="minReplicas" label="Min Replicas"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
        <Form.Item name="maxReplicas" label="Max Replicas"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
      </div>
      <Form.Item name="taskTimeout" label="Task Timeout (seconds)"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
      <button type="button" className="btn btn-primary" onClick={handleSave} disabled={saving}>{saving ? 'Saving...' : 'Save'}</button>
    </Form>
  )
}

function EnvVarsPanel({ endpoint, onRefresh }: { endpoint: EndpointData; onRefresh: () => void }) {
  const [envVars, setEnvVars] = useState<{ key: string; value: string }[]>(() => {
    const env = (endpoint as any).env || {}
    const vars = Object.entries(env).map(([key, value]) => ({ key, value: String(value) }))
    return vars.length > 0 ? vars : [{ key: '', value: '' }]
  })
  const [saving, setSaving] = useState(false)
  const handleSave = async () => {
    const env: Record<string, string> = {}
    envVars.filter(v => v.key).forEach(v => { env[v.key] = v.value })
    setSaving(true)
    try {
      await updateEndpoint(endpoint.logical_name!, { env })
      message.success('Saved')
      onRefresh()
    } catch { message.error('Failed to save') }
    finally { setSaving(false) }
  }
  return (
    <div>
      {envVars.map((v, i) => (
        <div key={i} className="flex gap-2 mb-2">
          <Input value={v.key} onChange={e => { const n = [...envVars]; n[i].key = e.target.value; setEnvVars(n) }} placeholder="KEY" style={{ width: 200 }} />
          <Input value={v.value} onChange={e => { const n = [...envVars]; n[i].value = e.target.value; setEnvVars(n) }} placeholder="value" style={{ flex: 1 }} />
          <button type="button" className="btn btn-outline btn-sm" onClick={() => setEnvVars(envVars.filter((_, idx) => idx !== i))}><DeleteOutlined /></button>
        </div>
      ))}
      <button type="button" className="btn btn-outline btn-sm mb-3" onClick={() => setEnvVars([...envVars, { key: '', value: '' }])}><PlusOutlined /> Add</button>
      <div><button type="button" className="btn btn-primary" onClick={handleSave} disabled={saving}>{saving ? 'Saving...' : 'Save'}</button></div>
    </div>
  )
}

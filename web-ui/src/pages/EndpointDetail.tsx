import React, { useState, useEffect, useRef, useMemo, useCallback } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { message, Popconfirm, Modal, Form, Input, InputNumber, Select, Collapse, Drawer, Tabs, Tooltip, Timeline } from 'antd'
import { ArrowLeftOutlined, DeleteOutlined, ReloadOutlined, EditOutlined, PlayCircleOutlined, CopyOutlined, PlusOutlined, SyncOutlined, SearchOutlined, EyeOutlined, StopOutlined } from '@ant-design/icons'
import * as echarts from 'echarts'
import { getEndpoint, getEndpointWorkers, getWorkerLogs, getEndpointMetrics, getEndpointStats, updateEndpoint, updateEndpointConfig, deleteEndpoint, submitTask, getTaskStatus, getEndpointTasks, cancelTask, getTaskTimeline, getTaskExecutionHistory, getEndpointStatistics } from '../api/client'
import Terminal from '../components/Terminal'

type TabKey = 'overview' | 'metrics' | 'workers' | 'tasks' | 'settings'

interface EndpointData {
  logical_name?: string
  physical_name?: string
  cluster_id?: string
  price_per_hour?: number
  name?: string
  specName?: string
  status?: string
  replicas?: number
  minReplicas?: number
  maxReplicas?: number
  readyReplicas?: number
  availableReplicas?: number
  image?: string
  imagePrefix?: string
  imageDigest?: string
  imageLastChecked?: string
  imageUpdateAvailable?: boolean
  taskTimeout?: number
  maxPendingTasks?: number
  createdAt?: string
  shmSize?: string
  displayName?: string
  description?: string
  namespace?: string
  type?: string
  env?: Record<string, string>
  scaleUpThreshold?: number
  scaleDownIdleTime?: number
  scaleUpCooldown?: number
  scaleDownCooldown?: number
  priority?: number
  enableDynamicPrio?: boolean
  highLoadThreshold?: number
  priorityBoost?: number
  autoscalerEnabled?: string
  enablePtrace?: boolean
  volumeMounts?: { pvcName: string; mountPath: string }[]
}

interface Worker {
  id: string
  worker_id: string
  pod_name: string
  status: string
  current_jobs: number
  concurrency?: number
  total_tasks_completed: number
  total_tasks_failed: number
  last_heartbeat: string
  last_task_time: string
  pod_created_at: string
  pod_started_at: string
  cold_start_duration_ms: number
  version?: string
  podStatus?: string
}

interface Task {
  id: string
  endpoint?: string
  status: string
  workerId?: string
  delayTime?: number
  executionTime?: number
  createdAt?: string
  input?: Record<string, any>
  output?: Record<string, any>
  error?: string
}

interface TaskStats {
  pendingTasks?: number
  runningTasks?: number
  completedTasks?: number
  failedTasks?: number
  onlineWorkers?: number
  busyWorkers?: number
}

export default function EndpointDetail() {
  const { name } = useParams<{ name: string }>()
  const navigate = useNavigate()
  const [endpoint, setEndpoint] = useState<EndpointData | null>(null)
  const [workers, setWorkers] = useState<Worker[]>([])
  const [taskStats, setTaskStats] = useState<TaskStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<TabKey>('overview')
  const [editModalOpen, setEditModalOpen] = useState(false)
  const [editLoading, setEditLoading] = useState(false)
  const [form] = Form.useForm()

  const fetchData = async () => {
    if (!name) return
    try {
      const [ep, wk, stats] = await Promise.all([
        getEndpoint(name),
        getEndpointWorkers(name).catch(() => []),
        getEndpointStatistics(name).catch(() => null),
      ])
      setEndpoint(ep)
      setWorkers(Array.isArray(wk) ? wk : wk?.workers || [])
      setTaskStats(stats)
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
      setEditLoading(true)
      const values = await form.validateFields()
      await updateEndpoint(name!, values)
      message.success('Updated')
      setEditModalOpen(false)
      fetchData()
    } catch {
      message.error('Failed to update')
    } finally {
      setEditLoading(false)
    }
  }

  const openEdit = () => {
    form.setFieldsValue({ replicas: endpoint?.replicas || 1, image: endpoint?.image })
    setEditModalOpen(true)
  }

  const getStatusClass = (s: string) => s === 'Running' ? 'running' : s === 'Pending' ? 'pending' : s === 'Failed' ? 'failed' : 'stopped'

  if (loading) return <div className="loading"><div className="spinner"></div></div>
  if (!endpoint) return <div className="empty-state"><p>Endpoint not found</p><Link to="/endpoints" className="btn btn-outline mt-4">Back</Link></div>

  const ep = endpoint

  return (
    <>
      {/* Header */}
      <div className="flex justify-between items-center mb-4">
        <div className="flex items-center gap-3">
          <Link to="/endpoints" className="btn btn-outline btn-icon"><ArrowLeftOutlined /></Link>
          <h2 style={{ margin: 0, fontSize: 20, fontWeight: 600 }}>{name}</h2>
          <span className={`tag ${getStatusClass(ep.status || '')}`}>{ep.status}</span>
        </div>
        <div className="flex gap-2">
          <button className="btn btn-outline" onClick={() => fetchData()}><ReloadOutlined /> Refresh</button>
          <button className="btn btn-outline" onClick={openEdit}><EditOutlined /> Edit</button>
          <Popconfirm title="Delete this endpoint?" onConfirm={handleDelete} okText="Yes" cancelText="No">
            <button className="btn btn-outline" style={{ color: '#f56565', borderColor: '#f56565' }}><DeleteOutlined /> Delete</button>
          </Popconfirm>
        </div>
      </div>

      {/* Meta */}
      <div style={{ fontSize: 13, color: 'var(--text-secondary)', marginBottom: 12 }}>
        {ep.specName} • {ep.cluster_id || '-'} • ${(ep.price_per_hour || 0).toFixed(2)}/hr • Created {ep.createdAt ? new Date(ep.createdAt).toLocaleDateString() : '-'}
      </div>

      {/* Stats */}
      <div className="stats-grid mb-4" style={{ gridTemplateColumns: 'repeat(5, 1fr)' }}>
        <div className="stat-card"><div className="stat-content"><div className="stat-label">Workers</div><div className="stat-value" style={{ color: '#3b82f6' }}><span style={{ fontSize: 28 }}>{taskStats?.busyWorkers || 0}</span><span style={{ fontSize: 14, color: 'var(--text-secondary)' }}> / {taskStats?.onlineWorkers || workers.length}</span></div></div></div>
        <div className="stat-card"><div className="stat-content"><div className="stat-label">Replicas</div><div className="stat-value" style={{ color: '#10b981' }}><span style={{ fontSize: 28 }}>{ep.readyReplicas || 0}</span><span style={{ fontSize: 14, color: 'var(--text-secondary)' }}> / {ep.replicas || 0}</span></div></div></div>
        <div className="stat-card"><div className="stat-content"><div className="stat-label">Running</div><div className="stat-value" style={{ color: '#3b82f6', fontSize: 28 }}>{taskStats?.runningTasks || 0}</div></div></div>
        <div className="stat-card"><div className="stat-content"><div className="stat-label">Pending</div><div className="stat-value" style={{ color: '#f59e0b', fontSize: 28 }}>{taskStats?.pendingTasks || 0}</div></div></div>
        <div className="stat-card"><div className="stat-content"><div className="stat-label">Total</div><div className="stat-value" style={{ fontSize: 28 }}>{(taskStats?.pendingTasks || 0) + (taskStats?.runningTasks || 0) + (taskStats?.completedTasks || 0) + (taskStats?.failedTasks || 0)}</div></div></div>
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
      {activeTab === 'overview' && <OverviewTab endpoint={ep} />}
      {activeTab === 'metrics' && <MetricsTab name={name!} />}
      {activeTab === 'workers' && <WorkersTab workers={workers} endpointName={name!} />}
      {activeTab === 'tasks' && <TasksTab endpointName={name!} />}
      {activeTab === 'settings' && <SettingsTab endpoint={ep} onRefresh={fetchData} />}

      {/* Edit Modal */}
      <Modal title="Quick Edit" open={editModalOpen} onOk={handleUpdate} onCancel={() => setEditModalOpen(false)} confirmLoading={editLoading}>
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="replicas" label="Replicas"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="image" label="Docker Image"><Input /></Form.Item>
        </Form>
      </Modal>
    </>
  )
}

function OverviewTab({ endpoint }: { endpoint: EndpointData }) {
  const [apiMethod, setApiMethod] = useState<'run' | 'runsync' | 'status'>('run')
  const [codeLang, setCodeLang] = useState<'curl' | 'python' | 'js'>('curl')
  const [inputJson, setInputJson] = useState('{"prompt": "a beautiful sunset"}')
  const [taskId, setTaskId] = useState('')
  const [result, setResult] = useState<string>('')
  const [loading, setLoading] = useState(false)

  const apiUrl = window.location.origin
  const ep = endpoint.logical_name || endpoint.name

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
    const authHeader = '-H "Authorization: Bearer WAVESPEED_API_KEY"'
    if (codeLang === 'curl') {
      if (apiMethod === 'status') return `curl -X GET ${apiUrl}/v1/status/${taskId || 'TASK_ID'} \\\n  ${authHeader}`
      return `curl -X POST ${apiUrl}/v1/${ep}/${apiMethod} \\\n  ${authHeader} \\\n  -H "Content-Type: application/json" \\\n  -d '{"input": ${inputJson}}'`
    }
    if (codeLang === 'python') {
      const headers = 'headers = {"Authorization": "Bearer WAVESPEED_API_KEY"}'
      if (apiMethod === 'status') return `import requests\n\n${headers}\nres = requests.get("${apiUrl}/v1/status/${taskId || 'TASK_ID'}", headers=headers)\nprint(res.json())`
      return `import requests\n\n${headers}\nres = requests.post(\n    "${apiUrl}/v1/${ep}/${apiMethod}",\n    headers=headers,\n    json={"input": ${inputJson}}\n)\nprint(res.json())`
    }
    const jsHeaders = '  headers: {\n    "Authorization": "Bearer WAVESPEED_API_KEY",\n    "Content-Type": "application/json"\n  },'
    if (apiMethod === 'status') return `const res = await fetch("${apiUrl}/v1/status/${taskId || 'TASK_ID'}", {\n  headers: { "Authorization": "Bearer YOUR_API_KEY" }\n});\nconsole.log(await res.json());`
    return `const res = await fetch("${apiUrl}/v1/${ep}/${apiMethod}", {\n  method: "POST",\n${jsHeaders}\n  body: JSON.stringify({ input: ${inputJson} })\n});\nconsole.log(await res.json());`
  }

  const copyCode = () => { navigator.clipboard.writeText(getCodeExample()); message.success('Copied!') }

  return (
    <>
      {/* Quick Start */}
      <Collapse defaultActiveKey={['quickstart']} className="mb-4" items={[{
        key: 'quickstart',
        label: '⚡ Quick Start',
        children: (
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
            <div>
              <div className="flex gap-2 mb-3">
                {(['run', 'runsync', 'status'] as const).map(m => (
                  <button key={m} className={`btn btn-sm ${apiMethod === m ? 'btn-blue' : 'btn-outline'}`} onClick={() => setApiMethod(m)}>
                    {m === 'run' ? 'Async' : m === 'runsync' ? 'Sync' : 'Status'}
                  </button>
                ))}
              </div>
              <div className="form-group mb-3" style={{ minHeight: 110 }}>
                {apiMethod !== 'status' ? (
                  <>
                    <label className="form-label">Input JSON</label>
                    <textarea className="form-input" rows={4} value={inputJson} onChange={e => setInputJson(e.target.value)} style={{ fontFamily: 'monospace', fontSize: 12 }} />
                  </>
                ) : (
                  <>
                    <label className="form-label">Task ID</label>
                    <Input value={taskId} onChange={e => setTaskId(e.target.value)} placeholder="Enter task ID" />
                  </>
                )}
              </div>
              <button className="btn btn-blue" onClick={handleSubmit} disabled={loading}>
                <PlayCircleOutlined /> {loading ? 'Loading...' : apiMethod === 'status' ? 'Query' : 'Submit'}
              </button>
              {result && <pre style={{ marginTop: 12, background: 'var(--bg-hover)', color: 'var(--text-primary)', padding: 10, borderRadius: 6, fontSize: 11, maxHeight: 120, overflow: 'auto' }}>{result}</pre>}
            </div>
            <div style={{ minWidth: 0 }}>
              <div className="flex gap-2 mb-2">
                {(['curl', 'python', 'js'] as const).map(l => (
                  <span key={l} className={`code-tab ${codeLang === l ? 'active' : ''}`} onClick={() => setCodeLang(l)}>{l}</span>
                ))}
                <button className="btn btn-outline btn-sm" style={{ marginLeft: 'auto', fontSize: 11 }} onClick={copyCode}><CopyOutlined /> Copy</button>
              </div>
              <pre style={{ background: 'var(--bg-hover)', color: 'var(--text-primary)', padding: 12, borderRadius: 6, fontSize: 11, overflow: 'auto', height: 180, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{getCodeExample()}</pre>
            </div>
          </div>
        ),
      }]} />

      {/* Basic Info */}
      <div className="card mb-4">
        <div className="card-header"><h3>Basic Information</h3></div>
        <table className="info-table"><tbody>
          <tr><td className="info-label">Endpoint Name</td><td>{endpoint.logical_name || endpoint.name}</td><td className="info-label">Spec</td><td>{endpoint.specName || 'N/A'}</td></tr>
          <tr><td className="info-label">Created At</td><td>{endpoint.createdAt ? new Date(endpoint.createdAt).toLocaleString() : '-'}</td><td className="info-label">Image</td><td><code style={{ background: 'var(--bg-hover)', padding: '2px 6px', borderRadius: 4, fontSize: 11 }}>{endpoint.image}</code></td></tr>
        </tbody></table>
      </div>

      {/* AutoScaler Config */}
      <div className="card mb-4">
        <div className="card-header"><h3>AutoScaler Configuration</h3></div>
        <table className="info-table"><tbody>
          <tr><td className="info-label">Min Replicas</td><td>{endpoint.minReplicas || 0}</td><td className="info-label">Max Replicas</td><td>{endpoint.maxReplicas || 10}</td></tr>
        </tbody></table>
      </div>
    </>
  )
}

const ChartCard = React.memo(({ chartRef, title, total, legend, hasData }: { chartRef: React.RefObject<HTMLDivElement>; title: string; total?: string; legend?: { color: string; label: string }[]; hasData: boolean }) => (
  <div className="chart-card">
    <div className="chart-header"><span className="chart-title">{title}</span>{total && <span className="chart-total">{total}</span>}</div>
    {legend && <div className="chart-legend">{legend.map((l, i) => <span key={i} className="legend-item"><span className="legend-dot" style={{ background: l.color }}></span>{l.label}</span>)}</div>}
    <div className="chart-container" ref={chartRef} style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      {!hasData && <span style={{ color: 'var(--text-secondary)', fontSize: 13 }}>No data for selected time range</span>}
    </div>
  </div>
))

function MetricsTab({ name }: { name: string }) {
  const [timeRange, setTimeRange] = useState('1h')
  const [liveMode, setLiveMode] = useState(true)
  const [showRangeDropdown, setShowRangeDropdown] = useState(false)
  const [realtimeData, setRealtimeData] = useState<any>(null)
  const [statsData, setStatsData] = useState<any[]>([])
  const [granularity, setGranularity] = useState('1h')
  const requestsRef = useRef<HTMLDivElement>(null)
  const executionRef = useRef<HTMLDivElement>(null)
  const delayRef = useRef<HTMLDivElement>(null)
  const coldStartCountRef = useRef<HTMLDivElement>(null)
  const coldStartTimeRef = useRef<HTMLDivElement>(null)
  const workersRef = useRef<HTMLDivElement>(null)
  const utilizationRef = useRef<HTMLDivElement>(null)
  const idleTimeRef = useRef<HTMLDivElement>(null)

  const rangeMs: Record<string, number> = { '1h': 3600000, '6h': 6 * 3600000, '24h': 24 * 3600000, '7d': 7 * 24 * 3600000, '30d': 30 * 24 * 3600000 }
  const rangeLabels: Record<string, string> = { '1h': 'Last 1 hour', '6h': 'Last 6 hours', '24h': 'Last 24 hours', '7d': 'Last 7 days', '30d': 'Last 30 days' }

  useEffect(() => {
    const fetchRealtime = async () => {
      try {
        const data = await getEndpointMetrics(name)
        setRealtimeData(data)
      } catch {}
    }
    fetchRealtime()
    if (liveMode) {
      const interval = setInterval(fetchRealtime, 5000)
      return () => clearInterval(interval)
    }
  }, [name, liveMode])

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const from = new Date(Date.now() - rangeMs[timeRange]).toISOString()
        const to = new Date().toISOString()
        const res = await getEndpointStats(name, timeRange, from, to)
        console.log('Stats API response:', res)
        const stats = res?.stats || []
        setStatsData(Array.isArray(stats) ? stats : [])
        setGranularity(res?.granularity || '1h')
      } catch (e) {
        console.error('Stats API error:', e)
      }
    }
    fetchStats()
    if (liveMode) {
      const interval = setInterval(fetchStats, 60000)
      return () => clearInterval(interval)
    }
  }, [name, timeRange, liveMode])

  const grid = useMemo(() => ({ left: 50, right: 20, top: 20, bottom: 50 }), [])
  const axisLabel = useMemo(() => ({ fontSize: 11, color: '#9ca3af' }), [])
  const dataZoom = useMemo(() => [{ type: 'inside', start: 0, end: 100 }, { type: 'slider', start: 0, end: 100, height: 20, bottom: 5, textStyle: { color: '#9ca3af' } }], [])

  const { totals, avgValues } = useMemo(() => {
    const data = Array.isArray(statsData) ? statsData : []
    const sum = (key: string) => data.reduce((a: number, s: any) => a + (s[key] || 0), 0)
    const avg = (key: string) => data.length ? (sum(key) / data.length).toFixed(1) : '0'
    return {
      totals: { submitted: sum('tasks_submitted'), completed: sum('tasks_completed'), failed: sum('tasks_failed'), retried: sum('tasks_retried'), coldStarts: sum('cold_starts') },
      avgValues: { activeWorkers: avg('active_workers'), workerUtilization: avg('avg_worker_utilization') }
    }
  }, [statsData])

  const formatLabel = useCallback((ts: string) => {
    const d = new Date(ts)
    return granularity === '1d' ? d.toLocaleDateString('en', { month: 'short', day: 'numeric' }) : d.toLocaleTimeString('en', { hour: '2-digit', minute: '2-digit' })
  }, [granularity])

  useEffect(() => {
    const data = Array.isArray(statsData) ? statsData : []
    if (!data.length) return
    const labels = data.map((s: any) => formatLabel(s.timestamp))
    const init = (ref: React.RefObject<HTMLDivElement>, opt: echarts.EChartsOption) => {
      if (!ref.current) return
      let c = echarts.getInstanceByDom(ref.current)
      if (!c) c = echarts.init(ref.current)
      c.setOption({ ...opt, dataZoom }, true)
    }

    init(requestsRef, { grid, tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } }, xAxis: { type: 'category', data: labels, axisLabel }, yAxis: { type: 'value', axisLabel }, series: [
      { name: 'Submitted', type: 'bar', stack: 'total', data: data.map((s: any) => s.tasks_submitted || 0), itemStyle: { color: '#3b82f6' }, barMaxWidth: 10 },
      { name: 'Completed', type: 'bar', stack: 'total', data: data.map((s: any) => s.tasks_completed || 0), itemStyle: { color: '#48bb78' }, barMaxWidth: 10 },
      { name: 'Failed', type: 'bar', stack: 'total', data: data.map((s: any) => s.tasks_failed || 0), itemStyle: { color: '#f56565' }, barMaxWidth: 10 },
      { name: 'Retried', type: 'bar', stack: 'total', data: data.map((s: any) => s.tasks_retried || 0), itemStyle: { color: '#ecc94b' }, barMaxWidth: 10 },
    ]})
    init(executionRef, { grid, tooltip: { trigger: 'axis' }, xAxis: { type: 'category', data: labels, axisLabel }, yAxis: { type: 'value', axisLabel: { ...axisLabel, formatter: (v: number) => (v / 1000).toFixed(0) + 's' } }, series: [
      { name: 'P95', type: 'line', data: data.map((s: any) => s.p95_execution_ms || 0), smooth: true, symbol: 'none', lineStyle: { width: 1, color: '#f56565' }, areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(245,101,101,0.3)' }, { offset: 1, color: 'rgba(245,101,101,0.05)' }] } } },
      { name: 'P50', type: 'line', data: data.map((s: any) => s.p50_execution_ms || s.avg_execution_ms || 0), smooth: true, symbol: 'none', lineStyle: { width: 1, color: '#3b82f6' }, areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(59,130,246,0.3)' }, { offset: 1, color: 'rgba(59,130,246,0.05)' }] } } },
    ]})
    init(delayRef, { grid, tooltip: { trigger: 'axis' }, xAxis: { type: 'category', data: labels, axisLabel }, yAxis: { type: 'value', axisLabel: { ...axisLabel, formatter: (v: number) => (v / 1000).toFixed(0) + 's' } }, series: [{ name: 'Delay', type: 'bar', data: data.map((s: any) => s.avg_queue_wait_ms || 0), itemStyle: { color: '#f56565' }, barMaxWidth: 8 }] })
    init(coldStartCountRef, { grid, tooltip: { trigger: 'axis' }, xAxis: { type: 'category', data: labels, axisLabel }, yAxis: { type: 'value', axisLabel }, series: [{ name: 'Count', type: 'bar', data: data.map((s: any) => s.cold_starts || 0), itemStyle: { color: '#8b5cf6' }, barMaxWidth: 8 }] })
    init(coldStartTimeRef, { grid, tooltip: { trigger: 'axis' }, xAxis: { type: 'category', data: labels, axisLabel }, yAxis: { type: 'value', axisLabel: { ...axisLabel, formatter: (v: number) => v.toFixed(0) + 's' } }, series: [{ name: 'Time', type: 'bar', data: data.map((s: any) => (s.avg_cold_start_ms || 0) / 1000), itemStyle: { color: '#06b6d4' }, barMaxWidth: 8 }] })
    init(workersRef, { grid, tooltip: { trigger: 'axis', axisPointer: { type: 'cross' } }, xAxis: { type: 'category', data: labels, axisLabel }, yAxis: { type: 'value', minInterval: 1, axisLabel }, series: [
      { name: 'Active', type: 'line', stack: 'total', data: data.map((s: any) => s.active_workers || 0), smooth: true, symbol: 'none', lineStyle: { width: 0 }, areaStyle: { color: '#3b82f6' } },
      { name: 'Idle', type: 'line', stack: 'total', data: data.map((s: any) => s.idle_workers || 0), smooth: true, symbol: 'none', lineStyle: { width: 0 }, areaStyle: { color: '#94a3b8' } },
    ]})
    init(utilizationRef, { grid, tooltip: { trigger: 'axis' }, xAxis: { type: 'category', data: labels, axisLabel }, yAxis: { type: 'value', max: 100, axisLabel: { ...axisLabel, formatter: (v: number) => v + '%' } }, series: [{ name: 'Utilization', type: 'line', data: data.map((s: any) => s.avg_worker_utilization || 0), smooth: true, symbol: 'none', lineStyle: { width: 2, color: '#10b981' }, areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(16,185,129,0.3)' }, { offset: 1, color: 'rgba(16,185,129,0.05)' }] } } }] })
    init(idleTimeRef, { grid, tooltip: { trigger: 'axis' }, xAxis: { type: 'category', data: labels, axisLabel }, yAxis: { type: 'value', axisLabel: { ...axisLabel, formatter: (v: number) => v.toFixed(0) + 's' } }, series: [{ name: 'Idle Time', type: 'line', data: data.map((s: any) => s.avg_idle_duration_sec || 0), smooth: true, symbol: 'none', lineStyle: { width: 0 }, areaStyle: { color: '#f59e0b' } }] })
  }, [statsData, granularity, formatLabel, grid, axisLabel, dataZoom])

  const hasData = Array.isArray(statsData) && statsData.length > 0

  return (
    <>
      {/* Realtime Stats */}
      {realtimeData && (
        <div className="stats-grid mb-4" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
          <div className="stat-card"><div className="stat-content"><div className="stat-label">Active Workers</div><div className="stat-value">{realtimeData.workers?.active || 0}</div><div className="stat-change">{realtimeData.workers?.idle || 0} idle</div></div></div>
          <div className="stat-card"><div className="stat-content"><div className="stat-label">Tasks/min</div><div className="stat-value">{realtimeData.tasks?.completed_last_minute || 0}</div><div className="stat-change">{realtimeData.tasks?.running || 0} running</div></div></div>
          <div className="stat-card"><div className="stat-content"><div className="stat-label">Avg Execution</div><div className="stat-value">{((realtimeData.performance?.avg_execution_ms || 0) / 1000).toFixed(1)}s</div></div></div>
          <div className="stat-card"><div className="stat-content"><div className="stat-label">Queue Wait</div><div className="stat-value">{Math.round(realtimeData.performance?.avg_queue_wait_ms || 0)}ms</div></div></div>
        </div>
      )}
      {/* Controls */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16, padding: '12px 16px', background: 'var(--bg-card)', borderRadius: 8, border: '1px solid var(--border-color)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <div style={{ position: 'relative' }}>
            <button onClick={() => setShowRangeDropdown(!showRangeDropdown)} style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '8px 16px', border: '1px solid var(--border-color)', borderRadius: 6, background: 'var(--bg-card)', cursor: 'pointer', fontSize: 14, color: 'var(--text-primary)' }}>
              <span>📅</span><span>{rangeLabels[timeRange]}</span><span style={{ marginLeft: 4 }}>▼</span>
            </button>
            {showRangeDropdown && (
              <div style={{ position: 'absolute', top: '100%', left: 0, marginTop: 4, background: 'var(--bg-card)', border: '1px solid var(--border-color)', borderRadius: 8, boxShadow: '0 4px 12px rgba(0,0,0,0.15)', padding: 8, zIndex: 100, minWidth: 160 }}>
                {Object.entries(rangeLabels).map(([key, label]) => (
                  <div key={key} onClick={() => { setTimeRange(key); setShowRangeDropdown(false) }} style={{ padding: '10px 16px', cursor: 'pointer', borderRadius: 4, fontSize: 14, color: 'var(--text-primary)', background: timeRange === key ? 'var(--bg-hover)' : 'transparent' }}>{label}</div>
                ))}
              </div>
            )}
          </div>
          <span style={{ fontSize: 14, color: 'var(--text-secondary)' }}>Granularity: <strong>{granularity === '1m' ? 'Minute' : granularity === '1h' ? 'Hourly' : 'Daily'}</strong></span>
        </div>
        <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 14, color: 'var(--text-primary)', cursor: 'pointer' }}>
          <input type="checkbox" checked={liveMode} onChange={(e) => setLiveMode(e.target.checked)} style={{ width: 16, height: 16, cursor: 'pointer' }} />
          <span style={{ width: 8, height: 8, borderRadius: '50%', background: liveMode ? '#48bb78' : 'var(--border-color)' }} />
          <span>View live data</span>
        </label>
      </div>
      <div className="charts-grid">
        <ChartCard chartRef={requestsRef} title="Requests" total={`Submitted: ${totals.submitted}`} legend={[{ color: '#3b82f6', label: `Submitted: ${totals.submitted}` }, { color: '#48bb78', label: `Completed: ${totals.completed}` }, { color: '#f56565', label: `Failed: ${totals.failed}` }, { color: '#ecc94b', label: `Retried: ${totals.retried}` }]} hasData={hasData} />
        <ChartCard chartRef={executionRef} title="Execution Time" legend={[{ color: '#3b82f6', label: 'P50' }, { color: '#f56565', label: 'P95' }]} hasData={hasData} />
        <ChartCard chartRef={delayRef} title="Queue Wait Time" hasData={hasData} />
        <ChartCard chartRef={workersRef} title="Worker Count" total={`Avg: ${avgValues.activeWorkers}`} legend={[{ color: '#3b82f6', label: 'Active' }, { color: '#94a3b8', label: 'Idle' }]} hasData={hasData} />
        <ChartCard chartRef={utilizationRef} title="Worker Utilization" total={`Avg: ${avgValues.workerUtilization}%`} hasData={hasData} />
        <ChartCard chartRef={idleTimeRef} title="Worker Idle Time (Avg)" hasData={hasData} />
        <ChartCard chartRef={coldStartCountRef} title="Cold Start Count" total={`Total: ${totals.coldStarts}`} hasData={hasData} />
        <ChartCard chartRef={coldStartTimeRef} title="Cold Start Time" hasData={hasData} />
      </div>
    </>
  )
}

function WorkersTab({ workers, endpointName }: { workers: Worker[]; endpointName: string }) {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [selectedWorker, setSelectedWorker] = useState<Worker | null>(null)
  const [drawerTab, setDrawerTab] = useState('tasks')
  const [logs, setLogs] = useState<string>('')
  const [tasks, setTasks] = useState<Task[]>([])
  const [loadingLogs, setLoadingLogs] = useState(false)
  const [loadingTasks, setLoadingTasks] = useState(false)
  const [selectedTask, setSelectedTask] = useState<Task | null>(null)
  const [taskDetailOpen, setTaskDetailOpen] = useState(false)

  const sortedWorkers = [...workers].sort((a, b) => String(a.id || '').localeCompare(String(b.id || '')))
  const active = sortedWorkers.filter(w => w.current_jobs > 0).length

  const formatTime = (t: string) => {
    if (!t) return '-'
    const s = Math.floor((Date.now() - new Date(t).getTime()) / 1000)
    return s < 60 ? `${s}s ago` : s < 3600 ? `${Math.floor(s/60)}m ago` : new Date(t).toLocaleString()
  }

  const getIdleTag = (w: Worker) => {
    if (w.current_jobs > 0) return <span className="tag running">Active</span>
    if (!w.last_task_time) return <span className="tag" style={{ background: 'rgba(107,114,128,0.1)', color: 'var(--text-secondary)' }}>New</span>
    const m = Math.floor((Date.now() - new Date(w.last_task_time).getTime()) / 60000)
    if (m < 5) return <span className="tag success">{m}m idle</span>
    if (m < 30) return <span className="tag pending">{m}m idle</span>
    return <span className="tag failed">{m}m idle</span>
  }

  useEffect(() => {
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
      getEndpointTasks(endpointName, { worker_id: selectedWorker.worker_id || selectedWorker.id, limit: 100 })
        .then(data => setTasks(data?.tasks || []))
        .catch(() => setTasks([]))
        .finally(() => setLoadingTasks(false))
    }
  }, [drawerOpen, drawerTab, selectedWorker?.id, endpointName])

  return (
    <>
      <div className="stats-grid mb-4" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
        <div className="stat-card"><div className="stat-content"><div className="stat-label">Active</div><div className="stat-value" style={{ color: '#48bb78' }}>{active}</div></div></div>
        <div className="stat-card"><div className="stat-content"><div className="stat-label">Idle</div><div className="stat-value" style={{ color: '#f59e0b' }}>{sortedWorkers.length - active}</div></div></div>
        <div className="stat-card"><div className="stat-content"><div className="stat-label">Total</div><div className="stat-value">{sortedWorkers.length}</div></div></div>
        <div className="stat-card"><div className="stat-content"><div className="stat-label">Jobs</div><div className="stat-value">{sortedWorkers.reduce((s, w) => s + (w.current_jobs || 0), 0)}</div></div></div>
      </div>

      {sortedWorkers.map(w => (
        <div key={w.id} className="worker-card" onClick={() => { setSelectedWorker(w); setDrawerTab('tasks'); setDrawerOpen(true) }} style={{ cursor: 'pointer' }}>
          <div className={`worker-icon ${w.current_jobs > 0 ? '' : 'idle'}`}>🖥️</div>
          <div className="worker-info" style={{ flex: 1 }}>
            <div className="worker-name">{w.pod_name || w.worker_id || w.id}</div>
            <div className="worker-meta">
              <span className={`tag ${w.status?.toUpperCase() === 'ONLINE' ? 'running' : w.status?.toUpperCase() === 'BUSY' ? 'running' : w.status?.toUpperCase() === 'STARTING' ? 'pending' : w.status?.toUpperCase() === 'DRAINING' ? 'pending' : 'failed'}`} style={{ marginRight: 4 }}>{w.status}</span>
              {w.podStatus && <span className={`tag ${w.podStatus === 'Running' ? 'running' : 'pending'}`} style={{ marginRight: 8 }}>Pod: {w.podStatus}</span>}
              Jobs: {w.current_jobs || 0}/{w.concurrency || 1} • Heartbeat: {formatTime(w.last_heartbeat)} {w.version && `• v${w.version}`}
            </div>
          </div>
          {getIdleTag(w)}
        </div>
      ))}
      {sortedWorkers.length === 0 && <div className="card" style={{ padding: 40, textAlign: 'center', color: 'var(--text-secondary)' }}>No workers</div>}

      <Drawer title={`Worker: ${selectedWorker?.pod_name || selectedWorker?.worker_id || '-'}`} open={drawerOpen} width="70%" onClose={() => setDrawerOpen(false)}>
        <Tabs activeKey={drawerTab} onChange={setDrawerTab} items={[
          { key: 'tasks', label: '📋 Tasks', children: (
            <div className="table-container">
              {loadingTasks ? <div className="loading"><div className="spinner"></div></div> : (
                <table><thead><tr><th>Task ID</th><th>Status</th><th>Created</th><th>Exec Time</th><th>Delay</th><th>Actions</th></tr></thead>
                  <tbody>
                    {tasks.map((t: Task) => (
                      <tr key={t.id}>
                        <td style={{ fontFamily: 'monospace', fontSize: 11 }}>{t.id.substring(0, 20)}...</td>
                        <td><span className={`tag ${t.status === 'COMPLETED' ? 'success' : t.status === 'FAILED' ? 'failed' : t.status === 'IN_PROGRESS' ? 'running' : 'pending'}`}>{t.status}</span></td>
                        <td style={{ fontSize: 12 }}>{t.createdAt ? new Date(t.createdAt).toLocaleString() : '-'}</td>
                        <td style={{ fontSize: 12 }}>{t.executionTime ? `${(t.executionTime/1000).toFixed(2)}s` : '-'}</td>
                        <td style={{ fontSize: 12 }}>{t.delayTime ? `${(t.delayTime/1000).toFixed(2)}s` : '-'}</td>
                        <td><button className="btn btn-sm btn-outline" onClick={(e) => { e.stopPropagation(); setSelectedTask(t); setTaskDetailOpen(true) }}><EyeOutlined /> View</button></td>
                      </tr>
                    ))}
                    {tasks.length === 0 && <tr><td colSpan={6} style={{ textAlign: 'center', color: 'var(--text-secondary)', padding: 20 }}>No tasks</td></tr>}
                  </tbody>
                </table>
              )}
            </div>
          )},
          { key: 'logs', label: '📄 Logs', children: (
            <div>
              <button className="btn btn-outline mb-3" onClick={() => { if (selectedWorker?.pod_name) { setLoadingLogs(true); getWorkerLogs(endpointName, selectedWorker.pod_name).then(setLogs).finally(() => setLoadingLogs(false)) } }}><SyncOutlined spin={loadingLogs} /> Refresh</button>
              <pre style={{ background: '#1f2937', color: '#e5e7eb', padding: 16, borderRadius: 8, fontSize: 12, maxHeight: 500, overflow: 'auto' }}>{logs || 'No logs'}</pre>
            </div>
          )},
          { key: 'exec', label: '💻 Exec', children: selectedWorker ? (
            <div style={{ height: 500 }}><Terminal endpoint={endpointName} workerId={selectedWorker.pod_name || selectedWorker.worker_id || selectedWorker.id} /></div>
          ) : <div>No worker selected</div> },
        ]} />
      </Drawer>

      <Modal title="Task Details" open={taskDetailOpen} onCancel={() => setTaskDetailOpen(false)} footer={null} width={900}>
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
  const [tasks, setTasks] = useState<Task[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [selectedTask, setSelectedTask] = useState<Task | null>(null)
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
  const handleCancel = async (taskId: string) => {
    try {
      await cancelTask(taskId)
      message.success('Task cancelled')
      fetchTasks()
    } catch (e: any) {
      message.error(e.response?.data?.error || 'Failed to cancel')
    }
  }
  const getStatusClass = (s: string) => s === 'COMPLETED' ? 'success' : s === 'FAILED' ? 'failed' : s === 'IN_PROGRESS' ? 'running' : 'pending'

  return (
    <>
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
          <button className="btn btn-blue" onClick={handleSearch}><SearchOutlined /> Search</button>
          <button className="btn btn-outline" onClick={() => { setSearchInput(''); setStatusInput('all'); setSearch(''); setStatusFilter('all'); setPage(0) }}>Reset</button>
        </div>
        <span style={{ color: 'var(--text-secondary)', fontSize: 13 }}>Total: {total}</span>
      </div>

      <div className="card">
        <div className="table-container">
          <table>
            <thead><tr><th>Task ID</th><th>Status</th><th>Worker</th><th>Created</th><th>Exec Time</th><th>Delay</th><th>Actions</th></tr></thead>
            <tbody>
              {loading ? <tr><td colSpan={7}><div className="loading"><div className="spinner"></div></div></td></tr> :
               tasks.length === 0 ? <tr><td colSpan={7} style={{ textAlign: 'center', color: 'var(--text-secondary)', padding: 40 }}>No tasks</td></tr> :
               tasks.map((t: Task) => (
                 <tr key={t.id}>
                   <td><Tooltip title={t.id}><span style={{ fontFamily: 'monospace', fontSize: 11 }}>{t.id.substring(0, 20)}...</span></Tooltip></td>
                   <td><span className={`tag ${getStatusClass(t.status)}`}>{t.status}</span></td>
                   <td>{t.workerId ? <Tooltip title={t.workerId}><span style={{ fontSize: 12, color: 'var(--text-secondary)' }}>{t.workerId.substring(0, 12)}...</span></Tooltip> : '-'}</td>
                   <td style={{ fontSize: 12 }}>{t.createdAt ? new Date(t.createdAt).toLocaleString() : '-'}</td>
                   <td style={{ fontSize: 12 }}>{t.executionTime ? `${(t.executionTime/1000).toFixed(2)}s` : '-'}</td>
                   <td style={{ fontSize: 12 }}>{t.delayTime ? `${(t.delayTime/1000).toFixed(2)}s` : '-'}</td>
                   <td>
                     <div className="flex gap-2">
                       <button className="btn btn-sm btn-outline" onClick={() => { setSelectedTask(t); setDetailOpen(true) }}><EyeOutlined /></button>
                       {(t.status === 'PENDING' || t.status === 'IN_PROGRESS') && (
                         <Popconfirm title="Cancel?" onConfirm={() => handleCancel(t.id)}>
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
        <div className="flex justify-between items-center" style={{ padding: '12px 16px', borderTop: '1px solid var(--bg-hover)' }}>
          <div className="flex gap-2 items-center">
            <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>Show</span>
            <Select value={pageSize} onChange={v => { setPageSize(v); setPage(0) }} size="small" style={{ width: 70 }} options={[
              { value: 10, label: '10' }, { value: 20, label: '20' }, { value: 50, label: '50' }, { value: 100, label: '100' },
            ]} />
            <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>/ page • Total: {total}</span>
          </div>
          <div className="flex gap-2 items-center">
            <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>Page {page + 1} of {Math.ceil(total / pageSize) || 1}</span>
            <button className="btn btn-sm btn-outline" disabled={page === 0} onClick={() => setPage(p => p - 1)}>Prev</button>
            <button className="btn btn-sm btn-outline" disabled={(page + 1) * pageSize >= total} onClick={() => setPage(p => p + 1)}>Next</button>
          </div>
        </div>
      </div>

      <Modal title="Task Details" open={detailOpen} onCancel={() => setDetailOpen(false)} footer={null} width={900}>
        {selectedTask && <TaskDetailContent task={selectedTask} />}
      </Modal>
    </>
  )
}

function TaskDetailContent({ task }: { task: Task }) {
  const [fullTask, setFullTask] = useState<Task | null>(null)
  const [timeline, setTimeline] = useState<any>(null)
  const [execHistory, setExecHistory] = useState<any>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!task?.id) return
    setLoading(true)
    Promise.all([
      getTaskStatus(task.id).catch(() => null),
      getTaskTimeline(task.id).catch(() => null),
      getTaskExecutionHistory(task.id).catch(() => null),
    ]).then(([t, tl, eh]) => {
      setFullTask(t)
      setTimeline(tl)
      setExecHistory(eh)
    }).finally(() => setLoading(false))
  }, [task?.id])

  const t = fullTask || task
  const getStatusClass = (s: string) => s === 'COMPLETED' ? 'success' : s === 'FAILED' ? 'failed' : s === 'IN_PROGRESS' ? 'running' : 'pending'

  if (loading) return <div className="loading"><div className="spinner"></div></div>

  return (
    <div>
      <div className="form-row">
        <div className="form-group"><label className="form-label">Task ID</label><div style={{ fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-all' }}>{t.id}</div></div>
        <div className="form-group"><label className="form-label">Status</label><span className={`tag ${getStatusClass(t.status)}`}>{t.status}</span></div>
      </div>
      <div className="form-row">
        <div className="form-group"><label className="form-label">Endpoint</label><div>{t.endpoint}</div></div>
        <div className="form-group"><label className="form-label">Worker</label><div style={{ fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-all' }}>{t.workerId || '-'}</div></div>
      </div>
      <div className="form-row">
        <div className="form-group"><label className="form-label">Created</label><div>{t.createdAt ? new Date(t.createdAt).toLocaleString() : '-'}</div></div>
        <div className="form-group"><label className="form-label">Execution Time</label><div>{t.executionTime ? `${(t.executionTime/1000).toFixed(2)}s` : '-'}</div></div>
      </div>
      <div className="form-row">
        <div className="form-group"><label className="form-label">Delay Time</label><div>{t.delayTime ? `${(t.delayTime/1000).toFixed(2)}s` : '-'}</div></div>
      </div>

      <div style={{ marginTop: 16 }}>
        <label className="form-label">Input</label>
        <pre style={{ background: 'var(--bg-primary)', padding: 12, borderRadius: 6, fontSize: 12, overflow: 'auto', maxHeight: 200, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{t.input ? JSON.stringify(t.input, null, 2) : '-'}</pre>
      </div>

      {t.output && (
        <div style={{ marginTop: 16 }}>
          <label className="form-label">Output</label>
          <pre style={{ background: 'var(--bg-primary)', padding: 12, borderRadius: 6, fontSize: 12, overflow: 'auto', maxHeight: 300, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{JSON.stringify(t.output, null, 2)}</pre>
        </div>
      )}

      {t.error && (
        <div style={{ marginTop: 16 }}>
          <label className="form-label">Error</label>
          <pre style={{ background: '#fef2f2', padding: 12, borderRadius: 6, fontSize: 12, color: '#f56565', whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{t.error}</pre>
        </div>
      )}

      {timeline?.timeline && timeline.timeline.length > 0 && (
        <div style={{ marginTop: 20 }}>
          <label className="form-label">Timeline</label>
          <Timeline style={{ marginTop: 12 }} items={timeline.timeline.map((e: any) => ({
            color: e.to_status === 'COMPLETED' ? 'green' : e.to_status === 'FAILED' ? 'red' : e.to_status === 'IN_PROGRESS' ? 'blue' : 'gray',
            children: (
              <div>
                <div><strong>{e.event_type}</strong>{e.from_status && e.to_status && <span style={{ color: 'var(--text-secondary)' }}> ({e.from_status} → {e.to_status})</span>}</div>
                <div style={{ fontSize: 12, color: 'var(--text-secondary)' }}>{new Date(e.event_time).toLocaleString()}</div>
                {e.worker_id && <div style={{ fontSize: 11 }}><code>Worker: {e.worker_id.substring(0, 20)}...</code></div>}
                {e.worker_pod_name && <div style={{ fontSize: 11 }}><code>Pod: {e.worker_pod_name}</code></div>}
                {e.error_message && <div style={{ fontSize: 12, color: '#f56565' }}>{e.error_message}</div>}
              </div>
            ),
          }))} />
        </div>
      )}

      {execHistory?.history && execHistory.history.length > 0 && (
        <div style={{ marginTop: 16 }}>
          <label className="form-label">Execution History</label>
          <div style={{ background: 'var(--bg-primary)', padding: 12, borderRadius: 6, marginTop: 8 }}>
            <table style={{ width: '100%', fontSize: 12 }}>
              <thead><tr style={{ color: 'var(--text-secondary)' }}><th style={{ textAlign: 'left', padding: 4 }}>Worker</th><th style={{ textAlign: 'left', padding: 4 }}>Start</th><th style={{ textAlign: 'left', padding: 4 }}>End</th><th style={{ textAlign: 'left', padding: 4 }}>Duration</th></tr></thead>
              <tbody>
                {execHistory.history.map((h: any, i: number) => (
                  <tr key={i}>
                    <td style={{ padding: 4, fontFamily: 'monospace' }}>{h.worker_id?.substring(0, 16) || '-'}</td>
                    <td style={{ padding: 4 }}>{h.start_time ? new Date(h.start_time).toLocaleString() : '-'}</td>
                    <td style={{ padding: 4 }}>{h.end_time ? new Date(h.end_time).toLocaleString() : '-'}</td>
                    <td style={{ padding: 4 }}>{h.duration_seconds ? `${h.duration_seconds.toFixed(2)}s` : '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}

function SettingsTab({ endpoint, onRefresh }: { endpoint: EndpointData; onRefresh: () => void }) {
  return (
    <Collapse defaultActiveKey={['base']} accordion items={[
      { key: 'base', label: '⚙️ Base Configuration', children: <BaseConfigPanel endpoint={endpoint} onRefresh={onRefresh} /> },
      { key: 'scaling', label: '⚡ Scaling Configuration', children: <ScalingPanel endpoint={endpoint} onRefresh={onRefresh} /> },
      { key: 'env', label: '🔧 Environment Variables', children: <EnvVarsPanel endpoint={endpoint} onRefresh={onRefresh} /> },
    ]} />
  )
}

function BaseConfigPanel({ endpoint, onRefresh }: { endpoint: EndpointData; onRefresh: () => void }) {
  const [form] = Form.useForm()
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    form.setFieldsValue({
      replicas: endpoint.replicas || 0,
      image: endpoint.image || '',
      taskTimeout: endpoint.taskTimeout || 3600,
    })
  }, [endpoint, form])

  const handleSave = async () => {
    setSaving(true)
    try {
      const values = form.getFieldsValue()
      await updateEndpoint(endpoint.logical_name!, {
        replicas: values.replicas,
        image: values.image,
      })
      if (values.taskTimeout !== endpoint.taskTimeout) {
        await updateEndpointConfig(endpoint.logical_name!, { taskTimeout: values.taskTimeout })
      }
      message.success('Saved')
      onRefresh()
    } catch { message.error('Failed to save') }
    finally { setSaving(false) }
  }

  return (
    <Form form={form} layout="vertical">
      <div className="form-row">
        <Form.Item name="replicas" label="Replicas"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
        <Form.Item label="Spec"><Input value={endpoint.specName} disabled /></Form.Item>
      </div>
      <Form.Item name="image" label="Image"><Input /></Form.Item>
      <Form.Item name="taskTimeout" label="Task Timeout (seconds)"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
      <button type="button" className="btn btn-blue" onClick={handleSave} disabled={saving}>{saving ? 'Saving...' : 'Save'}</button>
    </Form>
  )
}

function ScalingPanel({ endpoint, onRefresh }: { endpoint: EndpointData; onRefresh: () => void }) {
  const [form] = Form.useForm()
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    form.setFieldsValue({
      minReplicas: endpoint.minReplicas || 0,
      maxReplicas: endpoint.maxReplicas || 10,
    })
  }, [endpoint, form])

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
      <div className="form-row">
        <Form.Item name="minReplicas" label="Min Replicas"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
        <Form.Item name="maxReplicas" label="Max Replicas"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
      </div>
      <button type="button" className="btn btn-blue" onClick={handleSave} disabled={saving}>{saving ? 'Saving...' : 'Save'}</button>
    </Form>
  )
}

function EnvVarsPanel({ endpoint, onRefresh }: { endpoint: EndpointData; onRefresh: () => void }) {
  const [envVars, setEnvVars] = useState<{ key: string; value: string }[]>(() => {
    const env = endpoint.env || {}
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
      <div><button type="button" className="btn btn-blue" onClick={handleSave} disabled={saving}>{saving ? 'Saving...' : 'Save'}</button></div>
    </div>
  )
}

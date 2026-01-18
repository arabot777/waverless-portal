import { useEffect, useState, useRef } from 'react'
import { Link } from 'react-router-dom'
import { CloudServerOutlined, TeamOutlined, UnorderedListOutlined, ThunderboltOutlined, ArrowUpOutlined } from '@ant-design/icons'
import * as echarts from 'echarts'
import { getEndpoints, getAllTasks } from '../api/client'

export default function Dashboard() {
  const [endpoints, setEndpoints] = useState<any[]>([])
  const [stats, setStats] = useState({ completed: 0, in_progress: 0, pending: 0, failed: 0 })
  const [loading, setLoading] = useState(true)
  const statusChartRef = useRef<HTMLDivElement>(null)
  const endpointChartRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [endpointsRes, tasksRes] = await Promise.all([
          getEndpoints().catch(() => ({ endpoints: [] })),
          getAllTasks({ limit: 1000 }).catch(() => ({ tasks: [] })),
        ])
        const eps = endpointsRes.endpoints || []
        setEndpoints(eps)
        
        // Calculate task stats
        const tasks = tasksRes.tasks || []
        const taskStats = { completed: 0, in_progress: 0, pending: 0, failed: 0 }
        tasks.forEach((t: any) => {
          if (t.status === 'COMPLETED') taskStats.completed++
          else if (t.status === 'IN_PROGRESS') taskStats.in_progress++
          else if (t.status === 'PENDING') taskStats.pending++
          else if (t.status === 'FAILED') taskStats.failed++
        })
        setStats(taskStats)
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [])

  // Task Status Pie Chart
  useEffect(() => {
    if (!statusChartRef.current || loading) return
    const chart = echarts.init(statusChartRef.current)
    chart.setOption({
      tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
      series: [{
        type: 'pie',
        radius: ['50%', '70%'],
        center: ['50%', '50%'],
        data: [
          { value: stats.completed, name: 'Completed', itemStyle: { color: '#48bb78' } },
          { value: stats.in_progress, name: 'Running', itemStyle: { color: '#1da1f2' } },
          { value: stats.pending, name: 'Pending', itemStyle: { color: '#f59e0b' } },
          { value: stats.failed, name: 'Failed', itemStyle: { color: '#f56565' } },
        ].filter(d => d.value > 0),
        label: { show: false },
      }],
    })
    return () => chart.dispose()
  }, [stats, loading])

  // Endpoint Workers Bar Chart
  useEffect(() => {
    if (!endpointChartRef.current || !endpoints.length) return
    const chart = echarts.init(endpointChartRef.current)
    const topEndpoints = endpoints.slice(0, 7)
    chart.setOption({
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
      grid: { left: 120, right: 20, top: 20, bottom: 30 },
      xAxis: { type: 'value', axisLabel: { color: '#6b7280' } },
      yAxis: { 
        type: 'category', 
        data: topEndpoints.map((e: any) => {
          const name = e.logical_name || e.name || ''
          return name.length > 18 ? name.substring(0, 18) + '...' : name
        }), 
        axisLabel: { color: '#6b7280' } 
      },
      series: [
        { name: 'Workers', type: 'bar', data: topEndpoints.map((e: any) => e.ready_replicas || 0), itemStyle: { color: '#1da1f2' } },
      ],
    })
    return () => chart.dispose()
  }, [endpoints])

  const totalEndpoints = endpoints.length
  const runningEndpoints = endpoints.filter((e: any) => e.status === 'Running' || e.status === 'running').length
  const totalWorkers = endpoints.reduce((sum: number, e: any) => sum + (e.ready_replicas || 0), 0)

  if (loading) return <div className="loading"><div className="spinner"></div></div>

  return (
    <>
      {/* Stats Grid */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-icon blue"><CloudServerOutlined /></div>
          <div className="stat-content">
            <div className="stat-label">Total Endpoints</div>
            <div className="stat-value">{totalEndpoints}</div>
            <div className="stat-change positive"><ArrowUpOutlined /> {runningEndpoints} running</div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon green"><TeamOutlined /></div>
          <div className="stat-content">
            <div className="stat-label">Active Workers</div>
            <div className="stat-value">{totalWorkers}</div>
            <div className="stat-change positive"><ArrowUpOutlined /> online</div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon purple"><UnorderedListOutlined /></div>
          <div className="stat-content">
            <div className="stat-label">Completed Tasks</div>
            <div className="stat-value">{stats.completed}</div>
            <div className="stat-change positive"><ArrowUpOutlined /> completed</div>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon orange"><ThunderboltOutlined /></div>
          <div className="stat-content">
            <div className="stat-label">Pending Tasks</div>
            <div className="stat-value">{stats.pending}</div>
            <div className="stat-change">{stats.in_progress} in progress</div>
          </div>
        </div>
      </div>

      {/* Charts */}
      <div className="charts-row">
        <div className="card">
          <div className="card-header">
            <h3>Workers by Endpoint</h3>
            <span className="subtitle">Top endpoints by worker count</span>
          </div>
          <div className="chart-container" ref={endpointChartRef}></div>
        </div>
        <div className="card">
          <div className="card-header"><h3>Task Status</h3></div>
          <div className="chart-container" ref={statusChartRef}></div>
        </div>
      </div>

      {/* Lists */}
      <div className="lists-row">
        <div className="card">
          <div className="card-header">
            <h3>Recent Endpoints</h3>
            <Link to="/endpoints" className="view-all">View All →</Link>
          </div>
          <div className="list-content">
            {endpoints.slice(0, 5).map((ep: any) => (
              <Link to={`/endpoints/${ep.logical_name || ep.name}`} key={ep.logical_name || ep.name} className="list-item" style={{ textDecoration: 'none' }}>
                <div className="item-icon gpu"><ThunderboltOutlined /></div>
                <div className="item-info">
                  <div className="item-name">{ep.logical_name || ep.name}</div>
                  <div className="item-meta">{ep.spec_name || ep.specName} • {ep.ready_replicas || 0} workers</div>
                </div>
                <span className={`tag ${ep.status === 'Running' || ep.status === 'running' ? 'running' : ep.status === 'Pending' || ep.status === 'pending' ? 'pending' : 'stopped'}`}>
                  {ep.status}
                </span>
              </Link>
            ))}
            {endpoints.length === 0 && (
              <div className="empty-state"><p>No endpoints yet</p></div>
            )}
          </div>
        </div>
        <div className="card">
          <div className="card-header"><h3>Quick Actions</h3></div>
          <div className="card-body">
            <Link to="/endpoints" className="btn btn-blue" style={{ width: '100%', justifyContent: 'center', marginBottom: 12 }}>
              <ThunderboltOutlined /> Create New Endpoint
            </Link>
            <Link to="/tasks" className="btn btn-outline" style={{ width: '100%', justifyContent: 'center' }}>
              <UnorderedListOutlined /> View All Tasks
            </Link>
          </div>
        </div>
      </div>
    </>
  )
}

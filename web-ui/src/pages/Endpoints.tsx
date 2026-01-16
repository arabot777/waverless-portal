import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { message, Popconfirm } from 'antd'
import { DeleteOutlined, PauseCircleOutlined, SearchOutlined } from '@ant-design/icons'
import { getEndpoints, deleteEndpoint, updateEndpoint } from '../api/client'

interface Endpoint {
  id: number
  logical_name: string
  spec_name: string
  spec_type: string
  cluster_id: string
  status: string
  replicas: number
  current_replicas: number
  price_per_hour: number
}

export default function Endpoints() {
  const navigate = useNavigate()
  const [endpoints, setEndpoints] = useState<Endpoint[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')

  const fetchEndpoints = async () => {
    try {
      const res = await getEndpoints()
      setEndpoints(res.endpoints || [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchEndpoints()
    const interval = setInterval(fetchEndpoints, 10000)
    return () => clearInterval(interval)
  }, [])

  const handleStop = async (name: string) => {
    try {
      await updateEndpoint(name, { replicas: 0 })
      message.success('Endpoint stopped')
      fetchEndpoints()
    } catch {
      message.error('Failed to stop endpoint')
    }
  }

  const handleDelete = async (name: string) => {
    try {
      await deleteEndpoint(name)
      message.success('Endpoint deleted')
      fetchEndpoints()
    } catch {
      message.error('Failed to delete endpoint')
    }
  }

  const filteredEndpoints = endpoints.filter(ep =>
    ep.logical_name.toLowerCase().includes(search.toLowerCase())
  )

  if (loading) return <div className="loading"><div className="spinner"></div></div>

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <div style={{ position: 'relative' }}>
          <SearchOutlined style={{ position: 'absolute', left: 12, top: 10, color: 'var(--text-muted)' }} />
          <input
            type="text"
            placeholder="Search endpoints..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            style={{
              padding: '8px 12px 8px 36px',
              background: 'var(--bg-hover)',
              border: '1px solid var(--border-color)',
              borderRadius: 6,
              color: 'var(--text-primary)',
              width: 240,
            }}
          />
        </div>
        <a href="/serverless" className="btn btn-primary">Create Endpoint</a>
      </div>

      <div className="card">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Spec</th>
              <th>Cluster</th>
              <th>Replicas</th>
              <th>Price/hr</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredEndpoints.length === 0 ? (
              <tr><td colSpan={7} className="empty-state">No endpoints</td></tr>
            ) : (
              filteredEndpoints.map(ep => (
                <tr key={ep.id}>
                  <td>
                    <a
                      onClick={() => navigate(`/endpoints/${ep.logical_name}`)}
                      style={{ color: 'var(--primary)', cursor: 'pointer', fontWeight: 500 }}
                    >
                      {ep.logical_name}
                    </a>
                  </td>
                  <td><span className={`tag ${ep.spec_type.toLowerCase()}`}>{ep.spec_name}</span></td>
                  <td>{ep.cluster_id}</td>
                  <td>{ep.current_replicas ?? 0}/{ep.replicas ?? 0}</td>
                  <td>${ep.price_per_hour?.toFixed(2)}</td>
                  <td><span className={`tag ${ep.status}`}>{ep.status}</span></td>
                  <td>
                    <div style={{ display: 'flex', gap: 8 }}>
                      <Popconfirm title="Stop this endpoint?" onConfirm={() => handleStop(ep.logical_name)}>
                        <button className="btn btn-outline" title="Stop">
                          <PauseCircleOutlined />
                        </button>
                      </Popconfirm>
                      <Popconfirm title="Delete this endpoint?" onConfirm={() => handleDelete(ep.logical_name)}>
                        <button className="btn btn-outline" title="Delete" style={{ color: 'var(--danger)' }}>
                          <DeleteOutlined />
                        </button>
                      </Popconfirm>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}

import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Modal, Form, Input, InputNumber, Select, Switch, message } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, ArrowLeftOutlined } from '@ant-design/icons'
import { getCluster, getClusterSpecs, createClusterSpec, updateClusterSpec, deleteClusterSpec, getAdminSpecs } from '../api/client'

interface ClusterSpec {
  id: number
  cluster_id: string
  cluster_spec_name: string
  spec_name: string
  total_capacity: number
  available_capacity: number
  is_available: boolean
}

interface Cluster {
  cluster_id: string
  cluster_name: string
  region: string
}

interface GlobalSpec {
  spec_name: string
  spec_type: string
  gpu_type: string
  gpu_count: number
  cpu_cores: number
  ram_gb: number
  default_price_per_hour: number
}

export default function ClusterDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [cluster, setCluster] = useState<Cluster | null>(null)
  const [specs, setSpecs] = useState<ClusterSpec[]>([])
  const [globalSpecs, setGlobalSpecs] = useState<GlobalSpec[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<ClusterSpec | null>(null)
  const [selectedSpec, setSelectedSpec] = useState<GlobalSpec | null>(null)
  const [form] = Form.useForm()

  const fetchData = async () => {
    try {
      const [clusterRes, specsRes, globalRes] = await Promise.all([
        getCluster(id!),
        getClusterSpecs(id!),
        getAdminSpecs()
      ])
      setCluster(clusterRes)
      setSpecs(specsRes.specs || [])
      setGlobalSpecs(globalRes.specs || [])
    } catch {
      message.error('Failed to load cluster')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchData() }, [id])

  const getSpecInfo = (specName: string) => globalSpecs.find(s => s.spec_name === specName)

  const openModal = (spec?: ClusterSpec) => {
    setEditing(spec || null)
    if (spec) {
      form.setFieldsValue(spec)
      setSelectedSpec(getSpecInfo(spec.spec_name) || null)
    } else {
      form.resetFields()
      form.setFieldsValue({ is_available: true, total_capacity: 10, available_capacity: 10 })
      setSelectedSpec(null)
    }
    setModalOpen(true)
  }

  const onSpecSelect = (specName: string) => {
    setSelectedSpec(getSpecInfo(specName) || null)
  }

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      if (editing) {
        await updateClusterSpec(id!, { id: editing.id, ...values })
        message.success('Spec updated')
      } else {
        await createClusterSpec(id!, values)
        message.success('Spec created')
      }
      setModalOpen(false)
      fetchData()
    } catch {
      message.error('Failed to save spec')
    }
  }

  const handleDelete = async (specId: number) => {
    if (!confirm('Delete this spec?')) return
    try {
      await deleteClusterSpec(id!, specId)
      message.success('Spec deleted')
      fetchData()
    } catch {
      message.error('Failed to delete spec')
    }
  }

  if (loading) return <div className="loading"><div className="spinner"></div></div>
  if (!cluster) return <div>Cluster not found</div>

  return (
    <div>
      <div className="flex items-center gap-4 mb-4">
        <button className="btn btn-outline" onClick={() => navigate('/clusters')}><ArrowLeftOutlined /></button>
        <div>
          <h2 style={{ margin: 0 }}>{cluster.cluster_name}</h2>
          <span style={{ color: 'var(--text-secondary)' }}>{cluster.cluster_id} · {cluster.region}</span>
        </div>
        <div style={{ marginLeft: 'auto' }}>
          <button className="btn btn-primary" onClick={() => openModal()}><PlusOutlined /> Add Spec</button>
        </div>
      </div>

      <div className="card">
        <table>
          <thead>
            <tr>
              <th>Cluster Spec</th>
              <th>Portal Spec</th>
              <th>Type</th>
              <th>GPU</th>
              <th>CPU/RAM</th>
              <th>Price/hr</th>
              <th>Capacity</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {specs.length === 0 ? (
              <tr><td colSpan={9} className="empty-state">No specs configured for this cluster</td></tr>
            ) : (
              specs.map(s => {
                const info = getSpecInfo(s.spec_name)
                return (
                  <tr key={s.id}>
                    <td style={{ fontFamily: 'monospace' }}>{s.cluster_spec_name}</td>
                    <td style={{ fontFamily: 'monospace' }}>{s.spec_name}</td>
                    <td><span className={`tag ${info?.spec_type?.toLowerCase() || ''}`}>{info?.spec_type || '-'}</span></td>
                    <td>{info?.spec_type === 'GPU' ? `${info.gpu_count}x ${info.gpu_type}` : '-'}</td>
                    <td>{info ? `${info.cpu_cores}C / ${info.ram_gb}GB` : '-'}</td>
                    <td>{info ? `$${info.default_price_per_hour.toFixed(2)}` : '-'}</td>
                    <td>{s.available_capacity}/{s.total_capacity}</td>
                    <td><span className={`tag ${s.is_available ? 'active' : 'offline'}`}>{s.is_available ? 'Available' : 'Disabled'}</span></td>
                    <td>
                      <div style={{ display: 'flex', gap: 8 }}>
                        <button className="btn btn-outline" onClick={() => openModal(s)}><EditOutlined /></button>
                        <button className="btn btn-outline" style={{ color: 'var(--danger)' }} onClick={() => handleDelete(s.id)}><DeleteOutlined /></button>
                      </div>
                    </td>
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>

      <Modal title={editing ? 'Edit Cluster Spec' : 'Add Cluster Spec'} open={modalOpen} onOk={handleSave} onCancel={() => setModalOpen(false)} okText="Save" width={500}>
        <Form form={form} layout="vertical" style={{ marginTop: 20 }}>
          <Form.Item name="cluster_spec_name" label="Cluster Spec Name" rules={[{ required: true }]} tooltip="The spec name used in the cluster">
            <Input placeholder="e.g. 4090-1" disabled={!!editing} />
          </Form.Item>
          <Form.Item name="spec_name" label="Portal Spec" rules={[{ required: true }]} tooltip="Maps to global spec pricing">
            <Select placeholder="Select a spec" onChange={onSpecSelect} disabled={!!editing}>
              {globalSpecs.map(s => (
                <Select.Option key={s.spec_name} value={s.spec_name}>{s.spec_name}</Select.Option>
              ))}
            </Select>
          </Form.Item>
          {selectedSpec && (
            <div style={{ background: 'var(--bg-primary)', padding: 12, borderRadius: 6, marginBottom: 16 }}>
              <div style={{ color: 'var(--text-secondary)', fontSize: 12 }}>Spec Info (from global config)</div>
              <div style={{ marginTop: 8 }}>
                Type: <strong>{selectedSpec.spec_type}</strong> · 
                {selectedSpec.spec_type === 'GPU' && <> GPU: <strong>{selectedSpec.gpu_count}x {selectedSpec.gpu_type}</strong> · </>}
                CPU: <strong>{selectedSpec.cpu_cores}</strong> · 
                RAM: <strong>{selectedSpec.ram_gb}GB</strong> · 
                Price: <strong>${selectedSpec.default_price_per_hour.toFixed(2)}/hr</strong>
              </div>
            </div>
          )}
          <div style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="total_capacity" label="Total Capacity" rules={[{ required: true }]} style={{ flex: 1 }}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="available_capacity" label="Available Capacity" rules={[{ required: true }]} style={{ flex: 1 }}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </div>
          <Form.Item name="is_available" label="Available" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

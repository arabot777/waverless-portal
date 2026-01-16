import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Modal, Form, Input, InputNumber, Select, message } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, SettingOutlined } from '@ant-design/icons'
import { getClusters, createCluster, updateCluster, deleteCluster } from '../api/client'

interface Cluster {
  cluster_id: string
  cluster_name: string
  region: string
  api_endpoint: string
  api_key: string
  status: string
  priority: number
  total_gpu_slots: number
  available_gpu_slots: number
}

export default function Clusters() {
  const navigate = useNavigate()
  const [clusters, setClusters] = useState<Cluster[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Cluster | null>(null)
  const [form] = Form.useForm()

  const fetchClusters = async () => {
    try {
      const res = await getClusters()
      setClusters(res.clusters || [])
    } catch {
      setClusters([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchClusters() }, [])

  const openModal = (cluster?: Cluster) => {
    setEditing(cluster || null)
    form.setFieldsValue(cluster || { status: 'active', priority: 100 })
    setModalOpen(true)
  }

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      if (editing) {
        await updateCluster(editing.cluster_id, values)
        message.success('Cluster updated')
      } else {
        await createCluster(values)
        message.success('Cluster created')
      }
      setModalOpen(false)
      form.resetFields()
      fetchClusters()
    } catch {
      message.error('Failed to save cluster')
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Delete this cluster?')) return
    try {
      await deleteCluster(id)
      message.success('Cluster deleted')
      fetchClusters()
    } catch {
      message.error('Failed to delete cluster')
    }
  }

  if (loading) return <div className="loading"><div className="spinner"></div></div>

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <div></div>
        <button className="btn btn-primary" onClick={() => openModal()}>
          <PlusOutlined /> Add Cluster
        </button>
      </div>

      <div className="card">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>Name</th>
              <th>Region</th>
              <th>API Endpoint</th>
              <th>Status</th>
              <th>Priority</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {clusters.length === 0 ? (
              <tr><td colSpan={7} className="empty-state">No clusters configured</td></tr>
            ) : (
              clusters.map(c => (
                <tr key={c.cluster_id}>
                  <td style={{ fontFamily: 'monospace' }}>{c.cluster_id}</td>
                  <td>{c.cluster_name}</td>
                  <td>{c.region}</td>
                  <td style={{ maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis' }}>{c.api_endpoint}</td>
                  <td><span className={`tag ${c.status}`}>{c.status}</span></td>
                  <td>{c.priority}</td>
                  <td>
                    <div style={{ display: 'flex', gap: 8 }}>
                      <button className="btn btn-outline" onClick={() => navigate(`/clusters/${c.cluster_id}`)} title="Manage Specs"><SettingOutlined /></button>
                      <button className="btn btn-outline" onClick={() => openModal(c)}><EditOutlined /></button>
                      <button className="btn btn-outline" style={{ color: 'var(--danger)' }} onClick={() => handleDelete(c.cluster_id)}><DeleteOutlined /></button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <Modal title={editing ? 'Edit Cluster' : 'Add Cluster'} open={modalOpen} onOk={handleSave} onCancel={() => setModalOpen(false)} okText="Save" width={500}>
        <Form form={form} layout="vertical" style={{ marginTop: 20 }}>
          <Form.Item name="cluster_id" label="Cluster ID" rules={[{ required: true }]}>
            <Input placeholder="cluster-1" disabled={!!editing} />
          </Form.Item>
          <Form.Item name="cluster_name" label="Cluster Name" rules={[{ required: true }]}>
            <Input placeholder="US East Cluster" />
          </Form.Item>
          <Form.Item name="region" label="Region" rules={[{ required: true }]}>
            <Input placeholder="us-east-1" />
          </Form.Item>
          <Form.Item name="api_endpoint" label="API Endpoint" rules={[{ required: true }]}>
            <Input placeholder="http://waverless-api:8080" />
          </Form.Item>
          <Form.Item name="api_key" label="API Key">
            <Input placeholder="Enter API key" />
          </Form.Item>
          <div style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="status" label="Status" style={{ flex: 1 }}>
              <Select>
                <Select.Option value="active">Active</Select.Option>
                <Select.Option value="maintenance">Maintenance</Select.Option>
                <Select.Option value="offline">Offline</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item name="priority" label="Priority" style={{ flex: 1 }}>
              <InputNumber min={0} max={1000} style={{ width: '100%' }} />
            </Form.Item>
          </div>
        </Form>
      </Modal>
    </div>
  )
}

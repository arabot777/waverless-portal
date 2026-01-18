import { useState, useEffect } from 'react'
import { Modal, Form, Input, InputNumber, Select, Switch, message } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { getAdminSpecs, createSpec, updateSpec, deleteSpec } from '../api/client'

interface Spec {
  id: number
  spec_name: string
  spec_type: string
  gpu_type: string
  gpu_count: number
  cpu_cores: number
  ram_gb: number
  disk_gb: number
  price_per_hour: number
  description: string
  is_available: boolean
}

export default function Specs() {
  const [specs, setSpecs] = useState<Spec[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Spec | null>(null)
  const [form] = Form.useForm()

  const fetchSpecs = async () => {
    try {
      const res = await getAdminSpecs()
      setSpecs(res.specs || [])
    } catch {
      setSpecs([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchSpecs() }, [])

  const openModal = (spec?: Spec) => {
    setEditing(spec || null)
    form.setFieldsValue(spec || { spec_type: 'GPU', gpu_count: 1, cpu_cores: 8, ram_gb: 32, disk_gb: 100, is_available: true })
    setModalOpen(true)
  }

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      if (editing) {
        await updateSpec({ id: editing.id, ...values })
        message.success('Spec updated')
      } else {
        await createSpec(values)
        message.success('Spec created')
      }
      setModalOpen(false)
      form.resetFields()
      fetchSpecs()
    } catch {
      message.error('Failed to save spec')
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this spec?')) return
    try {
      await deleteSpec(id)
      message.success('Spec deleted')
      fetchSpecs()
    } catch {
      message.error('Failed to delete spec')
    }
  }

  if (loading) return <div className="loading"><div className="spinner"></div></div>

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <div></div>
        <button className="btn btn-primary" onClick={() => openModal()}>
          <PlusOutlined /> Add Spec
        </button>
      </div>

      <div className="card">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Type</th>
              <th>GPU</th>
              <th>CPU</th>
              <th>RAM</th>
              <th>Price/hr</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {specs.length === 0 ? (
              <tr><td colSpan={8} className="empty-state">No specs configured</td></tr>
            ) : (
              specs.map(s => (
                <tr key={s.id}>
                  <td style={{ fontFamily: 'monospace' }}>{s.spec_name}</td>
                  <td><span className={`tag ${s.spec_type.toLowerCase()}`}>{s.spec_type}</span></td>
                  <td>{s.spec_type === 'GPU' ? `${s.gpu_count}x ${s.gpu_type}` : '-'}</td>
                  <td>{s.cpu_cores} cores</td>
                  <td>{s.ram_gb} GB</td>
                  <td>${s.price_per_hour.toFixed(2)}</td>
                  <td><span className={`tag ${s.is_available ? 'active' : 'offline'}`}>{s.is_available ? 'Available' : 'Disabled'}</span></td>
                  <td>
                    <div style={{ display: 'flex', gap: 8 }}>
                      <button className="btn btn-outline" onClick={() => openModal(s)}><EditOutlined /></button>
                      <button className="btn btn-outline" style={{ color: 'var(--danger)' }} onClick={() => handleDelete(s.id)}><DeleteOutlined /></button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <Modal title={editing ? 'Edit Spec' : 'Add Spec'} open={modalOpen} onOk={handleSave} onCancel={() => setModalOpen(false)} okText="Save" width={600}>
        <Form form={form} layout="vertical" style={{ marginTop: 20 }}>
          <div style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="spec_name" label="Spec Name" rules={[{ required: true }]} style={{ flex: 2 }}>
              <Input placeholder="GPU-H100-80GB" />
            </Form.Item>
            <Form.Item name="spec_type" label="Type" rules={[{ required: true }]} style={{ flex: 1 }}>
              <Select>
                <Select.Option value="GPU">GPU</Select.Option>
                <Select.Option value="CPU">CPU</Select.Option>
              </Select>
            </Form.Item>
          </div>
          <div style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="gpu_type" label="GPU Type" style={{ flex: 2 }}>
              <Input placeholder="H100-80GB" />
            </Form.Item>
            <Form.Item name="gpu_count" label="GPU Count" style={{ flex: 1 }}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </div>
          <div style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="cpu_cores" label="CPU Cores" rules={[{ required: true }]} style={{ flex: 1 }}>
              <InputNumber min={1} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="ram_gb" label="RAM (GB)" rules={[{ required: true }]} style={{ flex: 1 }}>
              <InputNumber min={1} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="disk_gb" label="Disk (GB)" style={{ flex: 1 }}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </div>
          <div style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="price_per_hour" label="Price per Hour ($)" rules={[{ required: true }]} style={{ flex: 1 }}>
              <InputNumber min={0} step={0.01} precision={2} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="is_available" label="Available" valuePropName="checked" style={{ flex: 1 }}>
              <Switch />
            </Form.Item>
          </div>
          <Form.Item name="description" label="Description">
            <Input.TextArea rows={2} placeholder="NVIDIA H100 80GB GPU, ideal for large model training" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

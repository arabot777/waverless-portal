import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Modal, Form, Input, InputNumber, Button, message } from 'antd'
import { PlusOutlined, ThunderboltOutlined, DeleteOutlined } from '@ant-design/icons'
import { getSpecs, createEndpoint } from '../api/client'

interface Spec {
  spec_name: string
  spec_type: string
  gpu_type: string
  gpu_count: number
  cpu_cores: number
  ram_gb: number
  price_per_hour: number
  description: string
  available_capacity: number
  total_capacity: number
}

export default function Serverless() {
  const navigate = useNavigate()
  const [specs, setSpecs] = useState<Spec[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [selectedSpec, setSelectedSpec] = useState<Spec | null>(null)
  const [creating, setCreating] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    getSpecs()
      .then(res => setSpecs(res.specs || []))
      .finally(() => setLoading(false))
  }, [])

  const generateRandomName = () => {
    const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
    return Array.from({ length: 10 }, () => chars[Math.floor(Math.random() * chars.length)]).join('')
  }

  const openCreateModal = (spec: Spec) => {
    setSelectedSpec(spec)
    form.setFieldsValue({
      name: generateRandomName(),
      spec_name: spec.spec_name,
      replicas: 1,
      min_replicas: 0,
      max_replicas: 3,
      task_timeout: 3600,
      envVars: [{ key: '', value: '' }],
    })
    setModalOpen(true)
  }

  const handleCreate = async () => {
    try {
      const values = await form.validateFields()
      setCreating(true)
      
      // 转换 envVars 为 env 对象
      const env: Record<string, string> = {}
      const envVars = values.envVars || []
      envVars.filter((v: { key: string; value: string }) => v?.key && v?.value)
        .forEach((v: { key: string; value: string }) => { env[v.key] = v.value })

      await createEndpoint({
        logical_name: values.name,
        spec_name: values.spec_name,
        image: values.image,
        replicas: values.replicas,
        min_replicas: values.min_replicas,
        max_replicas: values.max_replicas,
        task_timeout: values.task_timeout,
        env: Object.keys(env).length > 0 ? env : undefined,
      })
      message.success('Endpoint created')
      setModalOpen(false)
      form.resetFields()
      navigate(`/endpoints/${values.name}`)
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } }
      message.error(error.response?.data?.error || 'Failed to create endpoint')
    } finally {
      setCreating(false)
    }
  }

  const gpuSpecs = specs.filter(s => s.spec_type === 'GPU')
  const cpuSpecs = specs.filter(s => s.spec_type === 'CPU')

  if (loading) return <div className="loading"><div className="spinner"></div></div>

  return (
    <>
      <div className="card mb-5">
        <div className="card-header">
          <h3>GPU Specs</h3>
          <span style={{ color: 'var(--text-secondary)', fontSize: 13, marginLeft: 8 }}>
            Select a spec to create endpoint
          </span>
        </div>
        <div className="card-body">
          <div className="specs-grid">
            {gpuSpecs.map(spec => (
              <div key={spec.spec_name} className="spec-card" onClick={() => openCreateModal(spec)} style={{ cursor: 'pointer' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                  <span style={{ fontSize: 16, fontWeight: 600 }}>{spec.gpu_count}× GPU</span>
                  <span style={{ color: 'var(--text-secondary)' }}>{spec.gpu_type}</span>
                </div>
                <div style={{ fontSize: 14, color: 'var(--text-secondary)', marginBottom: 12 }}>
                  {spec.description || spec.spec_name}
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', paddingTop: 12, borderTop: '1px solid var(--border-color)' }}>
                  <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>
                    {spec.cpu_cores} CPU • {spec.ram_gb}GB
                  </span>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                    <span style={{ color: 'var(--success)', fontWeight: 600 }}>${spec.price_per_hour}/hr</span>
                    <button className="btn btn-primary" onClick={(e) => { e.stopPropagation(); openCreateModal(spec) }}>
                      <PlusOutlined /> Create
                    </button>
                  </div>
                </div>
              </div>
            ))}
            {gpuSpecs.length === 0 && (
              <div className="empty-state">
                <ThunderboltOutlined style={{ fontSize: 48, opacity: 0.3 }} />
                <p>No GPU specs available</p>
              </div>
            )}
          </div>
        </div>
      </div>

      {cpuSpecs.length > 0 && (
        <div className="card">
          <div className="card-header"><h3>CPU Specs</h3></div>
          <div className="card-body">
            <div className="specs-grid">
              {cpuSpecs.map(spec => (
                <div key={spec.spec_name} className="spec-card" onClick={() => openCreateModal(spec)} style={{ cursor: 'pointer' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                    <span style={{ fontSize: 16, fontWeight: 600 }}>{spec.cpu_cores} CPU</span>
                    <span style={{ color: 'var(--text-secondary)' }}>{spec.ram_gb}GB RAM</span>
                  </div>
                  <div style={{ fontSize: 14, color: 'var(--text-secondary)', marginBottom: 12 }}>
                    {spec.description || spec.spec_name}
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', paddingTop: 12, borderTop: '1px solid var(--border-color)' }}>
                    <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>CPU Only</span>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                      <span style={{ color: 'var(--success)', fontWeight: 600 }}>${spec.price_per_hour}/hr</span>
                      <button className="btn btn-primary" onClick={(e) => { e.stopPropagation(); openCreateModal(spec) }}>
                        <PlusOutlined /> Create
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      <Modal
        title="Create Endpoint"
        open={modalOpen}
        onOk={handleCreate}
        onCancel={() => setModalOpen(false)}
        confirmLoading={creating}
        okText="Create"
        width={500}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 20 }}>
          <Form.Item
            name="name"
            label="Endpoint Name"
            rules={[
              { required: true, message: 'Please enter endpoint name' },
              { pattern: /^[a-zA-Z0-9-]+$/, message: 'Only letters, numbers and hyphens' }
            ]}
          >
            <Input placeholder="my-endpoint" />
          </Form.Item>

          <Form.Item
            name="image"
            label="Docker Image"
            rules={[{ required: true, message: 'Please enter docker image' }]}
          >
            <Input placeholder="your-registry/your-image:tag" />
          </Form.Item>

          <Form.Item name="spec_name" label="Hardware Spec">
            <Input disabled value={selectedSpec?.spec_name} />
          </Form.Item>

          <div style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="replicas" label="Replicas" style={{ flex: 1 }}>
              <InputNumber min={0} max={100} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="min_replicas" label="Min Replicas" style={{ flex: 1 }}>
              <InputNumber min={0} max={100} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="max_replicas" label="Max Replicas" style={{ flex: 1 }}>
              <InputNumber min={1} max={100} style={{ width: '100%' }} />
            </Form.Item>
          </div>
          <Form.Item name="task_timeout" label="Timeout (s)">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item label="Environment Variables">
            <Form.List name="envVars">
              {(fields, { add, remove }) => (
                <>
                  {fields.map((field) => (
                    <div key={field.key} style={{ display: 'flex', gap: 8, marginBottom: 8 }}>
                      <Form.Item {...field} name={[field.name, 'key']} style={{ flex: 1, marginBottom: 0 }}>
                        <Input placeholder="KEY" />
                      </Form.Item>
                      <Form.Item {...field} name={[field.name, 'value']} style={{ flex: 2, marginBottom: 0 }}>
                        <Input placeholder="value" />
                      </Form.Item>
                      <Button type="text" danger icon={<DeleteOutlined />} onClick={() => remove(field.name)} />
                    </div>
                  ))}
                  <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>Add Variable</Button>
                </>
              )}
            </Form.List>
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

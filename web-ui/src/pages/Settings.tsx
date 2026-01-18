import { useState, useEffect } from 'react'
import { Button, Modal, Form, Input, message, Popconfirm } from 'antd'
import { PlusOutlined, DeleteOutlined, KeyOutlined } from '@ant-design/icons'
import { getRegistryCredentials, createRegistryCredential, deleteRegistryCredential } from '../api/client'

interface Credential {
  id: number
  name: string
  registry: string
  username: string
  created_at: string
}

export default function Settings() {
  const [credentials, setCredentials] = useState<Credential[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [form] = Form.useForm()

  const fetchCredentials = async () => {
    setLoading(true)
    try {
      const data = await getRegistryCredentials()
      setCredentials(data.credentials || [])
    } catch {
      message.error('Failed to load credentials')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchCredentials() }, [])

  const handleCreate = async (values: { name: string; registry?: string; username: string; password: string }) => {
    try {
      await createRegistryCredential(values)
      message.success('Credential created')
      setModalOpen(false)
      form.resetFields()
      fetchCredentials()
    } catch {
      message.error('Failed to create credential')
    }
  }

  const handleDelete = async (name: string) => {
    try {
      await deleteRegistryCredential(name)
      message.success('Credential deleted')
      fetchCredentials()
    } catch {
      message.error('Failed to delete credential')
    }
  }

  return (
    <div className="page-container">
      <div className="card">
        <div className="card-header">
          <h3><KeyOutlined style={{ marginRight: 8 }} />Registry Credentials</h3>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>Add Credential</Button>
        </div>
        <div className="card-body">
          <p style={{ marginBottom: 16, color: 'var(--text-secondary)' }}>
            Store credentials for private container registries. These can be used when deploying endpoints with private images.
          </p>
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Registry</th>
                <th>Username</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr><td colSpan={5} style={{ textAlign: 'center', padding: 20 }}>Loading...</td></tr>
              ) : credentials.length === 0 ? (
                <tr><td colSpan={5} style={{ textAlign: 'center', padding: 20, color: 'var(--text-muted)' }}>No credentials</td></tr>
              ) : credentials.map(c => (
                <tr key={c.id}>
                  <td>{c.name}</td>
                  <td>{c.registry}</td>
                  <td>{c.username}</td>
                  <td>{new Date(c.created_at).toLocaleString()}</td>
                  <td>
                    <Popconfirm title="Delete this credential?" onConfirm={() => handleDelete(c.name)}>
                      <Button type="text" danger icon={<DeleteOutlined />} size="small" />
                    </Popconfirm>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <Modal title="Add Registry Credential" open={modalOpen} onCancel={() => setModalOpen(false)} onOk={() => form.submit()} okText="Create">
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input placeholder="my-dockerhub" />
          </Form.Item>
          <Form.Item name="registry" label="Registry" initialValue="docker.io">
            <Input placeholder="docker.io" />
          </Form.Item>
          <Form.Item name="username" label="Username" rules={[{ required: true }]}>
            <Input placeholder="username" />
          </Form.Item>
          <Form.Item name="password" label="Password" rules={[{ required: true }]}>
            <Input.Password placeholder="password or access token" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

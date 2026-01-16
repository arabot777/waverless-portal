import { useEffect, useState } from 'react'
import { getSpecs } from '../api/client'

interface Spec {
  spec_name: string
  spec_type: string
  gpu_type: string
  gpu_count: number
  cpu_cores: number
  ram_gb: number
  price_per_hour: number
  description: string
  total_capacity: number
  available_capacity: number
}

export default function Specs() {
  const [specs, setSpecs] = useState<Spec[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getSpecs()
      .then(res => setSpecs(res.specs || []))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="loading"><div className="spinner"></div></div>

  const gpuSpecs = specs.filter(s => s.spec_type === 'GPU')
  const cpuSpecs = specs.filter(s => s.spec_type === 'CPU')

  const renderSpecCard = (spec: Spec) => (
    <div key={spec.spec_name} className="spec-card">
      <div style={{ marginBottom: 8 }}>
        <span className={`tag ${spec.spec_type.toLowerCase()}`}>{spec.spec_type}</span>
      </div>
      <div style={{ fontSize: 16, fontWeight: 600, marginBottom: 4 }}>{spec.spec_name}</div>
      <div style={{ fontSize: 13, color: 'var(--text-secondary)', marginBottom: 12 }}>{spec.description}</div>
      <div style={{ fontSize: 12, color: 'var(--text-muted)', marginBottom: 12 }}>
        {spec.gpu_type && <div>GPU: {spec.gpu_type} x{spec.gpu_count}</div>}
        <div>CPU: {spec.cpu_cores} cores · RAM: {spec.ram_gb} GB</div>
      </div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', paddingTop: 12, borderTop: '1px solid var(--border-color)' }}>
        <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>
          {spec.available_capacity}/{spec.total_capacity} available
        </div>
        <div style={{ fontSize: 18, fontWeight: 600, color: 'var(--success)' }}>
          ${spec.price_per_hour.toFixed(2)}/hr
        </div>
      </div>
    </div>
  )

  return (
    <div>
      {gpuSpecs.length > 0 && (
        <div className="mb-5">
          <h3 style={{ marginBottom: 16 }}>GPU Specs</h3>
          <div className="specs-grid">{gpuSpecs.map(renderSpecCard)}</div>
        </div>
      )}
      {cpuSpecs.length > 0 && (
        <div>
          <h3 style={{ marginBottom: 16 }}>CPU Specs</h3>
          <div className="specs-grid">{cpuSpecs.map(renderSpecCard)}</div>
        </div>
      )}
      {specs.length === 0 && <div className="empty-state">No specs available</div>}
    </div>
  )
}

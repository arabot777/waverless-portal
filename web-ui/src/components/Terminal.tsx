import { useEffect, useRef, useCallback, useState } from 'react'
import { Terminal as XTerm } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { SyncOutlined, DisconnectOutlined } from '@ant-design/icons'
import 'xterm/css/xterm.css'

interface TerminalProps {
  endpoint: string
  workerId: string
}

export default function Terminal({ endpoint, workerId }: TerminalProps) {
  const terminalRef = useRef<HTMLDivElement>(null)
  const xtermRef = useRef<XTerm | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const [status, setStatus] = useState<'connecting' | 'connected' | 'disconnected'>('connecting')
  const [reconnectKey, setReconnectKey] = useState(0)

  const handleReconnect = useCallback(() => {
    wsRef.current?.close()
    xtermRef.current?.dispose()
    setStatus('connecting')
    setReconnectKey(k => k + 1)
  }, [])

  useEffect(() => {
    if (!terminalRef.current) return

    const term = new XTerm({
      cursorBlink: true,
      fontSize: 13,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: { background: '#0d1117', foreground: '#c9d1d9', cursor: '#58a6ff' },
      rows: 20,
      cols: 100,
    })

    const fitAddon = new FitAddon()
    term.loadAddon(fitAddon)
    term.open(terminalRef.current)
    setTimeout(() => fitAddon.fit(), 0)

    xtermRef.current = term
    fitAddonRef.current = fitAddon

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/v1/endpoints/${endpoint}/workers/exec?worker_id=${workerId}`
    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    ws.onopen = () => {
      setStatus('connected')
      term.onData(data => { if (ws.readyState === WebSocket.OPEN) ws.send(data) })
    }

    ws.onmessage = (event) => {
      if (event.data instanceof Blob) {
        const reader = new FileReader()
        reader.onload = () => term.write(reader.result as string)
        reader.readAsText(event.data)
      } else {
        term.write(event.data)
      }
    }

    ws.onerror = () => setStatus('disconnected')
    ws.onclose = () => setStatus('disconnected')

    const handleResize = () => fitAddonRef.current?.fit()
    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      wsRef.current?.close()
      xtermRef.current?.dispose()
    }
  }, [endpoint, workerId, reconnectKey])

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', background: '#0d1117', borderRadius: 8, overflow: 'hidden' }}>
      <div style={{ padding: '6px 12px', background: '#161b22', borderBottom: '1px solid #30363d', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ width: 8, height: 8, borderRadius: '50%', background: status === 'connected' ? '#3fb950' : status === 'connecting' ? '#d29922' : '#f85149' }} />
          <span style={{ fontSize: 12, color: '#8b949e' }}>{status === 'connected' ? 'Connected' : status === 'connecting' ? 'Connecting...' : 'Disconnected'}</span>
        </div>
        <button 
          onClick={handleReconnect} 
          style={{ background: 'transparent', border: '1px solid #30363d', borderRadius: 6, padding: '4px 10px', color: '#c9d1d9', cursor: 'pointer', fontSize: 12, display: 'flex', alignItems: 'center', gap: 4 }}
        >
          {status === 'disconnected' ? <><DisconnectOutlined /> Reconnect</> : <><SyncOutlined /> Restart</>}
        </button>
      </div>
      <div ref={terminalRef} style={{ flex: 1 }} />
    </div>
  )
}

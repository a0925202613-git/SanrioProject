import { useState, useEffect } from 'react'
import { api } from '../api'

const STATUS_COLOR = {
  pending:   { bg: '#fff7ed', color: '#d97706' },
  paid:      { bg: '#eff6ff', color: '#2563eb' },
  shipped:   { bg: '#f5f3ff', color: '#7c3aed' },
  completed: { bg: '#f0fdf4', color: '#16a34a' },
  canceled:  { bg: '#fef2f2', color: '#dc2626' },
}

const STATUS_LABEL = {
  pending:   '待付款',
  paid:      '已付款',
  shipped:   '已出貨',
  completed: '已完成',
  canceled:  '已取消',
}

const CHAR_EMOJI = {
  hello_kitty: '🎀', cinnamoroll: '🐶', pompompurin: '🐾', my_melody: '🐰', kuromi: '💀',
}

const EMPTY_FORM = { product_id: '', quantity: '1' }

export default function OrdersPage({ onPurchase }) {
  const [orders, setOrders] = useState([])
  const [products, setProducts] = useState([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState(EMPTY_FORM)
  const [createLoading, setCreateLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  useEffect(() => {
    load()
    api.products.list(1, 100)
      .then(data => setProducts(Array.isArray(data) ? data : []))
      .catch(() => {})
  }, [])

  async function load() {
    setLoading(true)
    try {
      const data = await api.orders.list()
      setOrders(Array.isArray(data) ? data : [])
    } catch {
      setOrders([])
    } finally {
      setLoading(false)
    }
  }

  async function handleCreate(e) {
    e.preventDefault()
    setCreateLoading(true)
    setError('')
    try {
      const order = await api.orders.create({
        product_id: Number(form.product_id),
        quantity:   Number(form.quantity),
      })
      setSuccess(`訂單 #${order.id} 建立成功！`)
      setShowCreate(false)
      setForm(EMPTY_FORM)
      load()
      onPurchase?.()
      setTimeout(() => setSuccess(''), 4000)
    } catch (err) {
      setError(err.message)
    } finally {
      setCreateLoading(false)
    }
  }

  async function handleStatusChange(orderId, newStatus) {
    const prev = orders.find(o => o.id === orderId)?.status
    // Optimistic update
    setOrders(os => os.map(o => o.id === orderId ? { ...o, status: newStatus } : o))
    try {
      await api.orders.updateStatus(orderId, newStatus)
    } catch (err) {
      // Rollback
      setOrders(os => os.map(o => o.id === orderId ? { ...o, status: prev } : o))
      alert('狀態更新失敗：' + err.message)
    }
  }

  return (
    <div className="page">
      <div className="page-header">
        <h1>我的訂單</h1>
        <button className="btn btn-primary" onClick={() => { setShowCreate(s => !s); setError('') }}>
          {showCreate ? '取消' : '+ 建立訂單'}
        </button>
      </div>

      {success && <div className="alert alert-success">{success}</div>}
      {error   && <div className="alert alert-error">{error}</div>}

      {showCreate && (
        <div className="card create-form-card">
          <h2>建立訂單</h2>
          <p className="form-hint">
            原子性交易：鎖定商品 → 鎖定使用者 → 扣除餘額 → 減少庫存 → 建立訂單。
            任一步驟失敗則全部回滾。
          </p>
          <form onSubmit={handleCreate} className="form form-grid">
            <div className="form-group">
              <label>選擇商品 *</label>
              <select
                value={form.product_id}
                onChange={e => setForm(f => ({ ...f, product_id: e.target.value }))}
                required
              >
                <option value="">— 選擇商品 —</option>
                {products.map(p => (
                  <option key={p.id} value={p.id} disabled={p.stock === 0}>
                    {CHAR_EMOJI[p.character] || '🌸'} {p.name}（NT${Number(p.base_price).toLocaleString()}，剩 {p.stock} 件）
                  </option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label>數量 *</label>
              <input type="number" min="1" max="100" placeholder="1"
                value={form.quantity}
                onChange={e => setForm(f => ({ ...f, quantity: e.target.value }))}
                required
              />
            </div>
            <div className="form-actions form-full">
              <button type="submit" className="btn btn-primary" disabled={createLoading}>
                {createLoading ? '交易處理中…' : '建立訂單'}
              </button>
            </div>
          </form>
        </div>
      )}

      {loading ? (
        <div className="loading">訂單載入中...</div>
      ) : orders.length === 0 ? (
        <div className="empty">尚無訂單。購買商品或完成搶購後即可在此查看。</div>
      ) : (
        <div className="orders-table-wrap">
          <table className="orders-table">
            <thead>
              <tr>
                <th>訂單</th>
                <th>商品</th>
                <th>數量</th>
                <th>單價</th>
                <th>總價</th>
                <th>狀態</th>
                <th>建立時間</th>
              </tr>
            </thead>
            <tbody>
              {orders.map(o => {
                const s = STATUS_COLOR[o.status] || STATUS_COLOR.pending
                return (
                  <tr key={o.id}>
                    <td><span className="order-id">#{o.id}</span></td>
                    <td className="text-muted">#{o.product_id}</td>
                    <td>{o.quantity}</td>
                    <td>NT${Number(o.unit_price).toLocaleString()}</td>
                    <td><strong>NT${Number(o.total_price).toLocaleString()}</strong></td>
                    <td>
                      <select
                        className="status-select"
                        value={o.status}
                        style={{ background: s.bg, color: s.color }}
                        onChange={e => handleStatusChange(o.id, e.target.value)}
                      >
                        {Object.entries(STATUS_LABEL).map(([val, label]) => (
                          <option key={val} value={val}>{label}</option>
                        ))}
                      </select>
                    </td>
                    <td className="text-muted">{new Date(o.created_at).toLocaleString('zh-TW')}</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

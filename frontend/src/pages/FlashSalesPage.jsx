import { useState, useEffect } from 'react'
import { api } from '../api'

export default function FlashSalesPage({ onPurchase }) {
  const [sales, setSales] = useState([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [createForm, setCreateForm] = useState({
    product_id: '', sale_price: '', total_stock: '10',
    start_time: '', end_time: '',
  })
  const [createLoading, setCreateLoading] = useState(false)
  const [createError, setCreateError] = useState('')

  const [selectedId, setSelectedId] = useState('')
  const [goroutines, setGoroutines] = useState(100)
  const [safeResult, setSafeResult] = useState(null)
  const [unsafeResult, setUnsafeResult] = useState(null)
  const [safeLoading, setSafeLoading] = useState(false)
  const [unsafeLoading, setUnsafeLoading] = useState(false)

  // 單次搶購狀態（key: sale.id）
  const [purchaseMsgs, setPurchaseMsgs] = useState({})
  const [purchaseLoadings, setPurchaseLoadings] = useState({})

  useEffect(() => {
    load()
    const now = new Date()
    const start = new Date(now.getTime() - 5000)
    const end   = new Date(now.getTime() + 60 * 60 * 1000)
    setCreateForm(f => ({
      ...f,
      start_time: toLocal(start),
      end_time:   toLocal(end),
    }))
  }, [])

  function toLocal(d) {
    const pad = n => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
  }

  async function load() {
    setLoading(true)
    try {
      const data = await api.flashSales.list()
      setSales(Array.isArray(data) ? data : [])
    } catch {
      setSales([])
    } finally {
      setLoading(false)
    }
  }

  async function handleCreate(e) {
    e.preventDefault()
    setCreateLoading(true)
    setCreateError('')
    try {
      await api.flashSales.create({
        product_id:  Number(createForm.product_id),
        sale_price:  Number(createForm.sale_price),
        total_stock: Number(createForm.total_stock),
        start_time:  new Date(createForm.start_time).toISOString(),
        end_time:    new Date(createForm.end_time).toISOString(),
      })
      setShowCreate(false)
      load()
    } catch (err) {
      setCreateError(err.message)
    } finally {
      setCreateLoading(false)
    }
  }

  async function handlePurchase(saleId, mode) {
    setPurchaseLoadings(prev => ({ ...prev, [saleId]: mode }))
    setPurchaseMsgs(prev => ({ ...prev, [saleId]: null }))
    try {
      const fn = mode === 'safe' ? api.flashSales.purchaseSafe : api.flashSales.purchaseUnsafe
      await fn(saleId)
      setPurchaseMsgs(prev => ({ ...prev, [saleId]: { ok: true, text: '搶購成功！訂單已建立' } }))
      load()
      onPurchase?.()
      setTimeout(() => setPurchaseMsgs(prev => ({ ...prev, [saleId]: null })), 3000)
    } catch (err) {
      setPurchaseMsgs(prev => ({ ...prev, [saleId]: { ok: false, text: err.message } }))
    } finally {
      setPurchaseLoadings(prev => ({ ...prev, [saleId]: null }))
    }
  }

  async function runSafe() {
    if (!selectedId) return alert('請先選擇搶購活動')
    setSafeLoading(true)
    setSafeResult(null)
    try {
      const r = await api.flashSales.concurrentSafe(Number(selectedId), Number(goroutines))
      setSafeResult(r)
      load()
    } catch (err) {
      alert('測試失敗：' + err.message)
    } finally {
      setSafeLoading(false)
    }
  }

  async function runUnsafe() {
    if (!selectedId) return alert('請先選擇搶購活動')
    setUnsafeLoading(true)
    setUnsafeResult(null)
    try {
      const r = await api.flashSales.concurrentUnsafe(Number(selectedId), Number(goroutines))
      setUnsafeResult(r)
      load()
    } catch (err) {
      alert('測試失敗：' + err.message)
    } finally {
      setUnsafeLoading(false)
    }
  }

  const showResults = safeResult || unsafeResult || safeLoading || unsafeLoading

  return (
    <div className="page">
      <div className="page-header">
        <h1>限量搶購</h1>
        <button className="btn btn-primary" onClick={() => { setShowCreate(s => !s); setCreateError('') }}>
          {showCreate ? '取消' : '+ 新增搶購活動'}
        </button>
      </div>

      {showCreate && (
        <div className="card create-form-card">
          <h2>新增搶購活動</h2>
          {createError && <div className="alert alert-error" style={{ marginBottom: 12 }}>{createError}</div>}
          <form onSubmit={handleCreate} className="form form-grid">
            <div className="form-group">
              <label>商品 ID *</label>
              <input type="number" min="1" placeholder="1"
                value={createForm.product_id}
                onChange={e => setCreateForm(f => ({ ...f, product_id: e.target.value }))}
                required
              />
            </div>
            <div className="form-group">
              <label>搶購價格 (NT$) *</label>
              <input type="number" min="1" placeholder="800"
                value={createForm.sale_price}
                onChange={e => setCreateForm(f => ({ ...f, sale_price: e.target.value }))}
                required
              />
            </div>
            <div className="form-group">
              <label>總庫存 *</label>
              <input type="number" min="1" placeholder="10"
                value={createForm.total_stock}
                onChange={e => setCreateForm(f => ({ ...f, total_stock: e.target.value }))}
                required
              />
            </div>
            <div className="form-group" />
            <div className="form-group">
              <label>開始時間</label>
              <input type="datetime-local"
                value={createForm.start_time}
                onChange={e => setCreateForm(f => ({ ...f, start_time: e.target.value }))}
              />
            </div>
            <div className="form-group">
              <label>結束時間</label>
              <input type="datetime-local"
                value={createForm.end_time}
                onChange={e => setCreateForm(f => ({ ...f, end_time: e.target.value }))}
              />
            </div>
            <div className="form-actions form-full">
              <button type="submit" className="btn btn-primary" disabled={createLoading}>
                {createLoading ? '建立中...' : '建立搶購活動'}
              </button>
            </div>
          </form>
        </div>
      )}

      {loading ? (
        <div className="loading">搶購活動載入中...</div>
      ) : sales.length === 0 ? (
        <div className="empty">目前沒有進行中的搶購活動，請於上方新增。</div>
      ) : (
        <div className="flash-sale-grid">
          {sales.map(s => (
            <FlashCard
              key={s.id}
              sale={s}
              selected={String(s.id) === selectedId}
              onSelect={() => setSelectedId(String(s.id))}
              purchaseMsg={purchaseMsgs[s.id]}
              purchaseLoading={purchaseLoadings[s.id]}
              onPurchaseSafe={() => handlePurchase(s.id, 'safe')}
              onPurchaseUnsafe={() => handlePurchase(s.id, 'unsafe')}
            />
          ))}
        </div>
      )}

      {/* ── 並發搶購測試 ── */}
      <div className="demo-section">
        <div className="demo-header">
          <h2>🧪 並發搶購測試</h2>
          <p>
            同時啟動 <strong>N 個 goroutine</strong> 搶購，展示
            Race Condition 與 <code>SELECT FOR UPDATE</code> 行級鎖之間的差異。
          </p>
        </div>

        <div className="demo-controls">
          <div className="form-group">
            <label>搶購活動</label>
            <select value={selectedId} onChange={e => setSelectedId(e.target.value)}>
              <option value="">— 選擇搶購活動 —</option>
              {sales.map(s => (
                <option key={s.id} value={s.id}>
                  #{s.id} — NT${s.sale_price} &nbsp;（剩餘 {s.remaining_stock}/{s.total_stock}）
                </option>
              ))}
            </select>
          </div>
          <div className="form-group">
            <label>Goroutine 數量</label>
            <input type="number" min="1" max="500" value={goroutines}
              onChange={e => setGoroutines(Number(e.target.value))}
            />
          </div>
        </div>

        <div className="test-buttons">
          <button className="btn btn-safe" onClick={runSafe} disabled={safeLoading || !selectedId}>
            {safeLoading
              ? <><span className="spinner" /> 執行中…</>
              : '🔒 安全測試（SELECT FOR UPDATE）'}
          </button>
          <button className="btn btn-unsafe" onClick={runUnsafe} disabled={unsafeLoading || !selectedId}>
            {unsafeLoading
              ? <><span className="spinner" /> 執行中…</>
              : '⚡ 不安全測試（無鎖）'}
          </button>
        </div>

        {showResults && (
          <div className="results-grid">
            <ResultCard result={safeResult}   loading={safeLoading}   mode="safe" />
            <ResultCard result={unsafeResult} loading={unsafeLoading} mode="unsafe" />
          </div>
        )}
      </div>
    </div>
  )
}

/* ─── 搶購活動卡片 ──────────────────────────────────────────── */
function FlashCard({ sale, selected, onSelect, purchaseMsg, purchaseLoading, onPurchaseSafe, onPurchaseUnsafe }) {
  const pct = sale.total_stock > 0
    ? Math.max(0, (sale.remaining_stock / sale.total_stock) * 100)
    : 0
  const fill = pct < 20 ? 'danger' : pct < 50 ? 'warn' : 'good'
  const soldOut = sale.remaining_stock <= 0

  return (
    <div className={`flash-card ${selected ? 'selected' : ''}`} onClick={onSelect}>
      <div className="flash-card-header">
        <span className="flash-id">#{sale.id}</span>
        <div style={{ display: 'flex', gap: 6 }}>
          {soldOut  && <span className="badge badge-error">已售完</span>}
          {selected && <span className="badge badge-primary">已選取</span>}
        </div>
      </div>
      <div className="flash-price">NT${Number(sale.sale_price).toLocaleString()}</div>
      <div className="stock-bar-wrap">
        <div className="stock-bar">
          <div className={`stock-bar-fill ${fill}`} style={{ width: `${pct}%` }} />
        </div>
        <span className="stock-text">剩餘 {sale.remaining_stock} / {sale.total_stock}</span>
      </div>
      <div className="flash-times">
        結束：{new Date(sale.end_time).toLocaleString('zh-TW')}
      </div>

      {/* ── 單次搶購按鈕 ── */}
      <div className="flash-purchase-btns" onClick={e => e.stopPropagation()}>
        <button
          className="btn btn-safe btn-sm"
          disabled={soldOut || !!purchaseLoading}
          onClick={onPurchaseSafe}
        >
          {purchaseLoading === 'safe'
            ? <><span className="spinner" /> 處理中…</>
            : '🔒 安全搶購'}
        </button>
        <button
          className="btn btn-unsafe btn-sm"
          disabled={soldOut || !!purchaseLoading}
          onClick={onPurchaseUnsafe}
        >
          {purchaseLoading === 'unsafe'
            ? <><span className="spinner" /> 處理中…</>
            : '⚡ 不安全搶購'}
        </button>
      </div>
      {purchaseMsg && (
        <div
          className={`alert ${purchaseMsg.ok ? 'alert-success' : 'alert-error'}`}
          style={{ margin: '8px 0 0', fontSize: 13, padding: '6px 10px' }}
        >
          {purchaseMsg.text}
        </div>
      )}
    </div>
  )
}

/* ─── 測試結果卡片 ─────────────────────────────────────────── */
function ResultCard({ result, loading, mode }) {
  const isSafe = mode === 'safe'
  const title  = isSafe ? '🔒 安全模式' : '⚡ 不安全模式'
  const tag    = isSafe ? 'SELECT FOR UPDATE' : '無鎖'

  if (loading) {
    return (
      <div className={`result-card result-${mode} result-loading`}>
        <div className="result-mode-header">
          {title}
          <span className="result-subtitle-tag">{tag}</span>
        </div>
        <div className="result-loading-msg">
          <span className="spinner spinner-lg" />
          {isSafe
            ? 'Goroutine 競爭行鎖中，每次僅一個可繼續執行…'
            : 'Goroutine 同時競爭，留意超賣現象…'}
        </div>
      </div>
    )
  }

  if (!result) {
    return (
      <div className={`result-card result-${mode} result-empty`}>
        <div className="result-mode-header">
          {title}
          <span className="result-subtitle-tag">{tag}</span>
        </div>
        <p className="result-hint" style={{ marginTop: 8 }}>
          {isSafe
            ? '在 Transaction 中使用 SELECT FOR UPDATE，每次只有一個 goroutine 能持有行鎖，並發寫入被序列化，庫存精確不超賣。'
            : '不加鎖。所有 goroutine 同時讀取庫存值並通過檢查後全部寫入，產生經典的 TOCTOU Race Condition，導致超賣。'}
        </p>
      </div>
    )
  }

  const oversold = result.remaining_stock < 0

  return (
    <div className={`result-card result-${mode} ${oversold ? 'result-oversold' : ''}`}>
      <div className="result-mode-header">
        {title}
        <span className="result-subtitle-tag">{tag}</span>
      </div>

      <div className="result-stats">
        <div className="stat-item">
          <div className="stat-value">{result.goroutines}</div>
          <div className="stat-label">Goroutine</div>
        </div>
        <div className="stat-item">
          <div className="stat-value text-success">✅ {result.success_count}</div>
          <div className="stat-label">成功</div>
        </div>
        <div className="stat-item">
          <div className="stat-value text-error">❌ {result.failure_count}</div>
          <div className="stat-label">失敗</div>
        </div>
        <div className="stat-item">
          <div className="stat-value">{result.duration_ms}ms</div>
          <div className="stat-label">耗時</div>
        </div>
      </div>

      <div className="remaining-stock-display">
        <span className="remaining-label">剩餘庫存：</span>
        <span className={`remaining-value ${oversold ? 'text-error' : 'text-success'}`}>
          {result.remaining_stock}
        </span>
        {oversold
          ? <span className="badge badge-error">⚠️ 超賣！</span>
          : <span className="badge badge-success">✓ 無超賣</span>
        }
      </div>

      <p className="result-explanation">
        {isSafe
          ? `恰好 ${result.success_count} 筆購買成功 — 行鎖將 ${result.goroutines} 個 goroutine 序列化執行，剩餘庫存永遠不會低於 0。`
          : oversold
            ? `確認發生 Race Condition！${result.success_count} 個 goroutine 回報成功，但庫存降至 ${result.remaining_stock}。多個 goroutine 同時讀到庫存 > 0 並全部執行了扣減。`
            : `本次未觀察到超賣，但 Race Condition 仍然存在 — 請嘗試增加 goroutine 數量或減少搶購庫存再測試。`
        }
      </p>
    </div>
  )
}

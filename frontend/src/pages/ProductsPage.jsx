import { useState, useEffect, useRef } from 'react'
import { api, resolveImageUrl } from '../api'

const CHARACTERS = [
  { value: '',             label: '全部',        emoji: '🌸',  image: null },
  { value: 'hello_kitty', label: 'Hello Kitty',  emoji: '🎀',  image: '/characters/hello-kitty.png' },
  { value: 'cinnamoroll', label: 'Cinnamoroll',  emoji: '🐶',  image: '/characters/cinnamoroll.png' },
  { value: 'pompompurin', label: 'Pompompurin',  emoji: '🐾',  image: '/characters/pompompurin.png' },
  { value: 'my_melody',   label: 'My Melody',    emoji: '🐰',  image: '/characters/my-melody.png' },
  { value: 'kuromi',      label: 'Kuromi',       emoji: '💀',  image: '/characters/kuromi.png' },
  { value: 'hangyodon',    label: 'Hangyodon',    emoji: '🐟',  image: '/characters/hangyodon.png' },
  { value: 'badtz_maru', label: 'Badtz-Maru', emoji: '🐧',  image: '/characters/badtz-maru.png' },
]

const CHAR_EMOJI = {
  hello_kitty: '🎀', cinnamoroll: '🐶', pompompurin: '🐾', my_melody: '🐰', kuromi: '💀', hangtodon: '🐟', badtz_maru: '🐧'
}

const EMPTY_FORM = { name: '', character: 'cinnamoroll', description: '', base_price: '', stock: '', image_url: '' }

export default function ProductsPage({ onPurchase }) {
  const [products, setProducts] = useState([])
  const [loading, setLoading] = useState(true)
  const [character, setCharacter] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState(EMPTY_FORM)
  const [createLoading, setCreateLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  // 圖片上傳相關狀態
  const [imagePreview, setImagePreview] = useState('')
  const [uploadLoading, setUploadLoading] = useState(false)
  const [uploadError, setUploadError] = useState('')
  const fileInputRef = useRef(null)

  // 編輯商品相關狀態
  const [editingProduct, setEditingProduct] = useState(null)
  const [editForm, setEditForm] = useState({})
  const [editLoading, setEditLoading] = useState(false)
  const [editError, setEditError] = useState('')

  const isLoggedIn = !!localStorage.getItem('token')

  useEffect(() => { load() }, [character])

  async function load() {
    setLoading(true)
    try {
      const data = await api.products.list(1, 20, character)
      setProducts(Array.isArray(data) ? data : [])
    } catch {
      setProducts([])
    } finally {
      setLoading(false)
    }
  }

  async function handleFileChange(e) {
    const file = e.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = ev => setImagePreview(ev.target.result)
    reader.readAsDataURL(file)

    setUploadLoading(true)
    setUploadError('')
    try {
      const data = await api.upload.image(file)
      setForm(f => ({ ...f, image_url: data.url }))
    } catch (err) {
      setUploadError('圖片上傳失敗：' + err.message)
      setImagePreview('')
    } finally {
      setUploadLoading(false)
    }
  }

  function clearImage() {
    setImagePreview('')
    setUploadError('')
    setForm(f => ({ ...f, image_url: '' }))
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  async function handleCreate(e) {
    e.preventDefault()
    setCreateLoading(true)
    setError('')
    try {
      await api.products.create({
        ...form,
        base_price: Number(form.base_price),
        stock:      Number(form.stock),
      })
      setSuccess('商品已建立！')
      setShowCreate(false)
      setForm(EMPTY_FORM)
      setImagePreview('')
      load()
      setTimeout(() => setSuccess(''), 3000)
    } catch (err) {
      setError(err.message)
    } finally {
      setCreateLoading(false)
    }
  }

  function startEdit(p) {
    setEditingProduct(p)
    setEditForm({
      name:        p.name,
      character:   p.character,
      description: p.description || '',
      base_price:  String(p.base_price),
      stock:       String(p.stock),
      image_url:   p.image_url || '',
    })
    setEditError('')
    setShowCreate(false)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  function cancelEdit() {
    setEditingProduct(null)
    setEditForm({})
    setEditError('')
  }

  async function handleEdit(e) {
    e.preventDefault()
    setEditLoading(true)
    setEditError('')
    try {
      await api.products.update(editingProduct.id, {
        ...editForm,
        base_price: Number(editForm.base_price),
        stock:      Number(editForm.stock),
      })
      setSuccess('商品已更新！')
      setEditingProduct(null)
      load()
      setTimeout(() => setSuccess(''), 3000)
    } catch (err) {
      setEditError(err.message)
    } finally {
      setEditLoading(false)
    }
  }

  async function handleDelete(id) {
    if (!confirm('確定要刪除此商品？')) return
    try {
      await api.products.delete(id)
      setProducts(ps => ps.filter(p => p.id !== id))
      if (editingProduct?.id === id) cancelEdit()
    } catch (err) {
      alert(err.message)
    }
  }

  return (
    <div className="page">
      <div className="page-header">
        <h1>商品</h1>
        {isLoggedIn && (
          <button className="btn btn-primary" onClick={() => {
            setShowCreate(s => !s)
            cancelEdit()
            setError('')
            setImagePreview('')
            setUploadError('')
          }}>
            {showCreate ? '取消' : '+ 新增商品'}
          </button>
        )}
      </div>

      {success && <div className="alert alert-success">{success}</div>}
      {error   && <div className="alert alert-error">{error}</div>}

      {/* ── 新增商品表單 ── */}
      {showCreate && (
        <div className="card create-form-card">
          <h2>新增商品</h2>
          <form onSubmit={handleCreate} className="form form-grid">
            <div className="form-group">
              <label>商品名稱 *</label>
              <input placeholder="大耳狗玩偶 30cm"
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div className="form-group">
              <label>角色 *</label>
              <select value={form.character} onChange={e => setForm(f => ({ ...f, character: e.target.value }))}>
                {CHARACTERS.filter(c => c.value).map(c => (
                  <option key={c.value} value={c.value}>{c.emoji} {c.label}</option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label>定價 (NT$) *</label>
              <input type="number" min="1" placeholder="1200"
                value={form.base_price}
                onChange={e => setForm(f => ({ ...f, base_price: e.target.value }))}
                required
              />
            </div>
            <div className="form-group">
              <label>庫存 *</label>
              <input type="number" min="0" placeholder="50"
                value={form.stock}
                onChange={e => setForm(f => ({ ...f, stock: e.target.value }))}
                required
              />
            </div>
            <div className="form-group form-full">
              <label>商品描述</label>
              <input placeholder="超軟超Q的大耳狗玩偶，限量版藍色款式"
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>

            {/* ── 圖片上傳區塊 ── */}
            <div className="form-group form-full">
              <label>商品圖片</label>
              <div className="image-upload-row">
                {imagePreview ? (
                  <div className="image-preview-wrap">
                    <img src={imagePreview} alt="預覽" className="image-preview" />
                    <button type="button" className="image-clear-btn" onClick={clearImage} title="移除圖片">×</button>
                  </div>
                ) : (
                  <div
                    className="upload-area"
                    onClick={() => fileInputRef.current?.click()}
                  >
                    {uploadLoading
                      ? <><span className="spinner spinner-lg" style={{ borderTopColor: 'var(--primary)' }} /> 上傳中…</>
                      : <><span className="upload-icon">📷</span><span>點擊上傳圖片</span><span className="upload-hint">jpg、png、webp，上限 5MB</span></>
                    }
                  </div>
                )}
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".jpg,.jpeg,.png,.gif,.webp"
                  style={{ display: 'none' }}
                  onChange={handleFileChange}
                />
                <div className="upload-or">或</div>
                <div style={{ flex: 1 }}>
                  <input
                    placeholder="https://example.com/image.jpg"
                    value={form.image_url}
                    onChange={e => {
                      setForm(f => ({ ...f, image_url: e.target.value }))
                      setImagePreview(e.target.value)
                    }}
                  />
                  <div className="upload-hint" style={{ marginTop: 4 }}>或直接輸入圖片網址</div>
                </div>
              </div>
              {uploadError && <div className="alert alert-error" style={{ marginTop: 8 }}>{uploadError}</div>}
            </div>

            <div className="form-actions form-full">
              <button type="submit" className="btn btn-primary" disabled={createLoading || uploadLoading}>
                {createLoading ? '建立中...' : '建立商品'}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* ── 編輯商品表單 ── */}
      {editingProduct && (
        <div className="card create-form-card" style={{ borderLeft: '3px solid var(--primary)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
            <h2>編輯商品 #{editingProduct.id}</h2>
            <button className="btn btn-outline btn-sm" onClick={cancelEdit}>取消</button>
          </div>
          {editError && <div className="alert alert-error" style={{ marginBottom: 12 }}>{editError}</div>}
          <form onSubmit={handleEdit} className="form form-grid">
            <div className="form-group">
              <label>商品名稱 *</label>
              <input
                value={editForm.name}
                onChange={e => setEditForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </div>
            <div className="form-group">
              <label>角色 *</label>
              <select value={editForm.character} onChange={e => setEditForm(f => ({ ...f, character: e.target.value }))}>
                {CHARACTERS.filter(c => c.value).map(c => (
                  <option key={c.value} value={c.value}>{c.emoji} {c.label}</option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label>定價 (NT$) *</label>
              <input type="number" min="1"
                value={editForm.base_price}
                onChange={e => setEditForm(f => ({ ...f, base_price: e.target.value }))}
                required
              />
            </div>
            <div className="form-group">
              <label>庫存 *</label>
              <input type="number" min="0"
                value={editForm.stock}
                onChange={e => setEditForm(f => ({ ...f, stock: e.target.value }))}
                required
              />
            </div>
            <div className="form-group form-full">
              <label>商品描述</label>
              <input
                value={editForm.description}
                onChange={e => setEditForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div className="form-group form-full">
              <label>圖片網址</label>
              <input
                placeholder="https://example.com/image.jpg 或留空移除圖片"
                value={editForm.image_url}
                onChange={e => setEditForm(f => ({ ...f, image_url: e.target.value }))}
              />
            </div>
            <div className="form-actions form-full">
              <button type="submit" className="btn btn-primary" disabled={editLoading}>
                {editLoading ? '儲存中...' : '儲存變更'}
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="character-tabs">
        {CHARACTERS.map(c => (
          <button
            key={c.value}
            className={`char-tab ${character === c.value ? 'active' : ''}`}
            onClick={() => setCharacter(c.value)}
          >
            <CharAvatar image={c.image} emoji={c.emoji} label={c.label} />
            <span className="char-name">{c.label}</span>
          </button>
        ))}
      </div>

      {loading ? (
        <div className="loading">商品載入中...</div>
      ) : products.length === 0 ? (
        <div className="empty">
          {isLoggedIn ? '尚無商品，請於上方新增。' : '尚無商品，請先登入以新增商品。'}
        </div>
      ) : (
        <div className="product-grid">
          {products.map(p => (
            <ProductCard
              key={p.id}
              product={p}
              isLoggedIn={isLoggedIn}
              isEditing={editingProduct?.id === p.id}
              onEdit={startEdit}
              onDelete={handleDelete}
              onBuySuccess={onPurchase}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function CharAvatar({ image, emoji, label }) {
  const [failed, setFailed] = useState(false)
  return (
    <div className="char-avatar">
      {image && !failed
        ? <img src={image} alt={label} onError={() => setFailed(true)} />
        : <span className="char-emoji">{emoji}</span>
      }
    </div>
  )
}

function ProductCard({ product: p, isLoggedIn, isEditing, onEdit, onDelete, onBuySuccess }) {
  const [imgError, setImgError] = useState(false)
  const [buyLoading, setBuyLoading] = useState(false)
  const [buyMsg, setBuyMsg] = useState(null)

  const src = resolveImageUrl(p.image_url)
  const showImg = src && !imgError

  async function handleBuy() {
    setBuyLoading(true)
    setBuyMsg(null)
    try {
      const order = await api.orders.create({ product_id: p.id, quantity: 1 })
      setBuyMsg({ ok: true, text: `訂單 #${order.id} 建立成功！` })
      onBuySuccess?.()
      setTimeout(() => setBuyMsg(null), 3000)
    } catch (err) {
      setBuyMsg({ ok: false, text: err.message })
    } finally {
      setBuyLoading(false)
    }
  }

  return (
    <div className={`product-card ${isEditing ? 'product-card-editing' : ''}`}>
      {showImg ? (
        <div className="product-img-wrap">
          <img
            src={src}
            alt={p.name}
            className="product-img"
            onError={() => setImgError(true)}
          />
        </div>
      ) : (
        <div className="product-emoji">{CHAR_EMOJI[p.character] || '🌸'}</div>
      )}

      <div className="product-info">
        <h3>{p.name}</h3>
        <span className="character-tag">{p.character.replace(/_/g, ' ')}</span>
        {p.description && <p className="product-desc">{p.description}</p>}
      </div>

      <div className="product-footer">
        <span className="product-price">NT${Number(p.base_price).toLocaleString()}</span>
        <span className={`stock-badge ${p.stock === 0 ? 'out' : p.stock < 10 ? 'low' : 'ok'}`}>
          {p.stock === 0 ? '已售完' : `剩 ${p.stock} 件`}
        </span>
      </div>

      {isLoggedIn && (
        <div className="card-actions">
          <button
            className="btn btn-primary btn-sm"
            disabled={p.stock === 0 || buyLoading}
            onClick={handleBuy}
          >
            {buyLoading
              ? <><span className="spinner" /> 處理中…</>
              : p.stock === 0 ? '已售完' : '🛒 購買'}
          </button>
          <div className="card-actions-row">
            <button className="edit-btn" onClick={() => onEdit(p)}>✏️ 編輯</button>
            <button className="delete-btn" onClick={() => onDelete(p.id)}>× 刪除</button>
          </div>
          {buyMsg && (
            <div className={`alert ${buyMsg.ok ? 'alert-success' : 'alert-error'}`}
              style={{ fontSize: 13, padding: '6px 10px' }}>
              {buyMsg.text}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

import { useState, useEffect, useRef } from 'react'
import { api, resolveImageUrl } from '../api'

const CHARACTERS = [
  { value: '',            label: '全部',          emoji: '🌸' },
  { value: 'hello_kitty', label: 'Hello Kitty',  emoji: '🎀' },
  { value: 'cinnamoroll', label: 'Cinnamoroll',  emoji: '🐶' },
  { value: 'pompompurin', label: 'Pompompurin',  emoji: '🐾' },
  { value: 'my_melody',   label: 'My Melody',    emoji: '🐰' },
  { value: 'kuromi',      label: 'Kuromi',       emoji: '💀' },
  { value: 'hangyodon',  label: 'Hangyodon',     emoji: '🐟' },
  { value: 'badtz_maru',  label: 'Badtz maru',   emoji: '🐧' },

]

const CHAR_EMOJI = {
  hello_kitty: '🎀', cinnamoroll: '🐶', pompompurin: '🐾', my_melody: '🐰', kuromi: '💀', hangyodon: '🐟', badtz_maru: '🐧'
}

const EMPTY_FORM = { name: '', character: 'cinnamoroll', description: '', base_price: '', stock: '', image_url: '' }

export default function ProductsPage() {
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

    // 立即顯示本地預覽（不等上傳完成）
    const reader = new FileReader()
    reader.onload = ev => setImagePreview(ev.target.result)
    reader.readAsDataURL(file)

    // 上傳至伺服器
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

  async function handleDelete(id) {
    if (!confirm('確定要刪除此商品？')) return
    try {
      await api.products.delete(id)
      setProducts(ps => ps.filter(p => p.id !== id))
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
              <label>定價 (¥) *</label>
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
                {/* 預覽區 */}
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

                {/* 隱藏的 file input */}
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".jpg,.jpeg,.png,.gif,.webp"
                  style={{ display: 'none' }}
                  onChange={handleFileChange}
                />

                {/* 或直接輸入外部 URL */}
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

      <div className="character-tabs">
        {CHARACTERS.map(c => (
          <button
            key={c.value}
            className={`char-tab ${character === c.value ? 'active' : ''}`}
            onClick={() => setCharacter(c.value)}
          >
            {c.emoji} {c.label}
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
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}
    </div>
  )
};

function ProductCard({ product: p, isLoggedIn, onDelete }) {
  const [imgError, setImgError] = useState(false)
  const src = resolveImageUrl(p.image_url)
  const showImg = src && !imgError

  return (
    <div className="product-card">
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
        <span className="product-price">¥{Number(p.base_price).toLocaleString()}</span>
        <span className={`stock-badge ${p.stock === 0 ? 'out' : p.stock < 10 ? 'low' : 'ok'}`}>
          {p.stock === 0 ? '已售完' : `剩 ${p.stock} 件`}
        </span>
      </div>

      {isLoggedIn && (
        <button className="delete-btn" title="刪除" onClick={() => onDelete(p.id)}>×</button>
      )}
    </div>
  )
}

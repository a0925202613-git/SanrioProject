import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api'

export default function LoginPage() {
  const navigate = useNavigate()
  const [tab, setTab] = useState('login')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const [loginForm, setLoginForm] = useState({ email: '', password: '' })
  const [regForm, setRegForm] = useState({ username: '', email: '', password: '', balance: 50000 })

  function switchTab(t) {
    setTab(t)
    setError('')
  }

  async function handleLogin(e) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const data = await api.auth.login(loginForm)
      localStorage.setItem('token', data.token)
      navigate('/flash-sales')
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  async function handleRegister(e) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await api.auth.register({ ...regForm, balance: Number(regForm.balance) })
      const data = await api.auth.login({ email: regForm.email, password: regForm.password })
      localStorage.setItem('token', data.token)
      navigate('/flash-sales')
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page-center">
      <div className="auth-card">
        <div className="auth-header">
          <span className="auth-logo">🌸</span>
          <h1>三麗鷗搶購系統</h1>
          <p>限量商品・先搶先得</p>
        </div>

        <div className="tab-group">
          <button className={`tab-btn ${tab === 'login' ? 'active' : ''}`} onClick={() => switchTab('login')}>
            登入
          </button>
          <button className={`tab-btn ${tab === 'register' ? 'active' : ''}`} onClick={() => switchTab('register')}>
            註冊
          </button>
        </div>

        {error && <div className="alert alert-error" style={{ marginBottom: 16 }}>{error}</div>}

        {tab === 'login' ? (
          <form onSubmit={handleLogin} className="form">
            <div className="form-group">
              <label>電子郵件</label>
              <input type="email" placeholder="kitty@sanrio.com"
                value={loginForm.email}
                onChange={e => setLoginForm(f => ({ ...f, email: e.target.value }))}
                required
              />
            </div>
            <div className="form-group">
              <label>密碼</label>
              <input type="password" placeholder="••••••••"
                value={loginForm.password}
                onChange={e => setLoginForm(f => ({ ...f, password: e.target.value }))}
                required
              />
            </div>
            <button type="submit" className="btn btn-primary btn-full" disabled={loading}>
              {loading ? '登入中...' : '登入'}
            </button>
          </form>
        ) : (
          <form onSubmit={handleRegister} className="form">
            <div className="form-group">
              <label>使用者名稱</label>
              <input placeholder="kitty_fan"
                value={regForm.username}
                onChange={e => setRegForm(f => ({ ...f, username: e.target.value }))}
                required
              />
            </div>
            <div className="form-group">
              <label>電子郵件</label>
              <input type="email" placeholder="kitty@sanrio.com"
                value={regForm.email}
                onChange={e => setRegForm(f => ({ ...f, email: e.target.value }))}
                required
              />
            </div>
            <div className="form-group">
              <label>密碼</label>
              <input type="password" placeholder="••••••••"
                value={regForm.password}
                onChange={e => setRegForm(f => ({ ...f, password: e.target.value }))}
                required
              />
            </div>
            <div className="form-group">
              <label>初始餘額 (¥)</label>
              <input type="number" min="0"
                value={regForm.balance}
                onChange={e => setRegForm(f => ({ ...f, balance: e.target.value }))}
              />
            </div>
            <button type="submit" className="btn btn-primary btn-full" disabled={loading}>
              {loading ? '註冊中...' : '註冊並登入'}
            </button>
          </form>
        )}
      </div>
    </div>
  )
}

import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { Store, Wallet, LogOut } from 'lucide-react'
import { api } from '../api'

export default function NavBar({ refreshKey = 0 }) {
  const location = useLocation()
  const navigate  = useNavigate()
  const [user, setUser]   = useState(null)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (token) {
      api.auth.me()
        .then(setUser)
        .catch(() => { localStorage.removeItem('token'); setUser(null) })
    } else {
      setUser(null)
    }
  }, [location.pathname, refreshKey])

  function logout() {
    localStorage.removeItem('token')
    setUser(null)
    navigate('/login')
  }

  const isActive = path => location.pathname === path

  return (
    <nav className="navbar">
      <Link to="/flash-sales" className="navbar-brand">
        <span className="navbar-logo"><Store size={20} /></span>
        <span className="navbar-title">三麗鷗搶購系統</span>
      </Link>

      <div className="navbar-links">
        <Link to="/products"    className={`nav-link ${isActive('/products')    ? 'active' : ''}`}>商品</Link>
        <Link to="/flash-sales" className={`nav-link ${isActive('/flash-sales') ? 'active' : ''}`}>限量搶購</Link>
        {user && (
          <Link to="/orders" className={`nav-link ${isActive('/orders') ? 'active' : ''}`}>我的訂單</Link>
        )}
      </div>

      <div className="navbar-auth">
        {user ? (
          <>
            <div className="navbar-user">
              <div className="navbar-avatar">
                {user.username.charAt(0).toUpperCase()}
              </div>
              <span>{user.username}</span>
            </div>
            <span className="navbar-balance">
              <Wallet size={13} />
              NT${Number(user.balance).toLocaleString()}
            </span>
            <button className="btn btn-ghost btn-sm" onClick={logout}>
              <LogOut size={14} />
              登出
            </button>
          </>
        ) : (
          <Link to="/login" className="btn btn-primary btn-sm">登入</Link>
        )}
      </div>
    </nav>
  )
}

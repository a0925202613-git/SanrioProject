import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { api } from '../api'

export default function NavBar() {
  const location = useLocation()
  const navigate = useNavigate()
  const [user, setUser] = useState(null)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (token) {
      api.auth.me()
        .then(setUser)
        .catch(() => {
          localStorage.removeItem('token')
          setUser(null)
        })
    } else {
      setUser(null)
    }
  }, [location.pathname])

  function logout() {
    localStorage.removeItem('token')
    setUser(null)
    navigate('/login')
  }

  const isActive = path => location.pathname === path

  return (
    <nav className="navbar">
      <div className="navbar-brand">
        <span className="navbar-logo">🌸</span>
        <span className="navbar-title">三麗鷗搶購系統</span>
      </div>

      <div className="navbar-links">
        <Link to="/products" className={`nav-link ${isActive('/products') ? 'active' : ''}`}>
          商品
        </Link>
        <Link to="/flash-sales" className={`nav-link ${isActive('/flash-sales') ? 'active' : ''}`}>
          限量搶購
        </Link>
        {user && (
          <Link to="/orders" className={`nav-link ${isActive('/orders') ? 'active' : ''}`}>
            我的訂單
          </Link>
        )}
      </div>

      <div className="navbar-auth">
        {user ? (
          <>
            <span className="navbar-user">👤 {user.username}</span>
            <span className="navbar-balance">¥{Number(user.balance).toLocaleString()}</span>
            <button className="btn btn-outline btn-sm" onClick={logout}>登出</button>
          </>
        ) : (
          <Link to="/login" className="btn btn-primary btn-sm">登入</Link>
        )}
      </div>
    </nav>
  )
}

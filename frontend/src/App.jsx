import { useState, useCallback } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import NavBar from './components/NavBar'
import LoginPage from './pages/LoginPage'
import ProductsPage from './pages/ProductsPage'
import FlashSalesPage from './pages/FlashSalesPage'
import OrdersPage from './pages/OrdersPage'

function ProtectedRoute({ children }) {
  return localStorage.getItem('token') ? children : <Navigate to="/login" replace />
}

export default function App() {
  const [refreshKey, setRefreshKey] = useState(0)
  const bumpRefresh = useCallback(() => setRefreshKey(k => k + 1), [])

  return (
    <BrowserRouter>
      <NavBar refreshKey={refreshKey} />
      <main className="main-content">
        <Routes>
          <Route path="/login" element={<LoginPage onLogin={bumpRefresh} />} />
          <Route path="/products" element={<ProductsPage onPurchase={bumpRefresh} />} />
          <Route path="/flash-sales" element={
            <ProtectedRoute><FlashSalesPage onPurchase={bumpRefresh} /></ProtectedRoute>
          } />
          <Route path="/orders" element={
            <ProtectedRoute><OrdersPage onPurchase={bumpRefresh} /></ProtectedRoute>
          } />
          <Route path="*" element={<Navigate to="/flash-sales" replace />} />
        </Routes>
      </main>
    </BrowserRouter>
  )
}

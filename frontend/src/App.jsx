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
  return (
    <BrowserRouter>
      <NavBar />
      <main className="main-content">
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/products" element={<ProductsPage />} />
          <Route path="/flash-sales" element={
            <ProtectedRoute><FlashSalesPage /></ProtectedRoute>
          } />
          <Route path="/orders" element={
            <ProtectedRoute><OrdersPage /></ProtectedRoute>
          } />
          <Route path="*" element={<Navigate to="/flash-sales" replace />} />
        </Routes>
      </main>
    </BrowserRouter>
  )
}

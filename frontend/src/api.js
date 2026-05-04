const BASE     = 'http://localhost:9090/api/v1'
const BASE_URL = 'http://localhost:9090'

// 將後端回傳的相對路徑（如 /uploads/xxx.jpg）補上 host
export function resolveImageUrl(url) {
  if (!url) return ''
  return url.startsWith('/') ? BASE_URL + url : url
}

function getToken() {
  return localStorage.getItem('token')
}

async function request(path, opts = {}) {
  const headers = { 'Content-Type': 'application/json', ...opts.headers }
  const token = getToken()
  if (token) headers.Authorization = `Bearer ${token}`

  const res = await fetch(BASE + path, { ...opts, headers })
  const json = await res.json()
  if (!res.ok) throw new Error(json.message || 'Request failed')
  return json.data
}

export const api = {
  auth: {
    register: b => request('/auth/register', { method: 'POST', body: JSON.stringify(b) }),
    login: b => request('/auth/login', { method: 'POST', body: JSON.stringify(b) }),
    me: () => request('/me'),
  },
  products: {
    list: (page = 1, size = 20, character = '') =>
      request(`/products?page=${page}&page_size=${size}${character ? `&character=${character}` : ''}`),
    get: id => request(`/products/${id}`),
    create: b => request('/products', { method: 'POST', body: JSON.stringify(b) }),
    update: (id, b) => request(`/products/${id}`, { method: 'PUT', body: JSON.stringify(b) }),
    delete: id => request(`/products/${id}`, { method: 'DELETE' }),
  },
  flashSales: {
    list: () => request('/flash-sales'),
    get: id => request(`/flash-sales/${id}`),
    create: b => request('/flash-sales', { method: 'POST', body: JSON.stringify(b) }),
    purchaseSafe: id => request(`/flash-sales/${id}/purchase/safe`, { method: 'POST' }),
    purchaseUnsafe: id => request(`/flash-sales/${id}/purchase/unsafe`, { method: 'POST' }),
    concurrentSafe: (id, goroutines) =>
      request(`/flash-sales/${id}/concurrent-test/safe`, { method: 'POST', body: JSON.stringify({ goroutines }) }),
    concurrentUnsafe: (id, goroutines) =>
      request(`/flash-sales/${id}/concurrent-test/unsafe`, { method: 'POST', body: JSON.stringify({ goroutines }) }),
  },
  upload: {
    // 上傳圖片，回傳 { url: "/uploads/<filename>" }
    // 注意：不能加 Content-Type header，讓瀏覽器自動設定 multipart boundary
    image: file => {
      const form = new FormData()
      form.append('image', file)
      const token = getToken()
      return fetch(BASE + '/upload', {
        method: 'POST',
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        body: form,
      }).then(async res => {
        const json = await res.json()
        if (!res.ok) throw new Error(json.message || '上傳失敗')
        return json.data
      })
    },
  },
  orders: {
    list: () => request('/orders'),
    get: id => request(`/orders/${id}`),
    create: b => request('/orders', { method: 'POST', body: JSON.stringify(b) }),
  },
}

import { useState, useEffect } from 'react'
import Auth from './components/Auth'
import Upload from './components/Upload'
import RecentUploads from './components/RecentUploads'

function App() {
  const [token, setToken] = useState(localStorage.getItem('auth_token') || '')
  const [recentImages, setRecentImages] = useState([])

  const handleLogin = (newToken) => {
    setToken(newToken)
    localStorage.setItem('auth_token', newToken)
  }

  const handleLogout = () => {
    setToken('')
    localStorage.removeItem('auth_token')
    setRecentImages([])
  }

  const handleUploadSuccess = (imageData) => {
    setRecentImages(prev => [imageData, ...prev])
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="container mx-auto px-4 py-8">
        <header className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-800 mb-2">图床系统</h1>
          <p className="text-gray-600">快速上传和分享图片</p>
        </header>

        {!token ? (
          <Auth onLogin={handleLogin} />
        ) : (
          <div className="space-y-8">
            <div className="flex justify-end">
              <button
                onClick={handleLogout}
                className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
              >
                退出登录
              </button>
            </div>
            <Upload token={token} onUploadSuccess={handleUploadSuccess} />
            <RecentUploads token={token} recentImages={recentImages} setRecentImages={setRecentImages} />
          </div>
        )}
      </div>
    </div>
  )
}

export default App

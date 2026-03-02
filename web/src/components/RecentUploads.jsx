import { useState, useEffect } from 'react'
import axios from 'axios'
import ImagePreview from './ImagePreview'

function RecentUploads({ token, recentImages, setRecentImages }) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    fetchRecentUploads()
  }, [token])

  const fetchRecentUploads = async () => {
    setLoading(true)
    setError('')
    try {
      const response = await axios.get('/api/recent', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (response.data.success) {
        setRecentImages(response.data.data || [])
      }
    } catch (err) {
      setError(err.response?.data?.message || '获取最近上传失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="bg-white rounded-lg shadow-lg p-6">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold text-gray-800">最近上传</h2>
        <button
          onClick={fetchRecentUploads}
          disabled={loading}
          className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors disabled:bg-gray-400 text-sm"
        >
          {loading ? '刷新中...' : '刷新'}
        </button>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
          {error}
        </div>
      )}

      {loading && recentImages.length === 0 ? (
        <div className="text-center py-8 text-gray-500">加载中...</div>
      ) : recentImages.length === 0 ? (
        <div className="text-center py-8 text-gray-500">暂无上传记录</div>
      ) : (
        <div className="space-y-4">
          {recentImages.map((image, index) => (
            <ImagePreview key={image.url || index} imageData={image} />
          ))}
        </div>
      )}
    </div>
  )
}

export default RecentUploads

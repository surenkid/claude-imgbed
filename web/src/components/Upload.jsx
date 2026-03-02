import { useState, useRef, useEffect } from 'react'
import axios from 'axios'
import ImagePreview from './ImagePreview'

function Upload({ token, onUploadSuccess }) {
  const [files, setFiles] = useState([])
  const [uploading, setUploading] = useState(false)
  const [uploadProgress, setUploadProgress] = useState({})
  const [dragActive, setDragActive] = useState(false)
  const [error, setError] = useState('')
  const fileInputRef = useRef(null)

  // 监听粘贴事件
  useEffect(() => {
    const handlePaste = (e) => {
      const items = e.clipboardData?.items
      if (!items) return

      const imageFiles = []
      for (let i = 0; i < items.length; i++) {
        if (items[i].type.indexOf('image') !== -1) {
          const file = items[i].getAsFile()
          if (file) imageFiles.push(file)
        }
      }

      if (imageFiles.length > 0) {
        handleFiles(imageFiles)
      }
    }

    document.addEventListener('paste', handlePaste)
    return () => document.removeEventListener('paste', handlePaste)
  }, [])

  const handleFiles = (newFiles) => {
    const validFiles = Array.from(newFiles).filter(file => {
      if (!file.type.startsWith('image/')) {
        setError('只能上传图片文件')
        return false
      }
      if (file.size > 5 * 1024 * 1024) {
        setError('图片大小不能超过5MB')
        return false
      }
      return true
    })

    if (validFiles.length > 0) {
      setError('')
      setFiles(prev => [...prev, ...validFiles.map(file => ({
        file,
        preview: URL.createObjectURL(file),
        id: Math.random().toString(36).substr(2, 9)
      }))])
    }
  }

  const handleFileSelect = (e) => {
    if (e.target.files?.length) {
      handleFiles(e.target.files)
    }
  }

  const handleDrag = (e) => {
    e.preventDefault()
    e.stopPropagation()
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true)
    } else if (e.type === 'dragleave') {
      setDragActive(false)
    }
  }

  const handleDrop = (e) => {
    e.preventDefault()
    e.stopPropagation()
    setDragActive(false)

    if (e.dataTransfer.files?.length) {
      handleFiles(e.dataTransfer.files)
    }
  }

  const removeFile = (id) => {
    setFiles(prev => {
      const file = prev.find(f => f.id === id)
      if (file) URL.revokeObjectURL(file.preview)
      return prev.filter(f => f.id !== id)
    })
  }

  const uploadFile = async (fileObj) => {
    const formData = new FormData()
    formData.append('file', fileObj.file)

    try {
      const response = await axios.post('/api/upload', formData, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'multipart/form-data'
        },
        onUploadProgress: (progressEvent) => {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          setUploadProgress(prev => ({ ...prev, [fileObj.id]: progress }))
        }
      })

      if (response.data.success) {
        onUploadSuccess(response.data.data)
        return response.data.data
      }
    } catch (err) {
      throw new Error(err.response?.data?.message || '上传失败')
    }
  }

  const handleUpload = async () => {
    if (files.length === 0) return

    setUploading(true)
    setError('')

    try {
      await Promise.all(files.map(fileObj => uploadFile(fileObj)))

      // 清理预览
      files.forEach(f => URL.revokeObjectURL(f.preview))
      setFiles([])
      setUploadProgress({})
    } catch (err) {
      setError(err.message)
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="bg-white rounded-lg shadow-lg p-6">
      <h2 className="text-2xl font-bold text-gray-800 mb-4">上传图片</h2>

      {/* 上传区域 */}
      <div
        className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
          dragActive ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-gray-400'
        }`}
        onDragEnter={handleDrag}
        onDragLeave={handleDrag}
        onDragOver={handleDrag}
        onDrop={handleDrop}
      >
        <div className="space-y-4">
          <div className="text-gray-600">
            <p className="text-lg font-medium">拖拽图片到此处</p>
            <p className="text-sm mt-2">或者</p>
          </div>
          <button
            type="button"
            onClick={() => fileInputRef.current?.click()}
            className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
          >
            选择文件
          </button>
          <input
            ref={fileInputRef}
            type="file"
            multiple
            accept="image/*"
            onChange={handleFileSelect}
            className="hidden"
          />
          <p className="text-sm text-gray-500">支持 Ctrl+V 粘贴上传</p>
          <p className="text-xs text-gray-400">支持 JPG、PNG、GIF、WebP，最大 5MB</p>
        </div>
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
          {error}
        </div>
      )}

      {/* 预览区域 */}
      {files.length > 0 && (
        <div className="mt-6">
          <h3 className="text-lg font-medium text-gray-800 mb-3">
            待上传 ({files.length})
          </h3>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {files.map(fileObj => (
              <div key={fileObj.id} className="relative group">
                <img
                  src={fileObj.preview}
                  alt="preview"
                  className="w-full h-32 object-cover rounded-lg"
                />
                <button
                  onClick={() => removeFile(fileObj.id)}
                  className="absolute top-2 right-2 bg-red-500 text-white rounded-full w-6 h-6 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  ×
                </button>
                {uploadProgress[fileObj.id] !== undefined && (
                  <div className="absolute bottom-0 left-0 right-0 bg-black bg-opacity-50 text-white text-xs p-1 rounded-b-lg">
                    {uploadProgress[fileObj.id]}%
                  </div>
                )}
              </div>
            ))}
          </div>
          <button
            onClick={handleUpload}
            disabled={uploading}
            className="mt-4 w-full py-3 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors disabled:bg-gray-400 disabled:cursor-not-allowed font-medium"
          >
            {uploading ? '上传中...' : `上传 ${files.length} 张图片`}
          </button>
        </div>
      )}
    </div>
  )
}

export default Upload

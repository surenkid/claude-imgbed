import { useState } from 'react'

function ImagePreview({ imageData }) {
  const [copied, setCopied] = useState(false)

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  return (
    <div className="bg-white rounded-lg shadow p-4">
      <div className="flex flex-col md:flex-row gap-4">
        <div className="flex-shrink-0">
          <img
            src={imageData.thumbnail || imageData.url}
            alt={imageData.filename}
            className="w-32 h-32 object-cover rounded-lg"
          />
        </div>
        <div className="flex-1 space-y-2">
          <div>
            <p className="text-sm text-gray-600">文件名</p>
            <p className="text-sm font-medium">{imageData.filename}</p>
          </div>
          <div>
            <p className="text-sm text-gray-600">大小</p>
            <p className="text-sm font-medium">{(imageData.size / 1024).toFixed(2)} KB</p>
          </div>
          {imageData.width && imageData.height && (
            <div>
              <p className="text-sm text-gray-600">尺寸</p>
              <p className="text-sm font-medium">{imageData.width} × {imageData.height}</p>
            </div>
          )}
          <div>
            <p className="text-sm text-gray-600 mb-1">直链</p>
            <div className="flex gap-2">
              <input
                type="text"
                value={imageData.url}
                readOnly
                className="flex-1 px-3 py-1 text-sm border border-gray-300 rounded bg-gray-50"
              />
              <button
                onClick={() => copyToClipboard(imageData.url)}
                className="px-4 py-1 bg-blue-500 text-white text-sm rounded hover:bg-blue-600 transition-colors"
              >
                {copied ? '已复制' : '复制'}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default ImagePreview

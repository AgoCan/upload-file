import SparkMD5 from 'spark-md5'

/**
 * 计算文件的MD5哈希值
 * @param {File} file 文件对象
 * @param {Function} onProgress 进度回调
 * @returns {Promise<string>} MD5哈希值
 */
export function calculateFileHash(file, onProgress) {
  return new Promise((resolve, reject) => {
    const chunkSize = 2 * 1024 * 1024 // 2MB
    const chunks = Math.ceil(file.size / chunkSize)
    let currentChunk = 0
    const spark = new SparkMD5.ArrayBuffer()
    const fileReader = new FileReader()

    fileReader.onload = (e) => {
      spark.append(e.target.result)
      currentChunk++

      if (onProgress) {
        onProgress(Math.floor((currentChunk / chunks) * 100))
      }

      if (currentChunk < chunks) {
        loadNext()
      } else {
        resolve(spark.end())
      }
    }

    fileReader.onerror = (e) => {
      reject(e)
    }

    function loadNext() {
      const start = currentChunk * chunkSize
      const end = start + chunkSize >= file.size ? file.size : start + chunkSize
      const chunk = file.slice(start, end)
      fileReader.readAsArrayBuffer(chunk)
    }

    loadNext()
  })
}

/**
 * 将文件分割成块
 * @param {File} file 文件对象
 * @param {number} chunkSize 块大小
 * @returns {Array<Blob>} 文件块数组
 */
export function sliceFile(file, chunkSize) {
  const chunks = []
  let start = 0

  while (start < file.size) {
    const end = Math.min(start + chunkSize, file.size)
    const chunk = file.slice(start, end)
    chunks.push(chunk)
    start = end
  }

  return chunks
}

/**
 * 格式化文件大小
 * @param {number} bytes 字节数
 * @returns {string} 格式化后的文件大小
 */
export function formatFileSize(bytes) {
  if (bytes === 0) return '0 B'

  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * 获取文件扩展名
 * @param {string} filename 文件名
 * @returns {string} 文件扩展名
 */
export function getFileExtension(filename) {
  return filename.slice((filename.lastIndexOf('.') - 1 >>> 0) + 2)
}

/**
 * 根据文件类型获取图标
 * @param {string} filename 文件名
 * @returns {string} 图标类名
 */
export function getFileIcon(filename) {
  const extension = getFileExtension(filename).toLowerCase()
  
  const iconMap = {
    pdf: 'Document',
    doc: 'Document',
    docx: 'Document',
    xls: 'Document',
    xlsx: 'Document',
    ppt: 'Document',
    pptx: 'Document',
    jpg: 'Picture',
    jpeg: 'Picture',
    png: 'Picture',
    gif: 'Picture',
    zip: 'Folder',
    rar: 'Folder',
    '7z': 'Folder',
    txt: 'Document',
    mp4: 'VideoCamera',
    mp3: 'Headset'
  }

  return iconMap[extension] || 'Document'
}
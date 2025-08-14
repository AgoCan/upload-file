import request from '@/utils/request'

// 初始化上传
export function initUpload(data) {
  return request({
    url: '/upload/init',
    method: 'post',
    data
  })
}

// 上传分片
export function uploadChunk(data, onUploadProgress) {
  return request({
    url: '/upload/chunk',
    method: 'post',
    data,
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    onUploadProgress
  })
}

// 完成上传
export function completeUpload(data) {
  return request({
    url: '/upload/complete',
    method: 'post',
    data
  })
}

// 获取上传状态
export function getUploadStatus(fileId) {
  return request({
    url: `/upload/status/${fileId}`,
    method: 'get'
  })
}

// 获取文件列表
export function getFileList() {
  return request({
    url: '/upload/files',
    method: 'get'
  })
}

// 下载文件
export function downloadFile(fileId) {
  window.open(`/api/v1/upload/file/${fileId}`)
}

// 删除文件
export function deleteFile(fileId) {
  return request({
    url: `/upload/file/${fileId}`,
    method: 'delete'
  })
}
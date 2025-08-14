<template>
  <div class="upload-container">
    <el-card class="upload-card">
      <template #header>
        <div class="card-header">
          <h2>文件上传</h2>
          <span>支持大文件上传和断点续传</span>
        </div>
      </template>
      
      <div class="upload-area" @drop.prevent="handleDrop" @dragover.prevent="handleDragOver" @dragleave.prevent="handleDragLeave" :class="{ 'is-dragover': isDragover }">
        <div v-if="!currentFile">
          <el-icon class="upload-icon"><Upload /></el-icon>
          <div class="upload-text">拖拽文件到此处或 <el-button type="primary" @click="handleClick">选择文件</el-button></div>
          <input ref="fileInput" type="file" style="display: none" @change="handleFileChange" />
        </div>
        <div v-else class="file-info">
          <div class="file-header">
            <div class="file-name">{{ currentFile.name }}</div>
            <div class="file-size">{{ formatFileSize(currentFile.size) }}</div>
          </div>
          
          <div v-if="hashProgress > 0 && hashProgress < 100" class="hash-progress">
            <span>计算文件哈希: {{ hashProgress }}%</span>
            <el-progress :percentage="hashProgress" />
          </div>
          
          <div v-if="uploadStatus" class="upload-progress">
            <div class="progress-header">
              <span>上传进度: {{ uploadStatus.progress }}%</span>
              <span>{{ formatFileSize(uploadedSize) }} / {{ formatFileSize(currentFile.size) }}</span>
            </div>
            <el-progress :percentage="uploadStatus.progress" />
            
            <div class="chunk-progress">
              <div v-for="i in uploadStatus.totalChunks" :key="i" class="chunk" :class="{ 'uploaded': uploadStatus.uploaded.includes(i - 1) }"></div>
            </div>
          </div>
          
          <div class="upload-actions">
            <el-button v-if="!uploading && !uploadStatus" type="primary" @click="startUpload" :disabled="hashProgress < 100">开始上传</el-button>
            <el-button v-if="uploading && !uploadPaused" type="warning" @click="pauseUpload">暂停上传</el-button>
            <el-button v-if="uploadPaused" type="primary" @click="resumeUpload">继续上传</el-button>
            <el-button v-if="uploadStatus && uploadStatus.progress === 100" type="success" disabled>上传完成</el-button>
            <el-button type="danger" @click="cancelUpload">取消上传</el-button>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Upload } from '@element-plus/icons-vue'
import { initUpload, uploadChunk as uploadChunkApi, completeUpload, getUploadStatus } from '@/utils/api'
import { calculateFileHash, sliceFile, formatFileSize } from '@/utils/file'

export default {
  components: {
    Upload
  },
  name: 'FileUpload',
  setup() {
    const fileInput = ref(null)
    const currentFile = ref(null)
    const fileHash = ref('')
    const hashProgress = ref(0)
    const uploading = ref(false)
    const uploadPaused = ref(false)
    const uploadStatus = ref(null)
    const isDragover = ref(false)
    const uploadQueue = ref([])
    const currentChunkIndex = ref(0)
    const fileId = ref(null)
    const chunkSize = ref(5 * 1024 * 1024) // 5MB
    const maxConcurrentUploads = 3
    const activeUploads = ref(0)
    const uploadedChunks = ref([])
    const uploadedSize = computed(() => {
      if (!uploadStatus.value || !currentFile.value) return 0
      return uploadStatus.value.uploaded.length * chunkSize.value
    })

    // 处理文件选择
    const handleFileChange = (e) => {
      const file = e.target.files[0]
      if (!file) return
      
      currentFile.value = file
      fileHash.value = ''
      hashProgress.value = 0
      uploadStatus.value = null
      
      // 计算文件哈希
      calculateFileHash(file, (progress) => {
        hashProgress.value = progress
      }).then(hash => {
        fileHash.value = hash
        ElMessage.success('文件哈希计算完成')
      }).catch(err => {
        ElMessage.error('文件哈希计算失败: ' + err.message)
        currentFile.value = null
      })
    }

    // 点击选择文件
    const handleClick = () => {
      fileInput.value.click()
    }

    // 拖拽相关处理
    const handleDragOver = () => {
      isDragover.value = true
    }
    
    const handleDragLeave = () => {
      isDragover.value = false
    }
    
    const handleDrop = (e) => {
      isDragover.value = false
      if (e.dataTransfer.files.length > 0) {
        const file = e.dataTransfer.files[0]
        currentFile.value = file
        fileHash.value = ''
        hashProgress.value = 0
        uploadStatus.value = null
        
        // 计算文件哈希
        calculateFileHash(file, (progress) => {
          hashProgress.value = progress
        }).then(hash => {
          fileHash.value = hash
          ElMessage.success('文件哈希计算完成')
        }).catch(err => {
          ElMessage.error('文件哈希计算失败: ' + err.message)
          currentFile.value = null
        })
      }
    }

    // 开始上传
    const startUpload = async () => {
      if (!currentFile.value || !fileHash.value) return
      
      try {
        uploading.value = true
        uploadPaused.value = false
        
        // 初始化上传
        const initResult = await initUpload({
          file_name: currentFile.value.name,
          file_size: currentFile.value.size,
          file_hash: fileHash.value,
          content_type: currentFile.value.type
        })
        
        fileId.value = initResult.file_id
        chunkSize.value = initResult.chunk_size
        uploadedChunks.value = initResult.uploaded || []
        
        // 如果文件已经上传完成
        if (initResult.status === 'completed') {
          ElMessage.success('文件已存在，无需重新上传')
          await updateUploadStatus()
          uploading.value = false
          return
        }
        
        // 分割文件
        const chunks = sliceFile(currentFile.value, chunkSize.value)
        uploadQueue.value = chunks.map((chunk, index) => ({
          chunk,
          index,
          status: uploadedChunks.value.includes(index) ? 'uploaded' : 'pending'
        }))
        
        // 开始上传分片
        uploadNextChunks()
        
        // 定时更新上传状态
        const statusInterval = setInterval(async () => {
          if (!uploading.value) {
            clearInterval(statusInterval)
            return
          }
          await updateUploadStatus()
        }, 2000)
        
      } catch (error) {
        ElMessage.error('上传初始化失败: ' + error.message)
        uploading.value = false
      }
    }
    
    // 上传下一批分片
    const uploadNextChunks = () => {
      if (uploadPaused.value) return
      
      while (activeUploads.value < maxConcurrentUploads) {
        const pendingChunk = uploadQueue.value.find(item => item.status === 'pending')
        if (!pendingChunk) break
        
        pendingChunk.status = 'uploading'
        activeUploads.value++
        
        uploadChunk(pendingChunk)
      }
      
      // 检查是否所有分片都已上传
      const allUploaded = uploadQueue.value.every(item => item.status === 'uploaded')
      if (allUploaded && activeUploads.value === 0) {
        finishUpload()
      }
    }
    
    // 上传单个分片
    const uploadChunk = async (chunkInfo) => {
      try {
        const formData = new FormData()
        formData.append('file_id', fileId.value)
        formData.append('chunk_num', chunkInfo.index)
        formData.append('chunk', chunkInfo.chunk)
        
        await uploadChunkApi(formData, (progress) => {
          // 可以在这里处理单个分片的上传进度
          // 使用progress参数避免ESLint警告
          if (progress && progress.loaded) {
            // 可以在这里添加进度处理逻辑
          }
        })
        
        chunkInfo.status = 'uploaded'
        uploadedChunks.value.push(chunkInfo.index)
        
      } catch (error) {
        chunkInfo.status = 'error'
        ElMessage.error(`分片 ${chunkInfo.index + 1} 上传失败: ${error.message}`)
      } finally {
        activeUploads.value--
        uploadNextChunks()
      }
    }
    
    // 完成上传
    const finishUpload = async () => {
      try {
        await completeUpload({ file_id: fileId.value })
        ElMessage.success('文件上传完成')
        await updateUploadStatus()
      } catch (error) {
        ElMessage.error('文件合并失败: ' + error.message)
      } finally {
        uploading.value = false
      }
    }
    
    // 更新上传状态
    const updateUploadStatus = async () => {
      if (!fileId.value) return
      
      try {
        const status = await getUploadStatus(fileId.value)
        uploadStatus.value = status
      } catch (error) {
        console.error('获取上传状态失败:', error)
      }
    }
    
    // 暂停上传
    const pauseUpload = () => {
      uploadPaused.value = true
      ElMessage.info('上传已暂停')
    }
    
    // 继续上传
    const resumeUpload = () => {
      uploadPaused.value = false
      uploadNextChunks()
      ElMessage.info('上传已继续')
    }
    
    // 取消上传
    const cancelUpload = () => {
      uploading.value = false
      uploadPaused.value = false
      currentFile.value = null
      fileHash.value = ''
      hashProgress.value = 0
      uploadStatus.value = null
      uploadQueue.value = []
      currentChunkIndex.value = 0
      fileId.value = null
      activeUploads.value = 0
      uploadedChunks.value = []
      ElMessage.info('上传已取消')
    }

    return {
      fileInput,
      currentFile,
      hashProgress,
      uploading,
      uploadPaused,
      uploadStatus,
      isDragover,
      uploadedSize,
      handleFileChange,
      handleClick,
      handleDragOver,
      handleDragLeave,
      handleDrop,
      startUpload,
      pauseUpload,
      resumeUpload,
      cancelUpload,
      formatFileSize
    }
  }
}
</script>

<style scoped>
.upload-container {
  max-width: 800px;
  margin: 0 auto;
}

.upload-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h2 {
  margin: 0;
  font-size: 1.5rem;
}

.upload-area {
  border: 2px dashed #d9d9d9;
  border-radius: 6px;
  padding: 40px;
  text-align: center;
  transition: all 0.3s;
  cursor: pointer;
}

.upload-area.is-dragover {
  border-color: #409EFF;
  background-color: rgba(64, 158, 255, 0.1);
}

.upload-icon {
  font-size: 48px;
  color: #909399;
  margin-bottom: 20px;
}

.upload-text {
  color: #606266;
  font-size: 16px;
  margin-bottom: 20px;
}

.file-info {
  width: 100%;
}

.file-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 15px;
}

.file-name {
  font-weight: bold;
  font-size: 16px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hash-progress, .upload-progress {
  margin-bottom: 20px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 5px;
}

.chunk-progress {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 10px;
}

.chunk {
  width: 16px;
  height: 16px;
  background-color: #f0f0f0;
  border-radius: 2px;
}

.chunk.uploaded {
  background-color: #67C23A;
}

.upload-actions {
  margin-top: 20px;
  display: flex;
  justify-content: center;
  gap: 10px;
}
</style>
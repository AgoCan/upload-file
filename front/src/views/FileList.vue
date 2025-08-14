<template>
  <div class="file-list-container">
    <el-card class="file-list-card">
      <template #header>
        <div class="card-header">
          <h2>文件列表</h2>
          <el-button type="primary" @click="refreshList">刷新列表</el-button>
        </div>
      </template>
      
      <el-table v-loading="loading" :data="fileList" style="width: 100%">
        <el-table-column label="文件名" prop="file_name" min-width="200">
          <template #default="{ row }">
            <div class="file-name-cell">
              <el-icon><Document /></el-icon>
              <span>{{ row.file_name }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="大小" prop="file_size" width="120">
          <template #default="{ row }">
            {{ formatFileSize(row.file_size) }}
          </template>
        </el-table-column>
        
        <el-table-column label="上传时间" prop="created_at" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="downloadFile(row.id)">下载</el-button>
            <el-popconfirm title="确定要删除这个文件吗？" @confirm="deleteFile(row.id)">
              <template #reference>
                <el-button type="danger" size="small">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      
      <div v-if="fileList.length === 0 && !loading" class="empty-list">
        <el-empty description="暂无文件" />
      </div>
    </el-card>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Document } from '@element-plus/icons-vue'
import { getFileList, downloadFile as apiDownloadFile, deleteFile as apiDeleteFile } from '@/utils/api'
import { formatFileSize } from '@/utils/file'

export default {
  name: 'FileList',
  components: {
    Document
  },
  setup() {
    const fileList = ref([])
    const loading = ref(false)
    
    // 获取文件列表
    const fetchFileList = async () => {
      loading.value = true
      try {
        fileList.value = await getFileList()
      } catch (error) {
        ElMessage.error('获取文件列表失败: ' + error.message)
      } finally {
        loading.value = false
      }
    }
    
    // 刷新列表
    const refreshList = () => {
      fetchFileList()
    }
    
    // 下载文件
    const downloadFile = (fileId) => {
      apiDownloadFile(fileId)
    }
    
    // 删除文件
    const deleteFile = async (fileId) => {
      try {
        await apiDeleteFile(fileId)
        ElMessage.success('文件删除成功')
        fetchFileList()
      } catch (error) {
        ElMessage.error('文件删除失败: ' + error.message)
      }
    }
    
    // 格式化日期
    const formatDate = (dateString) => {
      const date = new Date(dateString)
      return date.toLocaleString()
    }
    
    onMounted(() => {
      fetchFileList()
    })
    
    return {
      fileList,
      loading,
      refreshList,
      downloadFile,
      deleteFile,
      formatFileSize,
      formatDate
    }
  }
}
</script>

<style scoped>
.file-list-container {
  max-width: 1000px;
  margin: 0 auto;
}

.file-list-card {
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

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.empty-list {
  padding: 40px 0;
}
</style>
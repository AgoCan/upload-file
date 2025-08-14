<template>
  <div class="app-container">
    <el-container>
      <el-header>
        <div class="header-content">
          <h1>文件上传系统</h1>
          <el-menu mode="horizontal" :router="true" :default-active="activeMenu">
            <el-menu-item index="/upload">上传文件</el-menu-item>
            <el-menu-item index="/files">文件列表</el-menu-item>
          </el-menu>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
      <el-footer>
        <p>© 2025 文件上传系统 - 支持大文件上传和断点续传</p>
      </el-footer>
    </el-container>
  </div>
</template>

<script>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

export default {
  name: 'App',
  setup() {
    const route = useRoute()
    
    const activeMenu = computed(() => {
      return route.path
    })
    
    return {
      activeMenu
    }
  }
}
</script>

<style>
body {
  margin: 0;
  padding: 0;
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', Arial, sans-serif;
}

.app-container {
  min-height: 100vh;
}

.el-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.el-header {
  background-color: #409EFF;
  color: white;
  padding: 0;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 100%;
}

.header-content h1 {
  margin: 0;
  font-size: 1.5rem;
}

.el-menu {
  background-color: transparent;
  border-bottom: none;
}

.el-menu--horizontal > .el-menu-item {
  color: white;
  border-bottom: none;
}

.el-menu--horizontal > .el-menu-item.is-active {
  color: #fff;
  font-weight: bold;
  border-bottom: 2px solid #fff;
  background-color: rgba(255, 255, 255, 0.1);
}

.el-main {
  flex: 1;
  padding: 20px;
  background-color: #f5f7fa;
}

.el-footer {
  text-align: center;
  background-color: #f5f7fa;
  color: #909399;
  padding: 15px 0;
  font-size: 0.9rem;
}
</style>
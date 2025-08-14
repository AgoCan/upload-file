import { createRouter, createWebHistory } from 'vue-router'
import FileUpload from '../views/FileUpload.vue'
import FileList from '../views/FileList.vue'

const routes = [
  {
    path: '/',
    redirect: '/upload'
  },
  {
    path: '/upload',
    name: 'upload',
    component: FileUpload
  },
  {
    path: '/files',
    name: 'files',
    component: FileList
  }
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes
})

export default router
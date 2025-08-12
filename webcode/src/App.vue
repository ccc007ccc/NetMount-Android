<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'

// --- 响应式状态 ---
const config = ref({ mounts: [] })
const status = ref({})
const logs = ref([])
const precheckLogs = ref('')
const isLoading = ref(true)
const activeTab = ref('mounts') // 'mounts' or 'logs' or 'system'

const newMount = ref({
  name: '',
  type: 'webdav',
  remote: '',
  mountPoint: '/data/media/0/mnt/',
  authType: 'none',
  user: '',
  pass: ''
})
const isMountPointManuallyEdited = ref(false)

// 编辑挂载点的状态
const editingMount = ref(null)
const editMount = ref({
  name: '',
  type: 'webdav',
  remote: '',
  mountPoint: '/data/media/0/mnt/',
  authType: 'none',
  user: '',
  pass: ''
})

let pollInterval = null

// --- 侦听器 ---
watch(() => newMount.value.name, (newName) => {
  if (!isMountPointManuallyEdited.value) {
    newMount.value.mountPoint = `/data/media/0/mnt/${newName}`
  }
})

watch(() => newMount.value.mountPoint, () => {
  isMountPointManuallyEdited.value = true
})


// --- API 调用函数 ---
async function apiRequest(url, options = {}, isText = false) {
  try {
    const response = await fetch(url, options)
    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(`请求失败: ${response.status} ${errorText || response.statusText}`)
    }
    if (isText) {
      return response.text()
    }
    const contentType = response.headers.get("content-type")
    if (contentType && contentType.includes("application/json")) {
      return response.json()
    }
    return response.text()
  } catch (error) {
    console.error(`API 请求错误 ${url}:`, error)
    // 在这里可以添加一个全局的错误通知
    throw error
  }
}

// --- 核心逻辑 ---
async function loadData() {
  try {
    // Pre-check logs are loaded once, others are polled
    const [configData, statusData, logsData] = await Promise.all([
      apiRequest('/api/config'),
      apiRequest('/api/status'),
      apiRequest('/api/logs')
    ])
    config.value = configData || { mounts: [] }
    status.value = statusData || {}
    logs.value = logsData || []
  } catch (error) {
    console.error("轮询数据失败:", error)
  }
}

async function loadPrecheckData() {
  try {
    precheckLogs.value = await apiRequest('/api/precheck', {}, true);
  } catch (error) {
    precheckLogs.value = "加载系统检查日志失败: " + error.message;
  }
}


async function saveConfig() {
  try {
    await apiRequest('/api/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(config.value)
    })
    await loadData() // 保存后立即刷新数据
  } catch (error) {
    alert('保存配置失败!')
  }
}

async function handleAddMount() {
  if (!newMount.value.name || !newMount.value.remote || !newMount.value.mountPoint) {
    alert('请填写所有必填项！')
    return
  }
  
  const mountToAdd = {
    name: newMount.value.name,
    type: newMount.value.type,
    remote: newMount.value.remote,
    mountPoint: newMount.value.mountPoint,
    parameters: {}
  };

  if (newMount.value.type === 'webdav' || newMount.value.type === 'ftp' || newMount.value.type === 'smb') {
    mountToAdd.authType = newMount.value.authType;
    if (newMount.value.authType === 'password') {
      mountToAdd.user = newMount.value.user;
      mountToAdd.pass = newMount.value.pass;
    }
  }

  config.value.mounts.push(mountToAdd)
  await saveConfig()
  // 重置表单
  newMount.value = { 
    name: '', 
    type: 'webdav', 
    remote: '', 
    mountPoint: '/data/media/0/mnt/',
    authType: 'none',
    user: '',
    pass: ''
  }
  isMountPointManuallyEdited.value = false
}

async function handleDeleteMount(mountNameToDelete) {
  if (!confirm(`确定要删除挂载点 "${mountNameToDelete}" 吗? 这会先尝试卸载它。`)) return

  const mountStatus = status.value[mountNameToDelete]
  if (mountStatus && mountStatus.isMounted) {
    await handleUnmount(mountNameToDelete)
  }

  config.value.mounts = config.value.mounts.filter(m => m.name !== mountNameToDelete)
  await saveConfig()
}

function handleEditMount(mount) {
  editingMount.value = mount.name
  editMount.value = {
    name: mount.name,
    type: mount.type,
    remote: mount.remote,
    mountPoint: mount.mountPoint,
    authType: mount.authType || 'none',
    user: mount.user || '',
    pass: mount.pass || ''
  }
}

function cancelEdit() {
  editingMount.value = null
  editMount.value = {
    name: '',
    type: 'webdav',
    remote: '',
    mountPoint: '/data/media/0/mnt/',
    authType: 'none',
    user: '',
    pass: ''
  }
}

async function saveEdit() {
  if (!editMount.value.name || !editMount.value.remote || !editMount.value.mountPoint) {
    alert('请填写所有必填项！')
    return
  }

  // 如果正在挂载，先卸载
  const mountStatus = status.value[editingMount.value]
  if (mountStatus && mountStatus.isMounted) {
    await handleUnmount(editingMount.value)
  }

  // 找到要编辑的挂载点并更新
  const mountIndex = config.value.mounts.findIndex(m => m.name === editingMount.value)
  if (mountIndex !== -1) {
    const updatedMount = {
      name: editMount.value.name,
      type: editMount.value.type,
      remote: editMount.value.remote,
      mountPoint: editMount.value.mountPoint,
      parameters: config.value.mounts[mountIndex].parameters || {}
    }

    if (editMount.value.type === 'webdav' || editMount.value.type === 'ftp' || editMount.value.type === 'smb') {
      updatedMount.authType = editMount.value.authType;
      if (editMount.value.authType === 'password') {
        updatedMount.user = editMount.value.user;
        updatedMount.pass = editMount.value.pass;
      }
    }

    config.value.mounts[mountIndex] = updatedMount
    await saveConfig()
    cancelEdit()
  }
}

async function handleMount(mountName) {
  try {
    await apiRequest(`/api/mount/${mountName}`, { method: 'POST' })
    setTimeout(loadData, 500) // 给后端一点时间处理
  } catch (error) {
    alert(`挂载操作失败: ${error.message}`)
  }
}

// 获取不同类型的远程地址占位符
function getRemotePlaceholder(type) {
  switch (type) {
    case 'smb':
      return '192.168.1.100/sharename'
    case 'ftp':
      return '192.168.1.100/path'
    case 'webdav':
      return '192.168.1.100/webdav'
    case 'sftp':
      return '192.168.1.100/path'
    default:
      return 'server-address/path'
  }
}

// 获取不同类型的提示信息
function getRemoteHint(type) {
  switch (type) {
    case 'smb':
      return 'SMB 格式: IP地址/共享名，例如: 192.168.1.100/Public'
    case 'ftp':
      return 'FTP 格式: IP地址/路径，例如: 192.168.1.100/uploads'
    case 'webdav':
      return 'WebDAV 格式: IP地址/路径，例如: 192.168.1.100/webdav'
    case 'sftp':
      return 'SFTP 格式: IP地址/路径，例如: 192.168.1.100/home/user'
    default:
      return '请输入服务器地址和路径'
  }
}

async function handleUnmount(mountName) {
  try {
    await apiRequest(`/api/unmount/${mountName}`, { method: 'POST' })
    setTimeout(loadData, 500)
  } catch (error) {
    alert(`卸载操作失败: ${error.message}`)
  }
}

// --- 生命周期钩子 ---
onMounted(async () => {
  isLoading.value = true;
  await Promise.all([
    loadData(),
    loadPrecheckData()
  ]);
  isLoading.value = false;
  pollInterval = setInterval(loadData, 5000) // 每5秒轮询一次
})

onUnmounted(() => {
  clearInterval(pollInterval) // 组件卸载时停止轮询
})
</script>

<template>
  <header class="header">
    <h1>网络挂载 (NetMount) 控制面板</h1>
    <nav>
      <button @click="activeTab = 'mounts'" :class="{ active: activeTab === 'mounts' }">挂载管理</button>
      <button @click="activeTab = 'logs'" :class="{ active: activeTab === 'logs' }">守护进程日志</button>
      <button @click="activeTab = 'system'" :class="{ active: activeTab === 'system' }">系统检查</button>
    </nav>
  </header>

  <main>
    <div v-if="isLoading">正在加载...</div>
    
    <div v-else>
      <!-- 挂载管理 Tab -->
      <div v-show="activeTab === 'mounts'" class="mounts-panel">
        <h2>当前挂载点</h2>
        <ul class="mount-list">
          <li v-if="config.mounts.length === 0">没有已配置的挂载点。</li>
          <li v-for="mount in config.mounts" :key="mount.name" class="mount-item" :class="{
            'status-mounted': status[mount.name]?.isMounted,
            'status-error': status[mount.name]?.error
          }">
            <div class="mount-details">
              <strong>{{ mount.name }}</strong> ({{ mount.type }})
              <small>{{ mount.remote }} &rarr; {{ mount.mountPoint }}</small>
              <small v-if="mount.authType && mount.authType !== 'none'">
                认证: {{ mount.authType }}
                <span v-if="mount.authType === 'password'"> (用户: {{ mount.user }})</span>
              </small>
              <small class="status-text" :class="{
                'text-success': status[mount.name]?.isMounted,
                'text-danger': status[mount.name]?.error
              }">
                状态: 
                <span v-if="status[mount.name]?.isMounted">已挂载 (PID: {{ status[mount.name]?.pid || 'N/A' }})</span>
                <span v-else-if="status[mount.name]?.error">错误: {{ status[mount.name]?.error }}</span>
                <span v-else>未挂载</span>
              </small>
            </div>
            <div class="mount-actions">
              <button v-if="!status[mount.name]?.isMounted" @click="handleMount(mount.name)" class="btn btn-success">挂载</button>
              <button v-if="status[mount.name]?.isMounted" @click="handleUnmount(mount.name)" class="btn btn-secondary">卸载</button>
              <button @click="handleEditMount(mount)" class="btn btn-primary">编辑</button>
              <button @click="handleDeleteMount(mount.name)" class="btn btn-danger">删除</button>
            </div>
          </li>
        </ul>

        <!-- 编辑挂载对话框 -->
        <div v-if="editingMount" class="edit-modal">
          <div class="edit-modal-content">
            <h3>编辑挂载点: {{ editingMount }}</h3>
            <form @submit.prevent="saveEdit" class="edit-form">
              <div class="form-grid">
                <label for="edit-name">名称:</label>
                <input type="text" id="edit-name" v-model="editMount.name" placeholder="无空格的唯一名称" required pattern="^\S+$">

                <label for="edit-type">类型:</label>
                <select id="edit-type" v-model="editMount.type">
                  <option value="webdav">WebDAV</option>
                  <option value="ftp">FTP</option>
                  <option value="sftp">SFTP</option>
                  <option value="smb">SMB</option>
                </select>

                <label for="edit-remote">远程地址:</label>
                <input type="text" id="edit-remote" v-model="editMount.remote" 
                       :placeholder="getRemotePlaceholder(editMount.type)" required>
                <div class="remote-hint">{{ getRemoteHint(editMount.type) }}</div>

                <label for="edit-mountPoint">挂载路径:</label>
                <input type="text" id="edit-mountPoint" v-model="editMount.mountPoint" placeholder="e.g., /mnt/nas" required>

                <template v-if="editMount.type === 'webdav' || editMount.type === 'ftp' || editMount.type === 'smb'">
                  <label for="edit-authType">认证方式:</label>
                  <select id="edit-authType" v-model="editMount.authType">
                    <option value="none">无</option>
                    <option v-if="editMount.type === 'ftp'" value="anonymous">匿名</option>
                    <option value="password">密码</option>
                  </select>
                </template>

                <template v-if="editMount.authType === 'password' && (editMount.type === 'webdav' || editMount.type === 'ftp' || editMount.type === 'smb')">
                  <label for="edit-user">用户名:</label>
                  <input type="text" id="edit-user" v-model="editMount.user" placeholder="请输入用户名">
                  <label for="edit-pass">密码:</label>
                  <input type="password" id="edit-pass" v-model="editMount.pass" placeholder="请输入密码">
                </template>
              </div>
              <div class="edit-modal-actions">
                <button type="submit" class="btn btn-success">保存</button>
                <button type="button" @click="cancelEdit" class="btn btn-secondary">取消</button>
              </div>
            </form>
          </div>
        </div>

        <h2>添加新挂载</h2>
        <form @submit.prevent="handleAddMount" class="add-form">
          <div class="form-grid">
            <label for="name">名称:</label>
            <input type="text" id="name" v-model="newMount.name" placeholder="无空格的唯一名称" required pattern="^\S+$">

            <label for="type">类型:</label>
            <select id="type" v-model="newMount.type">
              <option value="webdav">WebDAV</option>
              <option value="ftp">FTP</option>
              <option value="sftp">SFTP</option>
              <option value="smb">SMB</option>
            </select>

            <label for="remote">远程地址:</label>
            <input type="text" id="remote" v-model="newMount.remote" 
                   :placeholder="getRemotePlaceholder(newMount.type)" required>
            <div class="remote-hint">{{ getRemoteHint(newMount.type) }}</div>

            <label for="mountPoint">挂载路径:</label>
            <input type="text" id="mountPoint" v-model="newMount.mountPoint" placeholder="e.g., /mnt/nas" required>

            <template v-if="newMount.type === 'webdav' || newMount.type === 'ftp' || newMount.type === 'smb'">
              <label for="authType">认证方式:</label>
              <select id="authType" v-model="newMount.authType">
                <option value="none">无</option>
                <option v-if="newMount.type === 'ftp'" value="anonymous">匿名</option>
                <option value="password">密码</option>
              </select>
            </template>

            <template v-if="newMount.authType === 'password' && (newMount.type === 'webdav' || newMount.type === 'ftp' || newMount.type === 'smb')">
              <label for="user">用户名:</label>
              <input type="text" id="user" v-model="newMount.user" placeholder="请输入用户名">
              <label for="pass">密码:</label>
              <input type="password" id="pass" v-model="newMount.pass" placeholder="请输入密码">
            </template>
          </div>
          <button type="submit" class="btn-primary">添加并保存</button>
        </form>
      </div>

      <!-- 日志 Tab -->
      <div v-show="activeTab === 'logs'" class="logs-panel">
        <h2>守护进程日志</h2>
        <div class="log-viewer">
          <pre v-if="logs.length > 0">{{ logs.join('\n') }}</pre>
          <span v-else>暂无日志。</span>
        </div>
      </div>

      <!-- 系统检查 Tab -->
      <div v-show="activeTab === 'system'" class="logs-panel">
        <h2>系统环境检查 (模块启动时)</h2>
        <div class="log-viewer">
          <pre>{{ precheckLogs }}</pre>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
/* 基础样式 */
.header {
  background-color: #24292e;
  color: white;
  padding: 1rem;
  margin-bottom: 1rem;
  border-radius: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}
.header h1 { 
  margin: 0; 
  font-size: 1.5rem;
  flex: 1;
  min-width: 200px;
}

nav { 
  display: flex; 
  gap: 0.5rem; 
  flex-wrap: wrap;
}
nav button {
  background: #444;
  color: white;
  border: none;
  padding: 0.75rem 1rem;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  min-height: 44px; /* iOS 推荐的最小触摸目标 */
  min-width: 80px;
  white-space: nowrap;
}
nav button.active {
  background: #0366d6;
}
nav button:hover {
  background: #555;
}
nav button.active:hover {
  background: #0251b8;
}

.mounts-panel, .logs-panel {
  background: #fff;
  padding: 1rem;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

h2 {
  margin-top: 0;
  border-bottom: 1px solid #e1e4e8;
  padding-bottom: 0.5rem;
  margin-bottom: 1rem;
  font-size: 1.3rem;
}

/* Mount List - 移动端优化 */
.mount-list { 
  list-style: none; 
  padding: 0; 
}
.mount-item {
  padding: 1rem;
  border-radius: 6px;
  margin-bottom: 1rem;
  border: 1px solid #d0d7de;
  border-left-width: 5px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.mount-item.status-mounted { border-left-color: #28a745; }
.mount-item.status-error { border-left-color: #dc3545; }

.mount-details { 
  display: flex; 
  flex-direction: column; 
  gap: 0.5rem; 
  flex: 1;
}
.mount-details strong { 
  font-size: 1.2rem; 
  color: #24292e;
}
.mount-details small { 
  color: #57606a; 
  line-height: 1.4;
  word-break: break-all;
}
.status-text { 
  font-weight: bold; 
  font-size: 0.9rem;
}
.text-success { color: #28a745; }
.text-danger { color: #dc3545; }

.mount-actions { 
  display: flex; 
  gap: 0.5rem; 
  flex-wrap: wrap;
  justify-content: stretch;
}

/* Forms - 移动端优化 */
.add-form .form-grid,
.edit-form .form-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.add-form label,
.edit-form label { 
  font-weight: 600; 
  margin-bottom: 0.25rem;
  display: block;
}

.add-form input, 
.add-form select,
.edit-form input,
.edit-form select {
  padding: 0.75rem;
  border-radius: 6px;
  border: 1px solid #d0d7de;
  width: 100%;
  box-sizing: border-box;
  font-size: 1rem;
  min-height: 44px; /* iOS 推荐的最小触摸目标 */
}

.add-form input:focus,
.add-form select:focus,
.edit-form input:focus,
.edit-form select:focus {
  outline: none;
  border-color: #0366d6;
  box-shadow: 0 0 0 3px rgba(3, 102, 214, 0.1);
}

.remote-hint {
  font-size: 0.8rem;
  color: #656d76;
  margin-top: -0.5rem;
  margin-bottom: 0.5rem;
  line-height: 1.4;
}

/* Buttons - 移动端优化 */
.btn, button {
  padding: 0.75rem 1.2rem;
  font-size: 0.9rem;
  border-radius: 6px;
  border: 1px solid rgba(27, 31, 36, 0.15);
  cursor: pointer;
  font-weight: 600;
  min-height: 44px; /* iOS 推荐的最小触摸目标 */
  white-space: nowrap;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  text-decoration: none;
  transition: all 0.15s ease;
}

.btn-primary { 
  background-color: #2da44e; 
  color: white; 
  border-color: #2da44e;
}
.btn-primary:hover {
  background-color: #2c974b;
}

.btn-success { 
  background-color: #2da44e; 
  color: white; 
  border-color: #2da44e;
}
.btn-success:hover {
  background-color: #2c974b;
}

.btn-secondary { 
  background-color: #f6f8fa; 
  color: #24292e; 
  border-color: #d0d7de;
}
.btn-secondary:hover {
  background-color: #f3f4f6;
}

.btn-danger { 
  background-color: #d73a49; 
  color: white; 
  border-color: #d73a49;
}
.btn-danger:hover {
  background-color: #cb2431;
}

/* Log Viewer */
.log-viewer {
  background-color: #f6f8fa;
  border: 1px solid #d0d7de;
  border-radius: 6px;
  padding: 1rem;
  height: 400px;
  overflow-y: scroll;
  font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, Courier, monospace;
  font-size: 0.85rem;
  -webkit-overflow-scrolling: touch; /* iOS 平滑滚动 */
}
.log-viewer pre { 
  margin: 0; 
  white-space: pre-wrap; 
  word-wrap: break-word; 
  line-height: 1.4;
}

/* 编辑对话框 - 移动端优化 */
.edit-modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: flex-start;
  z-index: 1000;
  padding: 1rem;
  box-sizing: border-box;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}

.edit-modal-content {
  background: white;
  padding: 1.5rem;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
  width: 100%;
  max-width: 500px;
  max-height: none;
  margin: 2rem 0;
}

.edit-modal-content h3 {
  margin-top: 0;
  margin-bottom: 1.5rem;
  color: #24292e;
  font-size: 1.2rem;
  word-break: break-word;
}

.edit-modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: stretch;
  margin-top: 1.5rem;
}

.edit-modal-actions .btn {
  flex: 1;
}

/* 移动端优化 */
@media (max-width: 768px) {
  .header {
    flex-direction: column;
    text-align: center;
    padding: 1rem;
  }
  
  .header h1 {
    font-size: 1.3rem;
    min-width: auto;
  }
  
  nav {
    width: 100%;
    justify-content: center;
  }
  
  nav button {
    flex: 1;
    min-width: 0;
    font-size: 0.85rem;
    padding: 0.75rem 0.5rem;
  }
  
  .mounts-panel, .logs-panel {
    padding: 1rem;
    border-radius: 6px;
    margin: 0 -0.5rem;
  }
  
  .mount-item {
    padding: 1rem;
    margin-bottom: 1rem;
  }
  
  .mount-actions {
    gap: 0.5rem;
  }
  
  .mount-actions .btn {
    flex: 1;
    min-width: 0;
    font-size: 0.85rem;
    padding: 0.75rem 0.5rem;
  }
  
  .log-viewer {
    height: 300px;
    font-size: 0.8rem;
  }
  
  .edit-modal {
    align-items: stretch;
    padding: 0;
  }
  
  .edit-modal-content {
    margin: 0;
    border-radius: 0;
    min-height: 100vh;
    display: flex;
    flex-direction: column;
  }
  
  .edit-form {
    flex: 1;
  }
  
  h2 {
    font-size: 1.2rem;
  }
}

/* 小屏幕设备优化 */
@media (max-width: 480px) {
  .header {
    padding: 0.75rem;
  }
  
  .header h1 {
    font-size: 1.2rem;
  }
  
  nav button {
    font-size: 0.8rem;
    padding: 0.6rem 0.4rem;
  }
  
  .mounts-panel, .logs-panel {
    padding: 0.75rem;
  }
  
  .mount-item {
    padding: 0.75rem;
  }
  
  .mount-details strong {
    font-size: 1.1rem;
  }
  
  .mount-details small {
    font-size: 0.8rem;
  }
  
  .mount-actions .btn {
    font-size: 0.8rem;
    padding: 0.6rem 0.4rem;
  }
  
  .add-form input, 
  .add-form select,
  .edit-form input,
  .edit-form select {
    padding: 0.6rem;
    font-size: 0.9rem;
  }
  
  .log-viewer {
    height: 250px;
    font-size: 0.75rem;
    padding: 0.75rem;
  }
  
  .edit-modal-content {
    padding: 1rem;
  }
  
  h2 {
    font-size: 1.1rem;
  }
}

/* 横屏平板优化 */
@media (min-width: 769px) and (max-width: 1024px) {
  .add-form .form-grid,
  .edit-form .form-grid {
    grid-template-columns: 120px 1fr;
    align-items: center;
  }
  
  .add-form label,
  .edit-form label {
    margin-bottom: 0;
  }
  
  .remote-hint {
    grid-column: 2;
  }
  
  .mount-item {
    flex-direction: row;
    align-items: center;
  }
  
  .mount-details {
    flex: 1;
  }
  
  .mount-actions {
    flex-shrink: 0;
    flex-wrap: nowrap;
  }
}

/* 大屏幕优化 */
@media (min-width: 1025px) {
  .add-form .form-grid,
  .edit-form .form-grid {
    grid-template-columns: 130px 1fr;
    align-items: center;
  }
  
  .add-form label,
  .edit-form label {
    margin-bottom: 0;
  }
  
  .remote-hint {
    grid-column: 2;
  }
  
  .mount-item {
    flex-direction: row;
    align-items: center;
  }
  
  .mount-details {
    flex: 1;
  }
  
  .mount-actions {
    flex-shrink: 0;
    flex-wrap: nowrap;
  }
  
  .edit-modal {
    align-items: center;
  }
  
  .edit-modal-content {
    margin: 0;
    max-height: 90vh;
    overflow-y: auto;
  }
  
  .edit-modal-actions {
    justify-content: flex-end;
  }
  
  .edit-modal-actions .btn {
    flex: none;
    min-width: 100px;
  }
}
</style>

# API 文档

## API 概述

NetMount-Android 提供 RESTful API 接口，用于管理网络存储挂载配置和状态。Web UI 通过这些 API 与后端守护进程通信。

## 基础信息

- **基础 URL**: `http://localhost:8088/api`
- **数据格式**: JSON
- **认证**: 无需认证（本地访问）

## API 端点

### 1. 配置管理

#### 获取配置

```http
GET /api/config
```

**响应示例：**
```json
{
  "mounts": [
    {
      "name": "MyWebDAV",
      "type": "webdav",
      "remote": "https://cloud.example.com/remote.php/dav/files/user/",
      "mountPoint": "/sdcard/NetMount/WebDAV",
      "parameters": {},
      "authType": "password",
      "user": "username",
      "pass": "password"
    }
  ]
}
```

#### 更新配置

```http
POST /api/config
Content-Type: application/json
```

**请求体：**
```json
{
  "mounts": [
    {
      "name": "MyWebDAV",
      "type": "webdav",
      "remote": "https://cloud.example.com/remote.php/dav/files/user/",
      "mountPoint": "/sdcard/NetMount/WebDAV",
      "parameters": {},
      "authType": "password",
      "user": "username",
      "pass": "password"
    }
  ]
}
```

### 2. 挂载状态

#### 获取所有挂载状态

```http
GET /api/status
```

**响应示例：**
```json
{
  "MyWebDAV": {
    "name": "MyWebDAV",
    "isMounted": true,
    "error": "",
    "pid": 12345
  },
  "MySMB": {
    "name": "MySMB",
    "isMounted": false,
    "error": "Connection failed",
    "pid": 0
  }
}
```

### 3. 挂载操作

#### 挂载单个存储

```http
POST /api/mount/{name}
```

**参数：**
- `name`: 挂载配置名称（URL 路径参数）

**响应：**
- `202 Accepted`: 挂载请求已接受
- `404 Not Found`: 未找到指定配置

#### 卸载单个存储

```http
POST /api/unmount/{name}
```

**参数：**
- `name`: 挂载配置名称（URL 路径参数）

**响应：**
- `202 Accepted`: 卸载请求已接受
- `400 Bad Request`: 卸载失败（含错误信息）

### 4. 日志查看

#### 获取实时日志

```http
GET /api/logs
```

**响应示例：**
```json
[
  "2024-01-01 12:00:00 - --- NetMount 守护进程启动 ---",
  "2024-01-01 12:00:01 - 使用配置文件: /data/adb/netmount/config.json",
  "2024-01-01 12:00:02 - ✅ 成功启动 'MyWebDAV' 的挂载进程, PID: 12345"
]
```

### 5. 环境检查

#### 获取环境检查日志

```http
GET /api/precheck
```

**响应：**
- Content-Type: `text/plain; charset=utf-8`
- 返回环境检查的详细日志文本

## 数据模型

### MountConfig

挂载配置对象：

```typescript
interface MountConfig {
  name: string;              // 挂载名称（唯一标识）
  type: string;              // 挂载类型：webdav, smb, ftp
  remote: string;            // 远程地址
  mountPoint: string;        // 本地挂载点路径
  parameters: {[key: string]: string}; // 额外参数
  authType?: string;         // 认证类型：password, anonymous
  user?: string;             // 用户名
  pass?: string;             // 密码
}
```

### MountStatus

挂载状态对象：

```typescript
interface MountStatus {
  name: string;              // 挂载名称
  isMounted: boolean;        // 是否已挂载
  error?: string;            // 错误信息（如果有）
  pid?: number;              // 进程 ID（如果已挂载）
}
```

## 错误处理

### HTTP 状态码

- `200 OK`: 请求成功
- `202 Accepted`: 异步操作已接受
- `400 Bad Request`: 请求参数错误
- `404 Not Found`: 资源未找到
- `405 Method Not Allowed`: 不支持的 HTTP 方法
- `500 Internal Server Error`: 服务器内部错误

### 错误响应格式

```json
{
  "error": "错误描述信息"
}
```

## 使用示例

### JavaScript/Fetch API

```javascript
// 获取挂载状态
async function getMountStatus() {
  const response = await fetch('/api/status');
  const status = await response.json();
  return status;
}

// 添加新的挂载配置
async function addMount(mountConfig) {
  const response = await fetch('/api/config');
  const config = await response.json();
  
  config.mounts.push(mountConfig);
  
  const updateResponse = await fetch('/api/config', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(config)
  });
  
  return updateResponse.ok;
}

// 挂载存储
async function mountStorage(name) {
  const response = await fetch(`/api/mount/${name}`, {
    method: 'POST'
  });
  return response.status === 202;
}
```

### cURL 示例

```bash
# 获取配置
curl -X GET http://localhost:8088/api/config

# 获取状态
curl -X GET http://localhost:8088/api/status

# 挂载存储
curl -X POST http://localhost:8088/api/mount/MyWebDAV

# 卸载存储
curl -X POST http://localhost:8088/api/unmount/MyWebDAV

# 获取日志
curl -X GET http://localhost:8088/api/logs
```
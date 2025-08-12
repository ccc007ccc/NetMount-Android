# API Documentation

## API Overview

NetMount-Android provides RESTful API interfaces for managing network storage mount configurations and status. The Web UI communicates with the backend daemon through these APIs.

## Basic Information

- **Base URL**: `http://localhost:8088/api`
- **Data Format**: JSON
- **Authentication**: No authentication required (local access)

## API Endpoints

### 1. Configuration Management

#### Get Configuration

```http
GET /api/config
```

**Response Example:**
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

#### Update Configuration

```http
POST /api/config
Content-Type: application/json
```

**Request Body:**
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

### 2. Mount Status

#### Get All Mount Status

```http
GET /api/status
```

**Response Example:**
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

### 3. Mount Operations

#### Mount Single Storage

```http
POST /api/mount/{name}
```

**Parameters:**
- `name`: Mount configuration name (URL path parameter)

**Response:**
- `202 Accepted`: Mount request accepted
- `404 Not Found`: Specified configuration not found

#### Unmount Single Storage

```http
POST /api/unmount/{name}
```

**Parameters:**
- `name`: Mount configuration name (URL path parameter)

**Response:**
- `202 Accepted`: Unmount request accepted
- `400 Bad Request`: Unmount failed (with error message)

### 4. Log Viewing

#### Get Real-time Logs

```http
GET /api/logs
```

**Response Example:**
```json
[
  "2024-01-01 12:00:00 - --- NetMount daemon startup ---",
  "2024-01-01 12:00:01 - Using config file: /data/adb/netmount/config.json",
  "2024-01-01 12:00:02 - ✅ Successfully started mount process for 'MyWebDAV', PID: 12345"
]
```

### 5. Environment Check

#### Get Environment Check Logs

```http
GET /api/precheck
```

**Response:**
- Content-Type: `text/plain; charset=utf-8`
- Returns detailed environment check log text

## Data Models

### MountConfig

Mount configuration object:

```typescript
interface MountConfig {
  name: string;              // Mount name (unique identifier)
  type: string;              // Mount type: webdav, smb, ftp
  remote: string;            // Remote address
  mountPoint: string;        // Local mount point path
  parameters: {[key: string]: string}; // Additional parameters
  authType?: string;         // Authentication type: password, anonymous
  user?: string;             // Username
  pass?: string;             // Password
}
```

### MountStatus

Mount status object:

```typescript
interface MountStatus {
  name: string;              // Mount name
  isMounted: boolean;        // Whether mounted
  error?: string;            // Error message (if any)
  pid?: number;              // Process ID (if mounted)
}
```

## Error Handling

### HTTP Status Codes

- `200 OK`: Request successful
- `202 Accepted`: Async operation accepted
- `400 Bad Request`: Request parameter error
- `404 Not Found`: Resource not found
- `405 Method Not Allowed`: Unsupported HTTP method
- `500 Internal Server Error`: Server internal error

### Error Response Format

```json
{
  "error": "Error description message"
}
```

## Usage Examples

### JavaScript/Fetch API

```javascript
// Get mount status
async function getMountStatus() {
  const response = await fetch('/api/status');
  const status = await response.json();
  return status;
}

// Add new mount configuration
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

// Mount storage
async function mountStorage(name) {
  const response = await fetch(`/api/mount/${name}`, {
    method: 'POST'
  });
  return response.status === 202;
}
```

### cURL Examples

```bash
# Get configuration
curl -X GET http://localhost:8088/api/config

# Get status
curl -X GET http://localhost:8088/api/status

# Mount storage
curl -X POST http://localhost:8088/api/mount/MyWebDAV

# Unmount storage
curl -X POST http://localhost:8088/api/unmount/MyWebDAV

# Get logs
curl -X GET http://localhost:8088/api/logs
```
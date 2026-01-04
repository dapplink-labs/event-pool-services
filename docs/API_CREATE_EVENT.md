# 预测市场事件接口文档

## 接口概述

创建预测市场事件的复合型 API 接口，支持一键创建包含事件、子事件、结果选项和标签的完整预测事件。

## 接口信息

- **Method**: `POST`
- **Path**: `/api/v1/admin/events`
- **Content-Type**: `application/json`

## 请求参数

### 请求体 (JSON)

```json
{
  "title": "2026 World Cup Final",
  "description": "Prediction for the final match",
  "image_url": "https://example.com/image.png",
  "start_date": 1767225600,
  "end_date": 1767312000,
  "tags": ["Sports", "Soccer"],
  "sub_events": [
    {
      "question": "Who will win?",
      "outcomes": [
        {"name": "France", "color": "#0000FF", "idx": 0},
        {"name": "Brazil", "color": "#00FF00", "idx": 1}
      ]
    }
  ]
}
```

### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 事件标题 |
| description | string | 否 | 事件描述 |
| image_url | string | 否 | 事件图片URL |
| start_date | int64 | 是 | 开始时间（Unix时间戳，秒） |
| end_date | int64 | 是 | 结束时间（Unix时间戳，秒） |
| tags | []string | 否 | 标签列表 |
| sub_events | []SubEvent | 是 | 子事件列表（至少一个） |

#### SubEvent 结构

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| question | string | 是 | 问题内容 |
| outcomes | []Outcome | 是 | 结果选项列表（至少两个） |

#### Outcome 结构

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 结果名称 |
| color | string | 是 | 结果颜色（十六进制） |
| idx | int | 是 | 排序索引 |

## 响应

### 成功响应 (201 Created)

```json
{
  "guid": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "title": "2026 World Cup Final",
  "description": "Prediction for the final match",
  "image_url": "https://example.com/image.png",
  "start_date": 1767225600,
  "end_date": 1767312000,
  "tags": [
    {
      "guid": "t1a2g3u4i5d6...",
      "name": "Sports",
      "created": 1704384000,
      "updated": 1704384000
    },
    {
      "guid": "t2a3g4u5i6d7...",
      "name": "Soccer",
      "created": 1704384000,
      "updated": 1704384000
    }
  ],
  "sub_events": [
    {
      "guid": "s1u2b3e4v5e6...",
      "event_guid": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
      "question": "Who will win?",
      "outcomes": [
        {
          "guid": "o1u2t3c4o5m6...",
          "sub_event_guid": "s1u2b3e4v5e6...",
          "name": "France",
          "color": "#0000FF",
          "idx": 0,
          "created": 1704384000,
          "updated": 1704384000
        },
        {
          "guid": "o2u3t4c5o6m7...",
          "sub_event_guid": "s1u2b3e4v5e6...",
          "name": "Brazil",
          "color": "#00FF00",
          "idx": 1,
          "created": 1704384000,
          "updated": 1704384000
        }
      ],
      "created": 1704384000,
      "updated": 1704384000
    }
  ],
  "created": 1704384000,
  "updated": 1704384000
}
```

### 错误响应

#### 400 Bad Request - 请求参数错误

```json
{
  "error": "invalid_request",
  "message": "Failed to parse request body"
}
```

#### 500 Internal Server Error - 创建失败

```json
{
  "error": "create_failed",
  "message": "title cannot be empty"
}
```

## 验证规则

1. **标题不能为空**
2. **至少包含一个子事件**
3. **每个子事件至少包含两个结果选项**
4. **开始时间必须早于结束时间**

## 数据库事务

所有操作在同一个数据库事务中执行，任何步骤失败都会全部回滚，保证数据一致性。

## 使用示例

### cURL 示例

```bash
curl -X POST http://localhost:8080/api/v1/admin/events \
  -H "Content-Type: application/json" \
  -d '{
    "title": "2026 World Cup Final",
    "description": "Prediction for the final match",
    "image_url": "https://example.com/image.png",
    "start_date": 1767225600,
    "end_date": 1767312000,
    "tags": ["Sports", "Soccer"],
    "sub_events": [
      {
        "question": "Who will win?",
        "outcomes": [
          {"name": "France", "color": "#0000FF", "idx": 0},
          {"name": "Brazil", "color": "#00FF00", "idx": 1}
        ]
      }
    ]
  }'
```

### JavaScript (Fetch) 示例

```javascript
fetch('http://localhost:8080/api/v1/admin/events', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    title: '2026 World Cup Final',
    description: 'Prediction for the final match',
    image_url: 'https://example.com/image.png',
    start_date: 1767225600,
    end_date: 1767312000,
    tags: ['Sports', 'Soccer'],
    sub_events: [
      {
        question: 'Who will win?',
        outcomes: [
          { name: 'France', color: '#0000FF', idx: 0 },
          { name: 'Brazil', color: '#00FF00', idx: 1 }
        ]
      }
    ]
  })
})
.then(response => response.json())
.then(data => console.log(data))
.catch(error => console.error('Error:', error));
```

## 数据库表结构

### events (事件主表)
- guid: 事件唯一标识 (32位UUID)
- title: 事件标题
- description: 事件描述
- image_url: 事件图片URL
- start_date: 开始时间
- end_date: 结束时间
- created: 创建时间
- updated: 更新时间

### sub_events (子事件表)
- guid: 子事件唯一标识
- event_guid: 关联事件GUID
- question: 问题内容
- created: 创建时间
- updated: 更新时间

### outcomes (结果选项表)
- guid: 结果唯一标识
- sub_event_guid: 关联子事件GUID
- name: 结果名称
- color: 结果颜色
- idx: 排序索引
- created: 创建时间
- updated: 更新时间

### tags (标签表)
- guid: 标签唯一标识
- name: 标签名称（唯一）
- created: 创建时间
- updated: 更新时间

### event_tags (事件-标签关联表)
- id: 自增主键
- event_guid: 事件GUID
- tag_guid: 标签GUID
- created: 创建时间

## 部署步骤

1. **执行数据库迁移**
```bash
./event-pod-services migrate --config ./event-pod-services-config.local.yaml
```

2. **启动 API 服务**
```bash
./event-pod-services api --config ./event-pod-services-config.local.yaml
```

3. **测试接口**
```bash
curl -X POST http://localhost:8080/api/v1/admin/events \
  -H "Content-Type: application/json" \
  -d @test_event.json
```

## 注意事项

1. 所有 GUID 为 32 位紧凑型 UUID（无横杠）
2. 时间字段使用 Unix 时间戳（秒级）
3. 标签会自动去重，如果标签已存在则复用
4. 支持级联删除：删除事件会自动删除关联的子事件、结果选项和标签关联

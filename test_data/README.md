# 复杂测试案例 - 手动执行命令

## 测试案例 1: 美国总统大选（5个标签，5个子事件，复杂结构）

```bash
curl -X POST http://localhost:8080/api/v1/admin/events \
  -H "Content-Type: application/json" \
  -d @test_data/complex_event.json
```

或者直接使用 JSON 数据：

```bash
curl -X POST http://localhost:8080/api/v1/admin/events \
  -H "Content-Type: application/json" \
  -d '{
  "title": "2024 美国总统大选预测市场",
  "description": "预测2024年美国总统大选的各项结果，包括总统候选人、关键州结果、投票率等多个维度",
  "image_url": "https://example.com/images/2024-us-election.jpg",
  "start_date": 1704067200,
  "end_date": 1730851200,
  "tags": ["政治", "美国", "选举", "2024", "总统"],
  "sub_events": [
    {
      "question": "谁将赢得2024年美国总统大选？",
      "outcomes": [
        {"name": "民主党候选人", "color": "#0015BC", "idx": 0},
        {"name": "共和党候选人", "color": "#E81B23", "idx": 1},
        {"name": "独立候选人", "color": "#FFD700", "idx": 2}
      ]
    },
    {
      "question": "宾夕法尼亚州（关键摇摆州）的选举结果？",
      "outcomes": [
        {"name": "民主党获胜", "color": "#0015BC", "idx": 0},
        {"name": "共和党获胜", "color": "#E81B23", "idx": 1}
      ]
    },
    {
      "question": "佐治亚州的选举结果？",
      "outcomes": [
        {"name": "民主党获胜", "color": "#0015BC", "idx": 0},
        {"name": "共和党获胜", "color": "#E81B23", "idx": 1}
      ]
    },
    {
      "question": "2024年大选的总投票率会是多少？",
      "outcomes": [
        {"name": "低于60%", "color": "#FF6B6B", "idx": 0},
        {"name": "60%-65%", "color": "#FFA500", "idx": 1},
        {"name": "65%-70%", "color": "#FFD700", "idx": 2},
        {"name": "超过70%", "color": "#32CD32", "idx": 3}
      ]
    },
    {
      "question": "参议院控制权归属？",
      "outcomes": [
        {"name": "民主党控制", "color": "#0015BC", "idx": 0},
        {"name": "共和党控制", "color": "#E81B23", "idx": 1},
        {"name": "50-50平局", "color": "#808080", "idx": 2}
      ]
    }
  ]
}'
```

**预期结果：**
- 创建 1 个事件
- 5 个标签
- 5 个子事件
- 总共 15 个结果选项

---

## 测试案例 2: 世界杯（5个标签，6个子事件，多种结果数量）

```bash
curl -X POST http://localhost:8080/api/v1/admin/events \
  -H "Content-Type: application/json" \
  -d @test_data/sports_event.json
```

**预期结果：**
- 创建 1 个事件
- 5 个标签（部分标签可能与案例1重复，会自动复用）
- 6 个子事件
- 总共 28 个结果选项（不同子事件有不同数量的选项）

---

## 测试案例 3: 科技预测（6个标签，7个子事件）

```bash
curl -X POST http://localhost:8080/api/v1/admin/events \
  -H "Content-Type: application/json" \
  -d @test_data/tech_event.json
```

**预期结果：**
- 创建 1 个事件
- 6 个标签
- 7 个子事件
- 总共 25 个结果选项

---

## 批量测试（执行所有测试）

```bash
./test_create_events.sh
```

这将自动执行所有3个测试案例并显示结果摘要。

---

## 验证数据是否正确创建

### 查看数据库中的数据

```bash
# 连接数据库
psql -U postgres -d your_database_name

# 查看事件总数
SELECT COUNT(*) FROM events;

# 查看所有事件标题
SELECT guid, title, created FROM events ORDER BY created DESC;

# 查看标签总数
SELECT COUNT(*) FROM tags;

# 查看所有标签
SELECT * FROM tags;

# 查看子事件总数
SELECT COUNT(*) FROM sub_events;

# 查看某个事件的所有子事件
SELECT se.guid, se.question, e.title
FROM sub_events se
JOIN events e ON se.event_guid = e.guid
ORDER BY e.created DESC;

# 查看结果选项总数
SELECT COUNT(*) FROM outcomes;

# 查看某个子事件的所有结果选项
SELECT o.name, o.color, o.idx, se.question
FROM outcomes o
JOIN sub_events se ON o.sub_event_guid = se.guid
ORDER BY o.idx;

# 查看事件-标签关联
SELECT e.title, t.name
FROM event_tags et
JOIN events e ON et.event_guid = e.guid
JOIN tags t ON et.tag_guid = t.guid
ORDER BY e.created DESC;

# 退出
\q
```

---

## 测试数据统计

### 案例 1: 美国总统大选
- **子事件**: 5个
- **结果选项**: 15个 (3+2+2+4+4)
- **标签**: 5个

### 案例 2: 世界杯
- **子事件**: 6个
- **结果选项**: 28个 (7+5+4+4+2+6)
- **标签**: 5个

### 案例 3: 科技预测
- **子事件**: 7个
- **结果选项**: 25个 (3+4+3+2+4+3+3)
- **标签**: 6个

### 总计（如果全部执行）
- **事件**: 3个
- **子事件**: 18个
- **结果选项**: 68个
- **标签**: ~12-16个（有些标签会复用）
- **事件-标签关联**: 16个

---

## 注意事项

1. 确保 API 服务已启动：`./event-pod-services api --config ./event-pod-services-config.local.yaml`
2. 确保数据库迁移已执行：`./event-pod-services migrate --config ./event-pod-services-config.local.yaml`
3. 所有时间戳都使用 Unix 时间戳（秒）
4. 标签会自动去重，相同名称的标签只会创建一次
5. 每个测试案例都有不同的复杂度，用于验证系统的健壮性

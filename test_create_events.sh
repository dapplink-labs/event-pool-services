#!/bin/bash

# 测试脚本 - 批量测试创建事件接口
# 使用方法: ./test_create_events.sh

API_URL="http://localhost:8080/api/v1/admin/events"
TEST_DATA_DIR="./test_data"

echo "================================================"
echo "开始测试创建预测事件接口"
echo "API URL: $API_URL"
echo "================================================"
echo ""

# 测试案例1: 美国总统大选
echo "📋 测试案例 1: 2024 美国总统大选预测市场"
echo "----------------------------------------"
response1=$(curl -s -X POST $API_URL \
  -H "Content-Type: application/json" \
  -d @$TEST_DATA_DIR/complex_event.json \
  -w "\nHTTP_CODE:%{http_code}")

http_code1=$(echo "$response1" | grep "HTTP_CODE" | cut -d':' -f2)
body1=$(echo "$response1" | sed '/HTTP_CODE/d')

if [ "$http_code1" = "201" ]; then
    echo "✅ 成功 (HTTP $http_code1)"
    event_guid1=$(echo "$body1" | grep -o '"guid":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "事件 GUID: $event_guid1"
    echo "子事件数量: $(echo "$body1" | grep -o '"sub_events":\[' | wc -l)"
    echo "标签数量: $(echo "$body1" | grep -o '"tags":\[' | wc -l)"
else
    echo "❌ 失败 (HTTP $http_code1)"
    echo "错误信息: $body1"
fi
echo ""

# 等待1秒
sleep 1

# 测试案例2: 世界杯
echo "⚽ 测试案例 2: 2026 FIFA 世界杯完整预测"
echo "----------------------------------------"
response2=$(curl -s -X POST $API_URL \
  -H "Content-Type: application/json" \
  -d @$TEST_DATA_DIR/sports_event.json \
  -w "\nHTTP_CODE:%{http_code}")

http_code2=$(echo "$response2" | grep "HTTP_CODE" | cut -d':' -f2)
body2=$(echo "$response2" | sed '/HTTP_CODE/d')

if [ "$http_code2" = "201" ]; then
    echo "✅ 成功 (HTTP $http_code2)"
    event_guid2=$(echo "$body2" | grep -o '"guid":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "事件 GUID: $event_guid2"
    echo "子事件数量: $(echo "$body2" | grep -o '"question":' | wc -l)"
else
    echo "❌ 失败 (HTTP $http_code2)"
    echo "错误信息: $body2"
fi
echo ""

# 等待1秒
sleep 1

# 测试案例3: 科技行业
echo "💻 测试案例 3: 2025年科技行业重大事件预测"
echo "----------------------------------------"
response3=$(curl -s -X POST $API_URL \
  -H "Content-Type: application/json" \
  -d @$TEST_DATA_DIR/tech_event.json \
  -w "\nHTTP_CODE:%{http_code}")

http_code3=$(echo "$response3" | grep "HTTP_CODE" | cut -d':' -f2)
body3=$(echo "$response3" | sed '/HTTP_CODE/d')

if [ "$http_code3" = "201" ]; then
    echo "✅ 成功 (HTTP $http_code3)"
    event_guid3=$(echo "$body3" | grep -o '"guid":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "事件 GUID: $event_guid3"
    echo "子事件数量: $(echo "$body3" | grep -o '"question":' | wc -l)"
else
    echo "❌ 失败 (HTTP $http_code3)"
    echo "错误信息: $body3"
fi
echo ""

# 汇总结果
echo "================================================"
echo "测试总结"
echo "================================================"
success_count=0
if [ "$http_code1" = "201" ]; then ((success_count++)); fi
if [ "$http_code2" = "201" ]; then ((success_count++)); fi
if [ "$http_code3" = "201" ]; then ((success_count++)); fi

echo "总测试数: 3"
echo "成功: $success_count"
echo "失败: $((3 - success_count))"
echo ""

if [ $success_count -eq 3 ]; then
    echo "🎉 所有测试通过！"
else
    echo "⚠️  有测试失败，请检查日志"
fi

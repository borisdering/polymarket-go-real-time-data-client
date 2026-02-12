# CLOB WebSocket 迁移方案

## 背景

Polymarket 宣布从 2026年1月26日（周一）开始，RTDS 系统将停止发送 CLOB WebSocket 消息。需要迁移到独立的 CLOB WebSocket 服务器。

## 端点信息

- **市场数据端点**: `wss://ws-subscriptions-clob.polymarket.com/ws/market`（无需认证）
- **用户数据端点**: `wss://ws-subscriptions-clob.polymarket.com/ws/user`（需要认证）

## 受影响的功能

### 需要迁移的主题
- `TopicClobUser` - CLOB 用户数据（订单、交易）
- `TopicClobMarket` - CLOB 市场数据（订单簿、价格变化、最后交易价格等）

### 不受影响的功能
以下主题继续使用 RTDS 系统：
- `TopicActivity` - 市场活动
- `TopicComments` - 评论
- `TopicRfq` - RFQ
- `TopicCryptoPrices` - 加密货币价格
- `TopicEquityPrices` - 股票价格

## 实现方案

### 方案1：双连接模式（推荐）

**架构设计**：
- 保持 RTDS 连接用于非 CLOB 主题
- 创建独立的 CLOB 连接用于 CLOB 主题
- Client 自动路由订阅到正确的连接

**优点**：
- 向后兼容，现有代码无需修改
- 自动处理连接管理
- 清晰的职责分离

**实现步骤**：
1. 创建 `clob_protocol.go` - CLOB WebSocket 协议实现
2. 创建 `clob_client.go` - CLOB 专用客户端（内部使用）
3. 修改 `Client` 支持双连接模式
4. 实现自动路由逻辑

### 方案2：单一连接模式（备选）

将所有订阅迁移到 CLOB 端点（如果支持所有主题）。

**缺点**：
- 需要确认 CLOB 端点是否支持所有主题
- 可能破坏向后兼容性

## 技术细节

### CLOB 协议差异

**RTDS 协议**（当前）：
```json
{
  "action": "subscribe",
  "subscriptions": [
    {
      "topic": "clob_market",
      "type": "agg_orderbook",
      "filters": "[\"token_id\"]"
    }
  ]
}
```

**CLOB WebSocket 协议**（新）：
```json
{
  "type": "subscribe",
  "channel": "market",
  "markets": ["condition_id"]
}
```

### 认证方式

用户数据端点需要在连接时进行认证。根据文档，可能需要：
- 在连接 URL 中添加认证参数，或
- 在连接后发送认证消息

需要进一步调研确认。

## 迁移计划

1. ✅ 调研 CLOB WebSocket 协议细节
2. ⏳ 创建 CLOB 协议实现
3. ⏳ 实现双连接模式
4. ⏳ 更新订阅路由逻辑
5. ⏳ 更新示例代码
6. ⏳ 添加测试
7. ⏳ 更新文档

## 向后兼容性

- ✅ 现有 API 保持不变
- ✅ 非 CLOB 订阅继续使用 RTDS
- ✅ CLOB 订阅自动迁移到新端点
- ✅ 用户无需修改代码

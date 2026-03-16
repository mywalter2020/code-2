# Alibaba Live Integration Checklist

在接入真实 Alibaba 平台前，至少逐项确认以下内容：

## 1. 协议文档确认

- [ ] 真实 base URL
- [ ] publish / update / on_shelf / off_shelf 的真实 path
- [ ] 对应 method 名是否与当前占位值一致
- [ ] 请求格式是 `json` 还是 `application/x-www-form-urlencoded`
- [ ] 鉴权参数放在 body / query / header 哪个位置
- [ ] `app_key` / `v` 是否必须显式携带
- [ ] timestamp 格式要求
- [ ] sign 串拼接规则是否与当前 MD5 骨架一致

## 2. 请求字段确认

- [ ] title 是否有长度限制
- [ ] subtitle 是否支持
- [ ] description/content 分别对应哪个平台字段
- [ ] images 是否要求主图/辅图拆分
- [ ] price 单位是元还是分
- [ ] category / brand / sku 的字段名与格式
- [ ] page 结构是否允许直接透传
- [ ] metadata/out_biz_id 是否有固定命名要求

## 3. 响应字段确认

- [ ] 成功标志是 `success` / `status` / `code` 哪个为准
- [ ] 远端商品 ID 字段名
- [ ] 错误码字段名
- [ ] 错误描述字段名
- [ ] 是否存在嵌套 `response` / `data` / `result`

## 4. 联调验证

- [ ] 使用测试账号和沙箱环境
- [ ] 首次只验证 publish
- [ ] 记录完整请求 body / query / headers（脱敏后）
- [ ] 记录完整响应 body
- [ ] 验证 remote_id 能被正确提取
- [ ] 验证失败场景能正确映射为 `failed`
- [ ] 验证重复请求是否需要幂等控制

## 5. 上线前

- [ ] 将默认占位 method/path 替换为真实值
- [ ] 固化 auth placement / request format
- [ ] 配置超时与重试参数
- [ ] 确认限流策略
- [ ] 确认错误重试策略
- [ ] 验证调试日志已脱敏
- [ ] 确认凭据保管方式

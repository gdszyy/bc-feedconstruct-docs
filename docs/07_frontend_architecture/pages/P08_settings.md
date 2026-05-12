# P08 — 设置

## 目的

承载语言、赔率格式、通知偏好、单点登录等用户偏好。

## 数据来源

| 设置项 | 来源 |
|---|---|
| 语言 | M12（驱动文案重渲染） |
| 赔率格式（decimal/fractional/american） | 本地偏好 + M05 渲染层换算（参考 `docs/05_odds_math/`） |
| 通知（结算/取消/回滚） | M14 + 浏览器通知 API |
| 时区 | 本地偏好 + 全局格式化 |

## 关键组件

- `<LocaleSelect/>`（M12）
- `<OddsFormatSelect/>`
- `<NotificationToggle/>`
- `<TimezoneSelect/>`

## 验收要点

- 切换语言不重连 WS、不重拉描述结构
- 切换赔率格式即时反映到所有可见盘口
- 通知偏好持久化（localStorage + 服务端可选）

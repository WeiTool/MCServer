// 属性中文标签映射
export const propertyLabels: Record<string, string> = {
  'accepts-transfers': '接受转移',
  'allow-flight': '允许飞行',
  'broadcast-console-to-ops': '控制台广播给OP',
  'broadcast-rcon-to-ops': 'RCON广播给OP',
  'bug-report-link': 'BUG报告链接',
  'chat-spam-threshold-seconds': '聊天防刷屏阈值(秒)',
  'command-spam-threshold-seconds': '命令防刷屏阈值(秒)',
  'difficulty': '难度',
  'enable-code-of-conduct': '启用行为准则',
  'enable-jmx-monitoring': '启用JMX监控',
  'enable-query': '启用Query',
  'enable-rcon': '启用RCON',
  'enable-status': '启用Status',
  'enforce-secure-profile': '强制安全档案',
  'enforce-whitelist': '强制白名单',
  'entity-broadcast-range-percentage': '实体广播范围百分比',
  'force-gamemode': '强制游戏模式',
  'function-permission-level': '函数权限等级',
  'gamemode': '游戏模式',
  'generate-structures': '生成结构',
  'generator-settings': '生成器设置',
  'hardcore': '极限模式',
  'hide-online-players': '隐藏在线玩家',
  'initial-disabled-packs': '初始禁用数据包',
  'initial-enabled-packs': '初始启用数据包',
  'level-name': '世界名称',
  'level-seed': '世界种子',
  'level-type': '世界类型',
  'log-ips': '记录IP',
  'management-server-allowed-origins': '管理服务器允许来源',
  'management-server-enabled': '启用管理服务器',
  'management-server-host': '管理服务器主机',
  'management-server-port': '管理服务器端口',
  'management-server-secret': '管理服务器密钥',
  'management-server-tls-enabled': '管理服务器TLS启用',
  'management-server-tls-keystore': '管理服务器TLS密钥库',
  'management-server-tls-keystore-password': '管理服务器TLS密钥库密码',
  'max-chained-neighbor-updates': '最大链式邻居更新',
  'max-players': '最大玩家数',
  'max-tick-time': '最大Tick时间',
  'max-world-size': '最大世界大小',
  'motd': '服务器标语',
  'network-compression-threshold': '网络压缩阈值',
  'online-mode': '正版验证',
  'op-permission-level': 'OP权限等级',
  'pause-when-empty-seconds': '空服暂停(秒)',
  'player-idle-timeout': '玩家闲置超时(秒)',
  'prevent-proxy-connections': '阻止代理连接',
  'query.port': 'Query端口',
  'rate-limit': '速率限制',
  'rcon.password': 'RCON密码',
  'rcon.port': 'RCON端口',
  'region-file-compression': '区域文件压缩',
  'require-resource-pack': '强制资源包',
  'resource-pack': '资源包URL',
  'resource-pack-id': '资源包ID',
  'resource-pack-prompt': '资源包提示',
  'resource-pack-sha1': '资源包SHA1',
  'server-ip': '服务器IP',
  'server-port': '服务器端口',
  'simulation-distance': '模拟距离',
  'spawn-protection': '出生点保护半径',
  'status-heartbeat-interval': '状态心跳间隔',
  'sync-chunk-writes': '同步区块写入',
  'text-filtering-config': '文本过滤配置',
  'text-filtering-version': '文本过滤版本',
  'use-native-transport': '使用原生传输',
  'view-distance': '视野距离',
  'white-list': '白名单'
}

// 游戏模式映射
export const gamemodeMap: Record<string, string> = {
  'survival': '生存模式',
  'creative': '创造模式',
  'adventure': '冒险模式',
  'spectator': '旁观模式'
}

// 难度映射
export const difficultyMap: Record<string, string> = {
  'peaceful': '和平',
  'easy': '简单',
  'normal': '普通',
  'hard': '困难'
}

// 世界类型映射
export const levelTypeMap: Record<string, string> = {
  'minecraft:normal': '普通',
  'minecraft:flat': '超平坦',
  'minecraft:large_biomes': '大型生物群系',
  'minecraft:amplified': '放大化',
  'minecraft:single_biome_surface': '单一生物群系'
}

// 获取属性的中文标签
export function getPropertyLabel(key: string): string {
  return propertyLabels[key] || key
}

// 翻译属性值
export function translatePropertyValue(key: string, value: string): string {
  if (key === 'gamemode') return gamemodeMap[value] || value
  if (key === 'difficulty') return difficultyMap[value] || value
  if (key === 'level-type') return levelTypeMap[value] || value
  return value || '未设置'
}

// 格式化整个配置对象（核心函数）
export function formatProperties(properties: Record<string, string> | null): Array<{ key: string; label: string; value: string }> {
  if (!properties) return []
  return Object.entries(properties).map(([key, value]) => ({
    key,
    label: getPropertyLabel(key),
    value: translatePropertyValue(key, value)
  }))
}

// 常用配置列表
export const commonPropertyKeys = [
  'server-port',
  'max-players',
  'gamemode',
  'difficulty',
  'online-mode',
  'spawn-protection',
  'enable-query',
  'enable-rcon'
]

export function getCommonProperties(properties: Record<string, string> | null): Array<{ key: string; label: string; value: string }> {
  if (!properties) return []
  return commonPropertyKeys.map(key => ({
    key,
    label: getPropertyLabel(key),
    value: translatePropertyValue(key, properties[key] || '')
  }))
}
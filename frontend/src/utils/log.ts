// 工具层：日志级别判断
// 根据日志文本识别级别，用于控制台着色

// 日志级别
export type LogLevel = 'info' | 'warning' | 'error' | 'command'

// 单条日志（文本 + 级别）
export interface LogLine {
  text: string
  level: LogLevel
}

// 根据日志文本判断级别
// INFO -> info(白)，WARNING -> warning(黄)，ERROR -> error(红)，命令回显 -> command(蓝)
export function detectLogLevel(text: string): LogLevel {
  if (text.startsWith('> ')) {
    return 'command'
  }
  const upper = text.toUpperCase()
  if (upper.includes('ERROR')) {
    return 'error'
  }
  if (upper.includes('WARNING')) {
    return 'warning'
  }
  return 'info'
}

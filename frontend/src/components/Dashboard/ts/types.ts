// 首页看板共享类型定义
// 供 useServerInfo / useRealtimeStats / useDashboard 等模块复用

/** 服务器实例（首页卡片展示的数据结构） */
export type ServerInstance = {
  /** 服务器显示名称 */
  name: string;
  /** 服务器文件夹绝对路径 */
  path: string;
  /** 是否包含有效的 jar 文件 */
  hasJar: boolean;
  /** jar 文件数量 */
  jarCount: number;
  /** jar 文件名列表 */
  jarFiles: string[];
};

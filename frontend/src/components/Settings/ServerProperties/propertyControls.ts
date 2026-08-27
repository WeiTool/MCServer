// ViewModel：基础配置（server.properties）控件渲染辅助
// Settings 中"基础配置"导航项的纯函数辅助：根据属性键/值决定渲染哪种控件
import { gamemodeMap, difficultyMap, levelTypeMap } from "../../../utils/serverProperties";

/** 属性值对应的控件类型 */
export type PropertyControlType = 'text' | 'select' | 'switch';

/** 需要宽输入框的长文本属性键 */
const longTextKeys = ['motd', 'resource-pack', 'generator-settings', 'level-seed'];

/**
 * 判断属性应渲染的控件类型
 * - gamemode / difficulty / level-type → 下拉选择
 * - 布尔字符串（true/false）→ 开关
 * - 其余 → 文本框
 */
export function getValueType(key: string, value: string): PropertyControlType {
  if (key === 'gamemode' || key === 'difficulty' || key === 'level-type') {
    return 'select';
  }
  if (value === 'true' || value === 'false') {
    return 'switch';
  }
  return 'text';
}

/** 获取下拉选择控件的选项列表 */
export function getSelectOptions(key: string): Array<{ label: string; value: string }> {
  if (key === 'gamemode') {
    return Object.entries(gamemodeMap).map(([value, label]) => ({ label, value }));
  }
  if (key === 'difficulty') {
    return Object.entries(difficultyMap).map(([value, label]) => ({ label, value }));
  }
  if (key === 'level-type') {
    return Object.entries(levelTypeMap).map(([value, label]) => ({ label, value }));
  }
  return [];
}

/** 将字符串值转为开关控件的布尔状态 */
export function getSwitchValue(value: string): boolean {
  return value === 'true';
}

/** 判断是否为长文本属性（需要更宽的输入框） */
export function isLongText(key: string): boolean {
  return longTextKeys.includes(key);
}

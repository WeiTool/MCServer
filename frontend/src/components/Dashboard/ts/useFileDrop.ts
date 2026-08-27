// ViewModel：首页拖拽添加服务器（自包含模块）
// 职责：Wails 文件拖放监听注册/注销、文件类型识别、调用后端复制/解压、状态提示
// 供 useDashboard 调用；组合式函数内直接使用 onMounted/onUnmounted 注册生命周期
import { ref, onMounted, onUnmounted } from "vue";
import { CopyJarFile, ExtractServerZip } from "../../../../wailsjs/go/main/App";
import { OnFileDrop, OnFileDropOff } from "../../../../wailsjs/runtime/runtime";
import { useMessage } from "naive-ui";

/**
 * useFileDrop
 * 首页"拖拽添加服务器"功能（Windows/Linux 通用）：
 * - 拖入 .jar → 复制到服务器目录（CopyJarFile）
 * - 拖入任意压缩包（zip / tar / tar.gz / tar.bz2 等）→ 解压为新服务器（ExtractServerZip）
 * 组件挂载时注册 Wails 拖放监听（只响应 --wails-drop-target 的 drop 目标元素），卸载时移除。
 */
export function useFileDrop() {
  const message = useMessage();
  // 是否正在处理文件（后端同步复制/解压，期间 UI 需要提示，避免误以为卡死）
  const isProcessing = ref(false);

  /** 调用后端复制 JAR 文件到服务器目录 */
  async function handleCopyJar(jarPath: string, serverName: string) {
    isProcessing.value = true;
    try {
      const response = await CopyJarFile(jarPath, serverName);
      if (response.success) {
        message.success(response.message, {
          closable: true,
          duration: 5000
        });
      } else {
        message.error(response.message);
      }
    } catch (error) {
      message.error("调用失败: " + error);
    } finally {
      isProcessing.value = false;
    }
  }

  /** 调用后端解压服务器压缩包 */
  async function handleExtractArchive(zipPath: string, serverName: string) {
    isProcessing.value = true;
    try {
      const response = await ExtractServerZip(zipPath, serverName);
      if (response.success) {
        message.success(response.message, {
          closable: true,
          duration: 5000
        });
      } else {
        message.error(response.message);
      }
    } catch (error) {
      message.error("调用失败: " + error);
    } finally {
      isProcessing.value = false;
    }
  }

  /**
   * Wails 拖放回调：识别文件类型并分发
   * - 文件名以 .jar 结尾 → 复制为服务器
   * - 其他任意类型 → 尝试按压缩包解压（格式由后端魔数识别）
   * 服务器名 = 去掉扩展名后的文件名
   */
  function handleFileDrop(_x: number, _y: number, paths: string[]) {
    if (!paths || paths.length === 0) {
      message.warning("未检测到文件");
      return;
    }

    const filePath = paths[0];
    // 从路径中提取文件名（兼容 Windows 和 Unix）
    const fileName = filePath.split(/[\/\\]/).pop() || "";
    if (!fileName) {
      message.warning("未检测到文件");
      return;
    }

    // 处理中禁止再次拖入（Wails 回调异步执行，避免并发重复操作）
    if (isProcessing.value) {
      message.warning("正在处理文件，请稍候");
      return;
    }

    const lower = fileName.toLowerCase();

    // .jar → 复制；其余任意类型 → 按压缩包解压
    if (lower.endsWith(".jar")) {
      const serverName = fileName.replace(/\.jar$/i, "");
      handleCopyJar(filePath, serverName);
    } else {
      // 去掉扩展名作为服务器名（兼容 .tar.gz / .tar.bz2 等多段扩展名）
      const base = fileName.replace(/\.[^.]+$/, "").replace(/\.(tar|tbz2|tgz)$/i, "");
      const serverName = base || "server";
      handleExtractArchive(filePath, serverName);
    }
  }

  // 组件挂载时注册拖拽监听，true 表示只响应设置了 --wails-drop-target 的 drop 目标元素
  onMounted(() => {
    OnFileDrop(handleFileDrop, true);
  });

  // 组件卸载时移除监听
  onUnmounted(() => {
    OnFileDropOff();
  });

  return {
    isProcessing,
    handleCopyJar,
    handleExtractArchive,
  };
}

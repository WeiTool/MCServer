import { ref, shallowRef } from "vue";
import { CheckVersion, DownloadUpdate } from "../../wailsjs/go/main/App";

/** 版本实例 */
export type VersionInstance = {
  // 是否有新版本（最新 release 版本号与当前版本不一致）
  hasUpdate: boolean;
  // 当前运行版本号（如 0.0.1）
  current: string;
  // 最新 release 版本号（如 0.0.1）
  latest: string;
  // 匹配当前平台的下载文件完整名称（如 MCServer-v0.0.1-beta1-windows-amd64.exe）
  assetName: string;
  // 下载地址
  downloadUrl: string;
  // 是否为 beta 预发布版本
  isBeta: boolean;
};

export function useVersionCheck() {
  const versionInfo = shallowRef<VersionInstance | null>(null);
  const versionError = ref<string | null>(null);
  const isDownloading = ref(false);
  const downloadComplete = ref(false);
  const downloadedVersion = ref<string>("");
  const hasUpdate = ref(false);

  async function checkVersion() {
    try {
      versionError.value = null;
      isDownloading.value = false;
      downloadComplete.value = false;
      hasUpdate.value = false;

      // 检查版本
      const response = await CheckVersion();

      versionInfo.value = {
        hasUpdate: response.hasUpdate,
        current: response.current,
        latest: response.latest,
        assetName: response.assetName,
        downloadUrl: response.downloadUrl,
        isBeta: response.isBeta,
      };

      hasUpdate.value = versionInfo.value.hasUpdate;

      return {
        versionInfo: versionInfo.value,
        hasUpdate: hasUpdate.value,
        downloadComplete: false,
        versionError: null,
        downloadedVersion: "",
      };
    } catch (err) {
      // 版本检查失败静默处理（如无网络），不打扰用户
      return {
        versionInfo: null,
        hasUpdate: false,
        downloadComplete: false,
        versionError: null,
        downloadedVersion: "",
      };
    }
  }

  /** 下载新版本（不退出、不替换，关闭应用后自动完成） */
  async function downloadUpdate() {
    if (!versionInfo.value?.hasUpdate || !versionInfo.value.downloadUrl) return;
    isDownloading.value = true;
    downloadComplete.value = false;
    try {
      await DownloadUpdate(versionInfo.value.downloadUrl);
      downloadComplete.value = true;
      downloadedVersion.value = versionInfo.value.latest;
    } catch (err) {
      versionError.value = err instanceof Error ? err.message : "下载失败";
      throw err; // 向外抛出，让调用方可捕获
    } finally {
      isDownloading.value = false;
    }
  }

  return {
    versionInfo,
    versionError,
    isDownloading,
    downloadComplete,
    downloadedVersion,
    hasUpdate,
    checkVersion,
    downloadUpdate,
  };
}

package api

import (
	"MCServer/backend/model"
	"MCServer/backend/service"
	"context"
)

type VersionApi struct {
	ctx context.Context
}

func NewVersion() *VersionApi { return &VersionApi{} }

// Startup 启动初始化
func (api *VersionApi) Startup(ctx context.Context) {
	api.ctx = ctx
}

func (api *VersionApi) CheckVersion() (model.VersionResponse, error) {
	return service.CheckVersion()
}

// DownloadUpdate 下载新版本（静默更新第一步）
// 只下载到临时文件并写入 config/update.json(pending)，
// 不替换不退出；替换动作在用户关闭应用时由 OnBeforeClose → ApplyPendingUpdate 触发
func (api *VersionApi) DownloadUpdate(downloadURL string) error {
	return service.DownloadUpdate(downloadURL)
}

// GetUpdateState 读取并清除更新状态（静默更新第三步）
// 应用更新重启后调用，返回上次更新结果（updated=已更新到某版本 / error=替换失败）
func (api *VersionApi) GetUpdateState() (model.UpdateState, error) {
	return service.GetUpdateState()
}

// ApplyPendingUpdate 应用待执行的更新（静默更新第二步，应用关闭时触发）
// 由 app.go OnBeforeClose 调用，检测到 pending 状态则启动替换脚本
func (api *VersionApi) ApplyPendingUpdate() error {
	return service.ApplyPendingUpdate()
}

// GetGlobalConfig 读取全局配置（config/global_config.json）
func (api *VersionApi) GetGlobalConfig() (model.GlobalConfig, error) {
	return service.GetGlobalConfig()
}

// SaveGlobalConfig 保存全局配置（config/global_config.json）
func (api *VersionApi) SaveGlobalConfig(cfg model.GlobalConfig) error {
	return service.SaveGlobalConfig(cfg)
}

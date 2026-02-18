package settings

type Repository interface {
	ServiceSettingDataManager
	ServiceSettingConfigurationDataManager
}

package settings

type (
	UserDataCollection struct {
		AccountSettings []ServiceSettingConfiguration `json:"accountSettings,omitempty"`
		UserSettings    []ServiceSettingConfiguration `json:"userSettings,omitempty"`
	}
)

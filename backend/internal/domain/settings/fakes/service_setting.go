package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/settings"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/settings/converters"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/filtering"
	"github.com/verygoodsoftwarenotvirus/platform/v4/pointer"
)

// BuildFakeServiceSetting builds a faked service setting.
func BuildFakeServiceSetting() *types.ServiceSetting {
	defaultValue := buildUniqueString()

	return &types.ServiceSetting{
		ID:           BuildFakeID(),
		Name:         buildUniqueString(),
		Type:         "user",
		Description:  buildUniqueString(),
		Enumeration:  []string{defaultValue},
		DefaultValue: pointer.To(defaultValue),
		AdminsOnly:   true,
		CreatedAt:    BuildFakeTime(),
	}
}

// BuildFakeServiceSettingsList builds a faked ServiceSettingList.
func BuildFakeServiceSettingsList() *filtering.QueryFilteredResult[types.ServiceSetting] {
	var examples []*types.ServiceSetting
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, BuildFakeServiceSetting())
	}

	return &filtering.QueryFilteredResult[types.ServiceSetting]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeServiceSettingCreationRequestInput builds a faked ServiceSettingCreationRequestInput.
func BuildFakeServiceSettingCreationRequestInput() *types.ServiceSettingCreationRequestInput {
	serviceSetting := BuildFakeServiceSetting()
	return converters.ConvertServiceSettingToServiceSettingCreationRequestInput(serviceSetting)
}

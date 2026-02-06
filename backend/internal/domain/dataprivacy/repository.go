package dataprivacy

type (
	Repository interface {
		DataPrivacyDataManager
		UserDataDisclosureDataManager
	}
)

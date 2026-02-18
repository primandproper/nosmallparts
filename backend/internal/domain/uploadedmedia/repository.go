package uploadedmedia

type (
	// Repository is a type that manages uploaded media in permanent storage.
	Repository interface {
		UploadedMediaDataManager
	}
)

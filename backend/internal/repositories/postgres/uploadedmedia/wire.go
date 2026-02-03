package uploadedmedia

import "github.com/google/wire"

var (
	// UploadedMediaRepoProviders represents what we provide to dependency injectors.
	UploadedMediaRepoProviders = wire.NewSet(ProvideUploadedMediaRepository)
)

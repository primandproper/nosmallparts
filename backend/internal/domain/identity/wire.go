package identity

import (
	"github.com/google/wire"
)

var (
	Providers = wire.NewSet(
		ProvideAccountDataManagerFromRepository,
		ProvideAccountInvitationDataManagerFromRepository,
		ProvideUserDataManagerFromRepository,
		ProvideAccountUserMembershipDataManagerFromRepository,
	)
)

func ProvideAccountDataManagerFromRepository(r Repository) AccountDataManager {
	return r
}

func ProvideAccountInvitationDataManagerFromRepository(r Repository) AccountInvitationDataManager {
	return r
}

func ProvideUserDataManagerFromRepository(r Repository) UserDataManager {
	return r
}

func ProvideAccountUserMembershipDataManagerFromRepository(r Repository) AccountUserMembershipDataManager {
	return r
}

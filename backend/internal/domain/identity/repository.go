package identity

type Repository interface {
	AccountDataManager
	AccountInvitationDataManager
	UserDataManager
	AccountUserMembershipDataManager
}

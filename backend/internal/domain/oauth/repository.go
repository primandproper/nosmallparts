package oauth

type Repository interface {
	OAuth2ClientDataManager
	OAuth2ClientTokenDataManager
}

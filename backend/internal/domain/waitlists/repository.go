package waitlists

// Repository describes persistence capabilities for waitlists and signups.
type Repository interface {
	WaitlistDataManager
	WaitlistSignupDataManager
}

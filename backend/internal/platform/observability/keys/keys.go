package keys

const (
	idSuffix = ".id"

	// RequesterIDKey is the standard key for referring to a requesting user's ID.
	RequesterIDKey = "request.made_by"
	// AccountIDKey is the standard key for referring to an account ID.
	AccountIDKey = "account" + idSuffix
	// AccountInvitationKey is the standard key for referring to an account ID.
	AccountInvitationKey = "account_invitation"
	// AccountInvitationIDKey is the standard key for referring to an account ID.
	AccountInvitationIDKey = AccountInvitationKey + idSuffix
	// AccountInvitationTokenKey is the standard key for referring to an account invitation token.
	AccountInvitationTokenKey = "account_invitation.token"
	// ActiveAccountIDKey is the standard key for referring to an active account ID.
	ActiveAccountIDKey = "active_account" + idSuffix
	// UserIDKey is the standard key for referring to a user ID.
	UserIDKey = "user" + idSuffix
	// UserNotificationIDKey is the standard key for referring to a user notification ID.
	UserNotificationIDKey = "user_notification" + idSuffix
	// UserEmailAddressKey is the standard key for referring to a user's email address.
	UserEmailAddressKey = "user.email_address"
	// UserIsServiceAdminKey is the standard key for referring to a user's admin status.
	UserIsServiceAdminKey = "user.is_admin"
	// UsernameKey is the standard key for referring to a username.
	UsernameKey = "user.username"
	// #nosec G101 UserEmailVerificationTokenKey is the standard key for referring to a username.
	UserEmailVerificationTokenKey = "user.email_verification_token"
	// NameKey is the standard key for referring to a name.
	NameKey = "name"
	// SpanIDKey is the standard key for referring to a span ID.
	SpanIDKey = "span" + idSuffix
	// TraceIDKey is the standard key for referring to a trace ID.
	TraceIDKey = "trace" + idSuffix
	// FilterCreatedAfterKey is the standard key for referring to a types.QueryFilter's CreatedAfter field.
	FilterCreatedAfterKey = "query_filter.created_after"
	// FilterCreatedBeforeKey is the standard key for referring to a types.QueryFilter's CreatedBefore field.
	FilterCreatedBeforeKey = "query_filter.created_before"
	// FilterUpdatedAfterKey is the standard key for referring to a types.QueryFilter's UpdatedAfter field.
	FilterUpdatedAfterKey = "query_filter.updated_after"
	// FilterUpdatedBeforeKey is the standard key for referring to a types.QueryFilter's UpdatedAfter field.
	FilterUpdatedBeforeKey = "query_filter.updated_before"
	// FilterSortByKey is the standard key for referring to a types.QueryFilter's SortBy field.
	FilterSortByKey = "query_filter.sort_by"
	// FilterCursorKey is the standard key for referring to a types.QueryFilter's next cursor.
	FilterCursorKey = "query_filter.cursor"
	// FilterLimitKey is the standard key for referring to a types.QueryFilter's limit.
	FilterLimitKey = "query_filter.limit"
	// FilterIsNilKey is the standard key for referring to a types.QueryFilter's null status.
	FilterIsNilKey = "query_filter.is_nil"
	// WebhookIDKey is the standard key for referring to a webhook's ID.
	WebhookIDKey = "webhook" + idSuffix
	// WebhookTriggerEventIDKey is the standard key for referring to a webhook trigger event's ID.
	WebhookTriggerEventIDKey = "webhook_trigger_event" + idSuffix
	// WaitlistIDKey is the standard key for referring to a waitlist ID.
	WaitlistIDKey = "waitlist" + idSuffix
	// WaitlistSignupIDKey is the standard key for referring to a waitlist signup ID.
	WaitlistSignupIDKey = "waitlist_signup" + idSuffix
	// IssueReportIDKey is the standard key for referring to an issue report ID.
	IssueReportIDKey = "issue_report" + idSuffix
	// UploadedMediaIDKey is the standard key for referring to an uploaded media ID.
	UploadedMediaIDKey = "uploaded_media" + idSuffix
	// AuditLogEntryIDKey is the standard key for referring to an audit log entry's ID.
	AuditLogEntryIDKey = "audit_log_entry" + idSuffix
	// AuditLogEntryResourceTypesKey is the standard key for referring to an audit log entry's resource type.
	AuditLogEntryResourceTypesKey = "audit_log_entry.resource_types"
	// URLKey is the standard key for referring to a URL.
	URLKey = "url"
	// PasswordResetTokenKey is the standard key for referring to a password reset token's ID.
	PasswordResetTokenKey = "password_reset_token"
	// PasswordResetTokenIDKey is the standard key for referring to a password reset token's ID.
	PasswordResetTokenIDKey = PasswordResetTokenKey + idSuffix
	// RequestHeadersKey is the standard key for referring to a http.Request's Headers.
	RequestHeadersKey = "request.headers"
	// RequestIDKey is the standard key for referring to a http.Request's ID.
	RequestIDKey = "request" + idSuffix
	// RequestMethodKey is the standard key for referring to a http.Request's Method.
	RequestMethodKey = "request.method"
	// RequestURIKey is the standard key for referring to a http.Request's URI.
	RequestURIKey = "request.uri"
	// ResponseStatusKey is the standard key for referring to a http.Request's status.
	ResponseStatusKey = "response.status"
	// ResponseHeadersKey is the standard key for referring to a http.Response's Headers.
	ResponseHeadersKey = "response.headers"
	// ReasonKey is the standard key for referring to a reason for a change.
	ReasonKey = "reason"
	// URLQueryKey is the standard key for referring to a URL query.
	URLQueryKey = "url.query"
	// SearchQueryKey is the standard key for referring to a search query parameter value.
	SearchQueryKey = "search_query"
	// UserAgentOSKey is the standard key for referring to a user agent's OS.
	UserAgentOSKey = "os"
	// UserAgentBotKey is the standard key for referring to a user agent's bot status.
	UserAgentBotKey = "is_bot"
	// UserAgentMobileKey is the standard key for referring to user agent's mobile status.
	UserAgentMobileKey = "is_mobile"
	// ValidationErrorKey is the standard key for referring to a struct validation error.
	ValidationErrorKey = "validation_error"
	// UserDataAggregationReportIDKey is the standard key for referring to a user data aggregation report.
	UserDataAggregationReportIDKey = "user_data_aggregation_report" + idSuffix
	// IndexNameKey is the standard key for referring to a given search index.
	IndexNameKey = "index.name"
	// UseDatabaseKey is the standard key for referring to whether or not the database was used in search.
	UseDatabaseKey = "use_database"

	// ValidInstrumentKey is the standard key for referring to a valid instrument.
	ValidInstrumentKey = "valid_instrument"
	// ValidInstrumentIDKey is the standard key for referring to a valid instrument's ID.
	ValidInstrumentIDKey = ValidInstrumentKey + idSuffix

	// ValidVesselIDKey is the standard key for referring to a valid vessel's ID.
	ValidVesselIDKey = "valid_vessel" + idSuffix

	// ValidIngredientKey is the standard key for referring to a valid ingredient.
	ValidIngredientKey = "valid_ingredient"
	// ValidIngredientIDKey is the standard key for referring to a valid ingredient's ID.
	ValidIngredientIDKey = ValidIngredientKey + idSuffix

	// ValidIngredientGroupIDKey is the standard key for referring to a valid ingredient group's ID.
	ValidIngredientGroupIDKey = "valid_ingredient_group" + idSuffix

	// ValidPreparationKey is the standard key for referring to a valid preparation.
	ValidPreparationKey = "valid_preparation"
	// ValidPreparationIDKey is the standard key for referring to a valid preparation's ID.
	ValidPreparationIDKey = ValidPreparationKey + idSuffix

	// ValidIngredientStateKey is the standard key for referring to a valid ingredient state.
	ValidIngredientStateKey = "valid_ingredient_state"
	// ValidIngredientStateIDKey is the standard key for referring to a valid ingredient state's ID.
	ValidIngredientStateIDKey = ValidIngredientStateKey + idSuffix

	// ValidIngredientStateIngredientKey is the standard key for referring to a valid ingredient state ingredient.
	ValidIngredientStateIngredientKey = "valid_ingredient_state_ingredient"
	// ValidIngredientStateIngredientIDKey is the standard key for referring to a valid ingredient state ingredient's ID.
	ValidIngredientStateIngredientIDKey = ValidIngredientStateIngredientKey + idSuffix

	// ValidIngredientPreparationKey is the standard key for referring to a valid preparation ingredient.
	ValidIngredientPreparationKey = "valid_ingredient_preparation"
	// ValidIngredientPreparationIDKey is the standard key for referring to a valid preparation ingredient's ID.
	ValidIngredientPreparationIDKey = ValidIngredientPreparationKey + idSuffix

	// ValidPrepTaskConfigKey is the standard key for referring to a valid prep task config.
	ValidPrepTaskConfigKey = "valid_prep_task_config"
	// ValidPrepTaskConfigIDKey is the standard key for referring to a valid prep task config's ID.
	ValidPrepTaskConfigIDKey = ValidPrepTaskConfigKey + idSuffix

	// ValidPreparationInstrumentKey is the standard key for referring to a valid preparation instrument.
	ValidPreparationInstrumentKey = "valid_preparation_instrument"
	// ValidPreparationInstrumentIDKey is the standard key for referring to a valid preparation instrument's ID.
	ValidPreparationInstrumentIDKey = ValidPreparationInstrumentKey + idSuffix

	// ValidIngredientMeasurementUnitKey is the standard key for referring to a valid ingredient measurement unit.
	ValidIngredientMeasurementUnitKey = "valid_ingredient_measurement_unit"
	// ValidIngredientMeasurementUnitIDKey is the standard key for referring to a valid ingredient measurement unit's ID.
	ValidIngredientMeasurementUnitIDKey = ValidIngredientMeasurementUnitKey + idSuffix

	// MealKey is the standard key for referring to a meal.
	MealKey = "meal"
	// MealIDKey is the standard key for referring to a meal's ID.
	MealIDKey = MealKey + idSuffix
	// MealListIDKey is the standard key for referring to a meal list's ID.
	MealListIDKey = "meal_list" + idSuffix
	// MealListItemIDKey is the standard key for referring to a meal list item's ID.
	MealListItemIDKey = "meal_list_item" + idSuffix

	// RecipeKey is the standard key for referring to a recipe.
	RecipeKey = "recipe"
	// RecipeIDKey is the standard key for referring to a recipe's ID.
	RecipeIDKey = RecipeKey + idSuffix
	// RecipeListIDKey is the standard key for referring to a recipe list's ID.
	RecipeListIDKey = "recipe_list" + idSuffix
	// RecipeListItemIDKey is the standard key for referring to a recipe list item's MealPlanTaskID.
	RecipeListItemIDKey = "recipe_list_item" + idSuffix

	// RecipeStepIDKey is the standard key for referring to a recipe step's MealPlanTaskID.
	RecipeStepIDKey = "recipe_step" + idSuffix

	// RecipePrepTaskIDKey is the standard key for referring to a recipe prep task's MealPlanTaskID.
	RecipePrepTaskIDKey = "recipe_prep_task" + idSuffix

	// RecipeStepInstrumentIDKey is the standard key for referring to a recipe step instrument's MealPlanTaskID.
	RecipeStepInstrumentIDKey = "recipe_step_instrument" + idSuffix

	// RecipeStepVesselIDKey is the standard key for referring to a recipe step vessel's MealPlanTaskID.
	RecipeStepVesselIDKey = "recipe_step_vessel" + idSuffix

	// RecipeStepIngredientIDKey is the standard key for referring to a recipe step ingredient's MealPlanTaskID.
	RecipeStepIngredientIDKey = "recipe_step_ingredient" + idSuffix

	// RecipeStepCompletionConditionIDKey is the standard key for referring to a recipe step completion condition's MealPlanTaskID.
	RecipeStepCompletionConditionIDKey = "recipe_step_completion_condition" + idSuffix

	// RecipeStepProductIDKey is the standard key for referring to a recipe step product's MealPlanTaskID.
	RecipeStepProductIDKey = "recipe_step_product" + idSuffix

	// MealPlanIDKey is the standard key for referring to a meal plan's MealPlanTaskID.
	MealPlanIDKey = "meal_plan" + idSuffix

	// MealPlanKey is the standard key for referring to a meal plan.
	MealPlanKey = "meal_plan"

	// MealPlanEventIDKey is the standard key for referring to a meal plan event's MealPlanTaskID.
	MealPlanEventIDKey = "meal_plan_event" + idSuffix

	// MealPlanOptionIDKey is the standard key for referring to a meal plan option's MealPlanTaskID.
	MealPlanOptionIDKey = "meal_plan_option" + idSuffix

	// MealPlanOptionVoteIDKey is the standard key for referring to a meal plan option vote's MealPlanTaskID.
	MealPlanOptionVoteIDKey = "meal_plan_option_vote" + idSuffix

	// ValidMeasurementUnitKey is the standard key for referring to a valid measurement unit's MealPlanTaskID.
	ValidMeasurementUnitKey = "valid_measurement_unit"
	// ValidMeasurementUnitIDKey is the standard key for referring to a valid measurement unit's MealPlanTaskID.
	ValidMeasurementUnitIDKey = ValidMeasurementUnitKey + idSuffix

	// MealPlanTaskIDKey is the standard key for referring to a meal plan task's MealPlanTaskID.
	MealPlanTaskIDKey = "meal_plan_task" + idSuffix

	// MealPlanGroceryListItemIDKey is the standard key for referring to a meal plan grocery list item's MealPlanTaskID.
	MealPlanGroceryListItemIDKey = "meal_plan_grocery_list_item" + idSuffix

	// ValidMeasurementUnitConversionIDKey is the standard key for referring to a valid measurement conversion's MealPlanTaskID.
	ValidMeasurementUnitConversionIDKey = "valid_measurement_conversion" + idSuffix

	// RecipeMediaIDKey is the standard key for referring to a recipe media's MealPlanTaskID.
	RecipeMediaIDKey = "recipe_media" + idSuffix

	// ServiceSettingIDKey is the standard key for referring to a service setting's MealPlanTaskID.
	ServiceSettingIDKey = "service_setting" + idSuffix

	// ServiceSettingNameKey is the standard key for referring to a service setting's MealPlanTaskID.
	ServiceSettingNameKey = "service_setting.name"

	// ServiceSettingConfigurationIDKey is the standard key for referring to a service setting configuration's MealPlanTaskID.
	ServiceSettingConfigurationIDKey = "service_setting_configuration" + idSuffix

	// ServiceSettingConfigurationNameKey is the standard key for referring to a service setting configuration's Name.
	ServiceSettingConfigurationNameKey = "service_setting_configuration.name"

	// UserIngredientPreferenceIDKey is the standard key for referring to a user ingredient preference's MealPlanTaskID.
	UserIngredientPreferenceIDKey = "user_ingredient_preference" + idSuffix

	// AccountInstrumentOwnershipIDKey is the standard key for referring to an account instrument ownership's MealPlanTaskID.
	AccountInstrumentOwnershipIDKey = "account_instrument_ownership" + idSuffix

	// RecipeRatingIDKey is the standard key for referring to a recipe rating's MealPlanTaskID.
	RecipeRatingIDKey = "recipe_rating" + idSuffix

	// OAuth2ClientIDKey is the standard key for referring to an OAuth2 client's database MealPlanTaskID.
	OAuth2ClientIDKey = "oauth2_clients" + idSuffix

	// OAuth2ClientClientIDKey is the standard key for referring to an OAuth2 client's client MealPlanTaskID.
	OAuth2ClientClientIDKey = "oauth2_clients.client_id"

	// OAuth2ClientTokenIDKey is the standard key for referring to an OAuth2 client token's MealPlanTaskID.
	/* #nosec G101 */
	OAuth2ClientTokenIDKey = "oauth2_client_tokens" + idSuffix

	// OAuth2ClientTokenCodeKey is the standard key for referring to an OAuth2 client token's code.
	/* #nosec G101 */
	OAuth2ClientTokenCodeKey = "oauth2_client_tokens.code"

	// OAuth2ClientTokenAccessKey is the standard key for referring to an OAuth2 client token's access.
	/* #nosec G101 */
	OAuth2ClientTokenAccessKey = "oauth2_client_tokens.access"

	// OAuth2ClientTokenRefreshKey is the standard key for referring to an OAuth2 client token's refresh.
	/* #nosec G101 */
	OAuth2ClientTokenRefreshKey = "oauth2_client_tokens.refresh"

	// ValidPreparationVesselIDKey is the standard key for referring to a valid preparation vessel's MealPlanTaskID.
	ValidPreparationVesselIDKey = "valid_preparation_vessels" + idSuffix
)

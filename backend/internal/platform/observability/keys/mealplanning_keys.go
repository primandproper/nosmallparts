package keys

const (
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
	// RecipeListItemIDKey is the standard key for referring to a recipe list item's ID.
	RecipeListItemIDKey = "recipe_list_item" + idSuffix

	// RecipeStepIDKey is the standard key for referring to a recipe step's ID.
	RecipeStepIDKey = "recipe_step" + idSuffix

	// RecipePrepTaskIDKey is the standard key for referring to a recipe prep task's ID.
	RecipePrepTaskIDKey = "recipe_prep_task" + idSuffix

	// RecipeStepInstrumentIDKey is the standard key for referring to a recipe step instrument's ID.
	RecipeStepInstrumentIDKey = "recipe_step_instrument" + idSuffix

	// RecipeStepVesselIDKey is the standard key for referring to a recipe step vessel's ID.
	RecipeStepVesselIDKey = "recipe_step_vessel" + idSuffix

	// RecipeStepIngredientIDKey is the standard key for referring to a recipe step ingredient's ID.
	RecipeStepIngredientIDKey = "recipe_step_ingredient" + idSuffix

	// RecipeStepCompletionConditionIDKey is the standard key for referring to a recipe step completion condition's ID.
	RecipeStepCompletionConditionIDKey = "recipe_step_completion_condition" + idSuffix

	// RecipeStepProductIDKey is the standard key for referring to a recipe step product's ID.
	RecipeStepProductIDKey = "recipe_step_product" + idSuffix

	// MealPlanIDKey is the standard key for referring to a meal plan's ID.
	MealPlanIDKey = "meal_plan" + idSuffix

	// MealPlanKey is the standard key for referring to a meal plan.
	MealPlanKey = "meal_plan"

	// MealPlanEventIDKey is the standard key for referring to a meal plan event's ID.
	MealPlanEventIDKey = "meal_plan_event" + idSuffix

	// MealPlanOptionIDKey is the standard key for referring to a meal plan option's ID.
	MealPlanOptionIDKey = "meal_plan_option" + idSuffix

	// MealPlanOptionVoteIDKey is the standard key for referring to a meal plan option vote's ID.
	MealPlanOptionVoteIDKey = "meal_plan_option_vote" + idSuffix

	// ValidMeasurementUnitKey is the standard key for referring to a valid measurement unit.
	ValidMeasurementUnitKey = "valid_measurement_unit"
	// ValidMeasurementUnitIDKey is the standard key for referring to a valid measurement unit's ID.
	ValidMeasurementUnitIDKey = ValidMeasurementUnitKey + idSuffix

	// MealPlanTaskIDKey is the standard key for referring to a meal plan task's ID.
	MealPlanTaskIDKey = "meal_plan_task" + idSuffix

	// MealPlanGroceryListItemIDKey is the standard key for referring to a meal plan grocery list item's ID.
	MealPlanGroceryListItemIDKey = "meal_plan_grocery_list_item" + idSuffix

	// ValidMeasurementUnitConversionIDKey is the standard key for referring to a valid measurement conversion's ID.
	ValidMeasurementUnitConversionIDKey = "valid_measurement_conversion" + idSuffix

	// RecipeMediaIDKey is the standard key for referring to a recipe media's ID.
	RecipeMediaIDKey = "recipe_media" + idSuffix

	// UserIngredientPreferenceIDKey is the standard key for referring to a user ingredient preference's ID.
	UserIngredientPreferenceIDKey = "user_ingredient_preference" + idSuffix

	// AccountInstrumentOwnershipIDKey is the standard key for referring to an account instrument ownership's ID.
	AccountInstrumentOwnershipIDKey = "account_instrument_ownership" + idSuffix

	// RecipeRatingIDKey is the standard key for referring to a recipe rating's ID.
	RecipeRatingIDKey = "recipe_rating" + idSuffix

	// ValidPreparationVesselIDKey is the standard key for referring to a valid preparation vessel's ID.
	ValidPreparationVesselIDKey = "valid_preparation_vessels" + idSuffix
)

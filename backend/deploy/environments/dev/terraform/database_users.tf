locals {
  api_database_username                       = "api_db_user"
  db_user_manager_username                    = "db_user_manager"
  async_message_handler_database_username     = "async_message_handler"
  db_cleaner_username                         = "db_cleaner"
  meal_plan_finalizer_username                = "meal_plan_finalizer"
  meal_plan_grocery_list_initializer_username = "meal_plan_grocery_list_initializer"
  meal_plan_task_creator_username             = "meal_plan_task_creator"
  search_data_index_scheduler_username        = "search_data_index_scheduler"
}

# api_database_username

resource "random_password" "db_user_manager_username_database_password" {
  length           = 64
  special          = true
  override_special = "#$*-_=+[]"
}

resource "google_sql_user" "db_user_manager_username" {
  name     = local.db_user_manager_username
  instance = google_sql_database_instance.dev.name
  password = random_password.db_user_manager_username_database_password.result
}

# api_database_username

resource "random_password" "api_user_database_password" {
  length           = 64
  special          = true
  override_special = "#$*-_=+[]"
}

resource "google_sql_user" "api_user" {
  name     = local.api_database_username
  instance = google_sql_database_instance.dev.name
  password = random_password.api_user_database_password.result
}

# async_message_handler_database_username

resource "random_password" "async_message_handler_database_user_database_password" {
  length           = 64
  special          = true
  override_special = "#$*-_=+[]"
}

resource "google_sql_user" "async_message_handler_database_user" {
  name     = local.async_message_handler_database_username
  instance = google_sql_database_instance.dev.name
  password = random_password.async_message_handler_database_user_database_password.result
}

# db_cleaner_username

resource "random_password" "db_cleaner_user_database_password" {
  length           = 64
  special          = true
  override_special = "#$*-_=+[]"
}

resource "google_sql_user" "db_cleaner_user" {
  name     = local.db_cleaner_username
  instance = google_sql_database_instance.dev.name
  password = random_password.db_cleaner_user_database_password.result
}

# meal_plan_finalizer_username

resource "random_password" "meal_plan_finalizer_user_database_password" {
  length           = 64
  special          = true
  override_special = "#$*-_=+[]"
}

resource "google_sql_user" "meal_plan_finalizer_user" {
  name     = local.meal_plan_finalizer_username
  instance = google_sql_database_instance.dev.name
  password = random_password.meal_plan_finalizer_user_database_password.result
}

# meal_plan_grocery_list_initializer_username

resource "random_password" "meal_plan_grocery_list_initializer_user_database_password" {
  length           = 64
  special          = true
  override_special = "#$*-_=+[]"
}

resource "google_sql_user" "meal_plan_grocery_list_initializer_user" {
  name     = local.meal_plan_grocery_list_initializer_username
  instance = google_sql_database_instance.dev.name
  password = random_password.meal_plan_grocery_list_initializer_user_database_password.result
}

# meal_plan_task_creator_username

resource "random_password" "meal_plan_task_creator_user_database_password" {
  length           = 64
  special          = true
  override_special = "#$*-_=+[]"
}

resource "google_sql_user" "meal_plan_task_creator_user" {
  name     = local.meal_plan_task_creator_username
  instance = google_sql_database_instance.dev.name
  password = random_password.meal_plan_task_creator_user_database_password.result
}

# search_data_index_scheduler_username

resource "random_password" "search_data_index_scheduler_user_database_password" {
  length           = 64
  special          = true
  override_special = "#$*-_=+[]"
}

resource "google_sql_user" "search_data_index_scheduler_user" {
  name     = local.search_data_index_scheduler_username
  instance = google_sql_database_instance.dev.name
  password = random_password.search_data_index_scheduler_user_database_password.result
}

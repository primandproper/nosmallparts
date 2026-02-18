resource "random_password" "db_user_manager_password" {
  length           = 64
  special          = true
  override_special = "#$*-_=+[]"
}

resource "google_sql_user" "db_user_manager" {
  name     = local.db_user_manager_username
  instance = google_sql_database_instance.main.name
  password = random_password.db_user_manager_password.result
}

resource "random_password" "api_user_database_password" {
  length           = 64
  special          = true
  override_special = "#$*-_=+[]"
}

resource "google_sql_user" "api_user" {
  name     = local.api_database_username
  instance = google_sql_database_instance.main.name
  password = random_password.api_user_database_password.result
}

resource "random_password" "async_message_handler_database_password" {
  length           = 64
  special          = true
  override_special = "#$*-_=+[]"
}

resource "google_sql_user" "async_message_handler" {
  name     = local.async_message_handler_database_username
  instance = google_sql_database_instance.main.name
  password = random_password.async_message_handler_database_password.result
}

resource "random_password" "db_cleaner_password" {
  length           = 64
  special          = true
  override_special = "#$*-_=+[]"
}

resource "google_sql_user" "db_cleaner" {
  name     = local.db_cleaner_username
  instance = google_sql_database_instance.main.name
  password = random_password.db_cleaner_password.result
}

resource "random_password" "search_data_index_scheduler_password" {
  length           = 64
  special          = true
  override_special = "#$*-_=+[]"
}

resource "google_sql_user" "search_data_index_scheduler" {
  name     = local.search_data_index_scheduler_username
  instance = google_sql_database_instance.main.name
  password = random_password.search_data_index_scheduler_password.result
}

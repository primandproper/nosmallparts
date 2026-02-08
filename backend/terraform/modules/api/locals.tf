locals {
  managed_by_label = "managed_by"
  zone             = coalesce(var.gcp_main_zone, "${var.gcp_region}-a")
  db_ssl_cn        = coalesce(var.db_ssl_common_name, "${var.project_id}-${var.environment}")

  api_database_username                   = "api_db_user"
  db_user_manager_username                = "db_user_manager"
  async_message_handler_database_username = "async_message_handler"
  db_cleaner_username                     = "db_cleaner"
  search_data_index_scheduler_username    = "search_data_index_scheduler"
}

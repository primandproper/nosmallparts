ALTER TABLE `stages` ADD `host_user_id` text REFERENCES user(id);--> statement-breakpoint
ALTER TABLE `stages` ADD `host_guest_id` text;
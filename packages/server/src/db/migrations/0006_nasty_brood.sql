CREATE TABLE `stage_ip_bans` (
	`id` text PRIMARY KEY NOT NULL,
	`stage_id` text NOT NULL,
	`ip` text NOT NULL,
	`banned_at` integer NOT NULL,
	FOREIGN KEY (`stage_id`) REFERENCES `stages`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
ALTER TABLE `play_scripts` ADD `enabled` integer DEFAULT true NOT NULL;
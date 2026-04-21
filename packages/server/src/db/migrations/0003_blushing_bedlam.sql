CREATE TABLE `stage_bans` (
	`id` text PRIMARY KEY NOT NULL,
	`stage_id` text NOT NULL,
	`user_id` text,
	`guest_id` text,
	`banned_at` integer NOT NULL,
	FOREIGN KEY (`stage_id`) REFERENCES `stages`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE set null
);
--> statement-breakpoint
ALTER TABLE `stage_participants` ADD `force_muted_by_host` integer DEFAULT false NOT NULL;--> statement-breakpoint
ALTER TABLE `stages` ADD `paused_by_host` integer DEFAULT false NOT NULL;
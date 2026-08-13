CREATE TABLE `account` (
	`id` text PRIMARY KEY NOT NULL,
	`userId` text NOT NULL,
	`accountId` text NOT NULL,
	`providerId` text NOT NULL,
	`accessToken` text,
	`refreshToken` text,
	`idToken` text,
	`accessTokenExpiresAt` integer,
	`refreshTokenExpiresAt` integer,
	`scope` text,
	`password` text,
	`createdAt` integer NOT NULL,
	`updatedAt` integer NOT NULL,
	FOREIGN KEY (`userId`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `play_script_cast_members` (
	`id` text PRIMARY KEY NOT NULL,
	`play_script_id` text NOT NULL,
	`name` text NOT NULL,
	`description` text,
	`sort_order` integer DEFAULT 0 NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`play_script_id`) REFERENCES `play_scripts`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `play_script_lines` (
	`id` text PRIMARY KEY NOT NULL,
	`play_script_id` text NOT NULL,
	`cast_member_id` text NOT NULL,
	`line_text` text NOT NULL,
	`line_index` integer NOT NULL,
	`direction` text,
	`scroll_mode` text DEFAULT 'auto',
	`auto_scroll_wpm` integer,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`play_script_id`) REFERENCES `play_scripts`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`cast_member_id`) REFERENCES `play_script_cast_members`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `play_script_ratings` (
	`id` text PRIMARY KEY NOT NULL,
	`play_script_id` text NOT NULL,
	`user_id` text NOT NULL,
	`rating` integer NOT NULL,
	`created_at` integer NOT NULL,
	`updated_at` integer NOT NULL,
	FOREIGN KEY (`play_script_id`) REFERENCES `play_scripts`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE cascade,
	CONSTRAINT "rating_range" CHECK("play_script_ratings"."rating" BETWEEN 1 AND 5)
);
--> statement-breakpoint
CREATE UNIQUE INDEX `play_script_ratings_unique` ON `play_script_ratings` (`play_script_id`,`user_id`);--> statement-breakpoint
CREATE TABLE `play_script_tags` (
	`play_script_id` text NOT NULL,
	`script_tag_id` text NOT NULL,
	PRIMARY KEY(`play_script_id`, `script_tag_id`),
	FOREIGN KEY (`play_script_id`) REFERENCES `play_scripts`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`script_tag_id`) REFERENCES `script_tags`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `play_scripts` (
	`id` text PRIMARY KEY NOT NULL,
	`title` text NOT NULL,
	`description` text DEFAULT '' NOT NULL,
	`author` text,
	`is_public` integer DEFAULT false NOT NULL,
	`created_by_user` text NOT NULL,
	`created_at` integer NOT NULL,
	`updated_at` integer NOT NULL,
	FOREIGN KEY (`created_by_user`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `script_tags` (
	`id` text PRIMARY KEY NOT NULL,
	`name` text NOT NULL,
	`description` text,
	`created_at` integer NOT NULL
);
--> statement-breakpoint
CREATE UNIQUE INDEX `script_tags_name_unique` ON `script_tags` (`name`);--> statement-breakpoint
CREATE TABLE `session` (
	`id` text PRIMARY KEY NOT NULL,
	`userId` text NOT NULL,
	`token` text NOT NULL,
	`expiresAt` integer NOT NULL,
	`ipAddress` text,
	`userAgent` text,
	`createdAt` integer NOT NULL,
	`updatedAt` integer NOT NULL,
	FOREIGN KEY (`userId`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE UNIQUE INDEX `session_token_unique` ON `session` (`token`);--> statement-breakpoint
CREATE TABLE `stage_participant_roles` (
	`stage_participant_id` text NOT NULL,
	`cast_member_id` text NOT NULL,
	PRIMARY KEY(`stage_participant_id`, `cast_member_id`),
	FOREIGN KEY (`stage_participant_id`) REFERENCES `stage_participants`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `stage_participants` (
	`id` text PRIMARY KEY NOT NULL,
	`stage_id` text NOT NULL,
	`user_id` text,
	`display_name` text NOT NULL,
	`is_director` integer DEFAULT false NOT NULL,
	`is_ready` integer DEFAULT false NOT NULL,
	`is_audience` integer DEFAULT false NOT NULL,
	`created_at` integer NOT NULL,
	`left_at` integer,
	FOREIGN KEY (`stage_id`) REFERENCES `stages`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE set null
);
--> statement-breakpoint
CREATE TABLE `stages` (
	`id` text PRIMARY KEY NOT NULL,
	`code` text NOT NULL,
	`play_script_id` text,
	`current_line_index` integer DEFAULT 0 NOT NULL,
	`phase` text DEFAULT 'lobby' NOT NULL,
	`created_at` integer NOT NULL,
	`updated_at` integer NOT NULL,
	FOREIGN KEY (`play_script_id`) REFERENCES `play_scripts`(`id`) ON UPDATE no action ON DELETE set null
);
--> statement-breakpoint
CREATE UNIQUE INDEX `stages_code_unique` ON `stages` (`code`);--> statement-breakpoint
CREATE TABLE `user` (
	`id` text PRIMARY KEY NOT NULL,
	`name` text NOT NULL,
	`email` text NOT NULL,
	`emailVerified` integer DEFAULT false NOT NULL,
	`image` text,
	`createdAt` integer NOT NULL,
	`updatedAt` integer NOT NULL
);
--> statement-breakpoint
CREATE UNIQUE INDEX `user_email_unique` ON `user` (`email`);--> statement-breakpoint
CREATE TABLE `verification` (
	`id` text PRIMARY KEY NOT NULL,
	`identifier` text NOT NULL,
	`value` text NOT NULL,
	`expiresAt` integer NOT NULL,
	`createdAt` integer,
	`updatedAt` integer
);

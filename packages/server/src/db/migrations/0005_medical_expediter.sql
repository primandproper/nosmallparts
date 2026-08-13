CREATE TABLE `play_script_line_speakers` (
	`line_id` text NOT NULL,
	`cast_member_id` text NOT NULL,
	PRIMARY KEY(`line_id`, `cast_member_id`),
	FOREIGN KEY (`line_id`) REFERENCES `play_script_lines`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`cast_member_id`) REFERENCES `play_script_cast_members`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TEMP TABLE `_line_speaker_backfill` (
	`line_id` text NOT NULL,
	`cast_member_id` text NOT NULL
);
--> statement-breakpoint
INSERT INTO `_line_speaker_backfill` (`line_id`, `cast_member_id`) SELECT `id`, `cast_member_id` FROM `play_script_lines`;--> statement-breakpoint
PRAGMA foreign_keys=OFF;--> statement-breakpoint
CREATE TABLE `__new_play_script_lines` (
	`id` text PRIMARY KEY NOT NULL,
	`play_script_id` text NOT NULL,
	`line_text` text NOT NULL,
	`line_index` integer NOT NULL,
	`direction` text,
	`scroll_mode` text DEFAULT 'auto',
	`auto_scroll_wpm` integer,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`play_script_id`) REFERENCES `play_scripts`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
INSERT INTO `__new_play_script_lines`("id", "play_script_id", "line_text", "line_index", "direction", "scroll_mode", "auto_scroll_wpm", "created_at") SELECT "id", "play_script_id", "line_text", "line_index", "direction", "scroll_mode", "auto_scroll_wpm", "created_at" FROM `play_script_lines`;--> statement-breakpoint
DROP TABLE `play_script_lines`;--> statement-breakpoint
ALTER TABLE `__new_play_script_lines` RENAME TO `play_script_lines`;--> statement-breakpoint
INSERT INTO `play_script_line_speakers` (`line_id`, `cast_member_id`) SELECT `line_id`, `cast_member_id` FROM `_line_speaker_backfill`;--> statement-breakpoint
DROP TABLE `_line_speaker_backfill`;--> statement-breakpoint
PRAGMA foreign_keys=ON;
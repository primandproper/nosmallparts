ALTER TABLE `stages` ADD `director_mode` integer DEFAULT false NOT NULL;--> statement-breakpoint
ALTER TABLE `stage_participants` DROP COLUMN `is_director`;
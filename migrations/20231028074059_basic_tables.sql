-- Create "users" table
CREATE TABLE `users` (`rowid` integer NULL, `created_at` text NOT NULL DEFAULT CURRENT_TIMESTAMP, `email` text NOT NULL, `name` text NOT NULL DEFAULT '', `timezone` text NOT NULL DEFAULT 'Australia/Melbourne', PRIMARY KEY (`rowid`));
-- Create index "users_email" to table: "users"
CREATE UNIQUE INDEX `users_email` ON `users` (`email`);
-- Create "media_category" table
CREATE TABLE `media_category` (`category` text NOT NULL, PRIMARY KEY (`category`)) WITHOUT ROWID;
-- Create "user_settings" table
CREATE TABLE `user_settings` (`user_id` integer NOT NULL, `max_afk` integer NOT NULL DEFAULT 60, `max_blur` integer NOT NULL DEFAULT 2, `inactivity_blur` integer NOT NULL DEFAULT 2, `menu_blur` integer NOT NULL DEFAULT 8, `last_update` text NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (`user_id`), CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`rowid`) ON UPDATE NO ACTION ON DELETE NO ACTION, CHECK (max_afk >= 30 AND max_afk <= 1200), CHECK (max_blur >= 0 AND max_blur <=10), CHECK (inactivity_blur >= 0 AND inactivity_blur <=10), CHECK (menu_blur >= 0 AND menu_blur <=10)) WITHOUT ROWID;
-- Create index "user_settings_index" to table: "user_settings"
CREATE INDEX `user_settings_index` ON `user_settings` (`user_id`, `last_update`);
-- Create "user_category_settings" table
CREATE TABLE `user_category_settings` (`user_id` integer NOT NULL, `category` text NOT NULL, `max_afk` integer NOT NULL DEFAULT 60, `inactivity_blur` integer NOT NULL DEFAULT 2, `menu_blur` integer NOT NULL DEFAULT 8, `last_update` text NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (`user_id`, `category`), CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`rowid`) ON UPDATE NO ACTION ON DELETE NO ACTION, CHECK (max_afk >= 30 AND max_afk <= 1200), CHECK (inactivity_blur >= 0 AND inactivity_blur <=10), CHECK (menu_blur >= 0 AND menu_blur <=10)) WITHOUT ROWID;
-- Create index "user_category_settings_index" to table: "user_category_settings"
CREATE INDEX `user_category_settings_index` ON `user_category_settings` (`user_id`, `last_update`);
-- Create "media" table
CREATE TABLE `media` (`identifier` text NOT NULL, `category` text NOT NULL, `series` text NOT NULL DEFAULT '', `user_id` integer NOT NULL, `display_name` text NOT NULL DEFAULT '', `last_update` text NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (`identifier`, `category`, `series`, `user_id`), CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`rowid`) ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT `1` FOREIGN KEY (`category`) REFERENCES `media_category` (`category`) ON UPDATE NO ACTION ON DELETE NO ACTION) WITHOUT ROWID;
-- Create "media_stats" table
CREATE TABLE `media_stats` (`media_identifier` text NOT NULL, `category` text NOT NULL, `series` text NOT NULL DEFAULT '', `user_id` integer NOT NULL, `immerse_date` text NOT NULL, `last_read` text NOT NULL DEFAULT CURRENT_TIMESTAMP, `read_time` integer NOT NULL DEFAULT 0, `read_chars` integer NOT NULL DEFAULT 0, `read_lines` integer NULL DEFAULT 0, `read_pages` integer NULL DEFAULT 0, `paused` integer NULL DEFAULT 0, `last_update` text NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (`media_identifier`, `category`, `user_id`, `immerse_date`), CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`rowid`) ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT `1` FOREIGN KEY (`category`) REFERENCES `media_category` (`category`) ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT `2` FOREIGN KEY (`media_identifier`, `category`, `series`, `user_id`) REFERENCES `media` (`identifier`, `category`, `series`, `user_id`) ON UPDATE NO ACTION ON DELETE NO ACTION, CHECK (read_time >= 0), CHECK (read_chars >= 0), CHECK (read_lines >= 0), CHECK (read_pages >= 0), CHECK (paused IN (0, 1))) WITHOUT ROWID;
-- Create index "media_stats_user_index" to table: "media_stats"
CREATE INDEX `media_stats_user_index` ON `media_stats` (`user_id`, `last_update`);
-- Create index "media_stats_user_category_index" to table: "media_stats"
CREATE INDEX `media_stats_user_category_index` ON `media_stats` (`user_id`, `category`, `last_update`);

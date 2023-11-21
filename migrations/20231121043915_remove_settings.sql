-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Drop "user_settings" table
DROP TABLE `user_settings`;
-- Drop "user_category_settings" table
DROP TABLE `user_category_settings`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

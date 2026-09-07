UPDATE users
SET middle_name = ''
WHERE middle_name IS NULL;

ALTER TABLE users
ALTER COLUMN middle_name SET DEFAULT '';

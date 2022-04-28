CREATE TABLE IF NOT EXISTS users (
    id		                TEXT PRIMARY KEY NOT NULL,
	name	                TEXT,
	deleted	                BOOLEAN NOT NULL,
	real_name               TEXT,
    tz                      TEXT,
    profile_status_text     TEXT,
    profile_status_emoji    TEXT,
    profile_image_512       TEXT
);



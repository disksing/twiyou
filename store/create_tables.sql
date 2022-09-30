CREATE TABLE users (
	id varchar(32) NOT NULL,
	name varchar(128) NOT NULL,
	user_name varchar(255) NOT NULL,
	profile_image varchar(255),

	follower_count INT NOT NULL,
	following_count INT NOT NULL,
	tweet_count INT NOT NULL DEFAULT 0,
	listed_count INT NOT NULL,

	is_following BOOLEAN NOT NULL DEFAULT FALSE,
	is_follower BOOLEAN NOT NULL DEFAULT FALSE,

	is_following1 BOOLEAN NOT NULL DEFAULT FALSE,
	is_follower1 BOOLEAN NOT NULL DEFAULT FALSE,
	tweet_count1 INT NOT NULL DEFAULT 0,

	last_active TIMESTAMP,
	updated_at TIMESTAMP,
	PRIMARY KEY (id),
	INDEX (last_active),
    INDEX (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE stats (
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP PRIMARY KEY,
    id varchar(32) NOT NULL,
    name varchar(128) NOT NULL,
    user_name varchar(255) NOT NULL,
    profile_image varchar(255),
    follower_count INT NOT NULL,
    following_count INT NOT NULL,
    tweet_count INT NOT NULL,
    listed_count INT NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE iterations (
    id INT AUTO_INCREMENT PRIMARY KEY,
	state varchar(255) NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	complete_fetch_followers_at TIMESTAMP,
	complete_fetch_following_at TIMESTAMP,
	complete_pull_users_at TIMESTAMP,
	complete_sum_events_at TIMESTAMP,
	complete_stash_users_at TIMESTAMP,
	completed_at TIMESTAMP,
	next_token varchar(255) NOT NULL DEFAULT ""
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO iterations (state) VALUES ('initial');

CREATE TABLE events (
	user_id varchar(32) NOT NULL,
	created_at TIMESTAMP,
	event varchar(32) NOT NULL,
	INDEX (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE logs (
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	level varchar(32) NOT NULL,
	message varchar(255) NOT NULL,
	INDEX (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE system (
    `key` varchar(255) NOT NULL PRIMARY KEY,
    value varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO system (`key`, value) VALUES ('dbver', '1');
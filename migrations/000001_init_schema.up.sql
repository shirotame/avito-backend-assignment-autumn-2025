CREATE TABLE users (
    id varchar(64) NOT NULL UNIQUE,
    username varchar(128) NOT NULL UNIQUE,
    team_name varchar(64) NOT NULL UNIQUE,
    is_active bool NOT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE teams (
    name varchar(128) NOT NULL UNIQUE,
    PRIMARY KEY(name)
);

CREATE TABLE pull_requests(
    id varchar(64) NOT NULL UNIQUE,
    name varchar(128) NOT NULL,
    author_id varchar(64) NOT NULL,
    status varchar(128) NOT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE pull_requests_users (
    user_id varchar(64) NOT NULL,
    pr_id varchar(64) NOT NULL,
    PRIMARY KEY (user_id, pr_id)
);

CREATE UNIQUE INDEX idx_users_id ON users (id);
CREATE UNIQUE INDEX idx_teams_name ON teams (name);
CREATE UNIQUE INDEX idx_pull_requests_id ON pull_requests (id);

CREATE INDEX idx_pull_requests_users_user_id ON pull_requests_users (user_id);
CREATE INDEX idx_pull_requests_users_pr_id ON pull_requests_users (pr_id);

CREATE INDEX idx_users_team_name ON users (team_name);
CREATE INDEX idx_users_id_active ON users (id) WHERE is_active = true;

ALTER TABLE users ADD CONSTRAINT FK_users_1 FOREIGN KEY (team_name) REFERENCES teams (name);
ALTER TABLE pull_requests ADD CONSTRAINT FK_users_2 FOREIGN KEY (author_id) REFERENCES users (id);
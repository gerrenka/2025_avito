CREATE TABLE IF NOT EXISTS teams (
    team_name VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);


CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_name VARCHAR(255) NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);


CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);


CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);


CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(500) NOT NULL,
    author_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMP NULL
);


CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id ON pull_requests(author_id);


CREATE INDEX IF NOT EXISTS idx_pull_requests_status ON pull_requests(status);


CREATE TABLE IF NOT EXISTS pr_reviewers (
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (pull_request_id, user_id)
);


CREATE INDEX IF NOT EXISTS idx_pr_reviewers_user_id ON pr_reviewers(user_id);


COMMENT ON TABLE teams IS 'Teams of users';
COMMENT ON TABLE users IS 'Team members who can be assigned as reviewers';
COMMENT ON TABLE pull_requests IS 'Pull requests requiring code review';
COMMENT ON TABLE pr_reviewers IS 'Junction table mapping reviewers to pull requests';

COMMENT ON COLUMN users.is_active IS 'Whether user can be assigned as reviewer';
COMMENT ON COLUMN pull_requests.status IS 'PR status: OPEN or MERGED';
COMMENT ON COLUMN pull_requests.merged_at IS 'Timestamp when PR was merged (NULL if not merged)';
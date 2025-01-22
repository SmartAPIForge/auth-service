ALTER TABLE users
ADD COLUMN role_id INT DEFAULT 2;

ALTER TABLE users
ADD CONSTRAINT fk_role
FOREIGN KEY (role_id) REFERENCES role (id);

CREATE INDEX IF NOT EXISTS idx_role_id ON users (role_id);
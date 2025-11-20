CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50),
    password VARCHAR(50)
);

-- Seeding data with plain text passwords
INSERT INTO users (username, password) VALUES ('admin', 'supersecretkey');
INSERT INTO users (username, password) VALUES ('dmitry', 'football2023');
INSERT INTO users (username, password) VALUES ('guest', 'guest123');
-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id VARCHAR(50) PRIMARY KEY, -- go's uuid converted to string (32 bytes or 36 bytes with dashes)
    name VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL -- надо посмотреть, сколько места занимает хэш бкрипта
    -- bitrhday
    -- address
);

CREATE INDEX email_idx ON users (email);

CREATE TABLE admins ( -- следят за магазином
    id VARCHAR(50) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    isCore BOOLEAN NOT NULL -- можно добавлять и удалять других админов + модеров
    -- другие возможные права
);

CREATE TABLE moders( -- только комментарии чистят
    id VARCHAR(50) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE
    -- другие возможные права
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
DROP INDEX email_idx;
DROP TABLE admins;
DROP TABLE moders;
-- +goose StatementEnd

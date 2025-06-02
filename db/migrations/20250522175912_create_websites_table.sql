-- +goose Up
-- +goose StatementBegin
CREATE TABLE webpages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT UNIQUE NOT NULL,
    title TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE webpages;
-- +goose StatementEnd

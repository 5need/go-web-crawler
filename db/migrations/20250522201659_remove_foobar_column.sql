-- +goose Up
-- +goose StatementBegin
ALTER TABLE webpages DROP COLUMN foobar;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE webpages ADD COLUMN foobar TEXT;
-- +goose StatementEnd

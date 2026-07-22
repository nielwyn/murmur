-- +goose Up
ALTER TABLE feeds RENAME COLUMN name TO title;
ALTER TABLE feeds RENAME COLUMN url TO link;
ALTER TABLE posts RENAME COLUMN url TO link;

-- +goose Down
ALTER TABLE posts RENAME COLUMN link TO url;
ALTER TABLE feeds RENAME COLUMN link TO url;
ALTER TABLE feeds RENAME COLUMN title TO name;

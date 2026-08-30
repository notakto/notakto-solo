-- +goose Up
-- +goose StatementBegin
ALTER TABLE Player
    ADD COLUMN profile_image_file_id TEXT,
    ADD COLUMN profile_image_file_path TEXT,
    ADD CONSTRAINT player_profile_image_fields_paired CHECK (
        (profile_image_file_id IS NULL) = (profile_image_file_path IS NULL)
    ),
    ADD CONSTRAINT player_profile_image_file_id_unique UNIQUE (profile_image_file_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE Player
    DROP CONSTRAINT IF EXISTS player_profile_image_file_id_unique,
    DROP CONSTRAINT IF EXISTS player_profile_image_fields_paired,
    DROP COLUMN IF EXISTS profile_image_file_path,
    DROP COLUMN IF EXISTS profile_image_file_id;
-- +goose StatementEnd

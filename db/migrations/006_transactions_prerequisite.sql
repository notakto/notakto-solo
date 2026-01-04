-- +goose Up
ALTER TABLE player
ADD CONSTRAINT player_uid_pk PRIMARY KEY (uid);

ALTER TABLE wallet
ADD CONSTRAINT wallet_uid_pk PRIMARY KEY (uid);

ALTER TABLE wallet
ADD CONSTRAINT wallet_uid_fk
FOREIGN KEY (uid) REFERENCES player(uid)
ON DELETE CASCADE;

-- +goose Down
ALTER TABLE wallet
DROP CONSTRAINT wallet_uid_fk;

ALTER TABLE wallet
DROP CONSTRAINT wallet_uid_pk;

ALTER TABLE player
DROP CONSTRAINT player_uid_pk;

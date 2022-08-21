BEGIN;

CREATE TABLE IF NOT EXISTS email_types
(
    id          SERIAL PRIMARY KEY,
    "name"      VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL
);

INSERT INTO email_types (name, description)
VALUES ('erc_register', 'Реестр выданных купонов от ЕРЦ'),
       ('correction', 'Исправление данных по купонам');

CREATE TABLE IF NOT EXISTS emails
(
    id                SERIAL PRIMARY KEY,
    type_id           INTEGER                  NOT NULL,
    message_id        VARCHAR(255)             NOT NULL UNIQUE,
    from_address      VARCHAR(255)             NOT NULL,
    datetime_received TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    datetime_parsed   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    file              bytea                    NOT NULL
);
ALTER TABLE emails
    ADD CONSTRAINT emails_type_id_fkey FOREIGN KEY (type_id) REFERENCES email_types (id);
CREATE INDEX IF NOT EXISTS emails_datetime_received_idx ON emails (datetime_received);

CREATE TABLE IF NOT EXISTS erc_updates
(
    id       SERIAL PRIMARY KEY,
    email_id INTEGER NOT NULL,
    "name"   VARCHAR NOT NULL
);
ALTER TABLE erc_updates
    ADD CONSTRAINT erc_updates_email_id_fkey FOREIGN KEY (email_id) REFERENCES emails (id);

CREATE TABLE IF NOT EXISTS persons_from_erc
(
    "id"            SERIAL PRIMARY KEY,
    "erc_update_id" INTEGER NOT NULL REFERENCES erc_updates (id) ON DELETE CASCADE,
    "snils"         VARCHAR NOT NULL,
    "birthdate"     DATE    NOT NULL,
    "family"        VARCHAR NOT NULL,
    "name"          VARCHAR NOT NULL,
    "patronymic"    VARCHAR NOT NULL DEFAULT '',
    "year"          INTEGER NOT NULL,
    "semester"      INTEGER NOT NULL,
    "color"         VARCHAR NOT NULL,
    "count"         INTEGER NOT NULL,
    "spent"         INTEGER NOT NULL,
    "date"          DATE    NOT NULL,
    cashier_id      INTEGER NOT NULL DEFAULT 0,
    cashier_name    VARCHAR NOT NULL DEFAULT '',
    errors          VARCHAR[]
);

CREATE TABLE IF NOT EXISTS rstk_update_types
(
    id          SERIAL PRIMARY KEY,
    "name"      VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL
);

INSERT INTO rstk_update_types (name, description)
VALUES ('stk', 'Список социальных карт'),
       ('mir', 'Список банковских карт');

CREATE TABLE IF NOT EXISTS rstk_updates
(
    id          SERIAL PRIMARY KEY,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    from_date   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    "type_id"   INTEGER                  NOT NULL REFERENCES rstk_update_types (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS persons_from_rstk
(
    "id"             SERIAL PRIMARY KEY,
    "rstk_update_id" INTEGER NOT NULL REFERENCES rstk_updates (id) ON DELETE CASCADE,
    "snils"          VARCHAR NOT NULL,
    "family"         VARCHAR NOT NULL,
    "name"           VARCHAR NOT NULL,
    "patronymic"     VARCHAR NOT NULL DEFAULT '',
    "date"           DATE    NOT NULL,
    "number"         VARCHAR NOT NULL UNIQUE,
    errors           VARCHAR[]
);

CREATE TABLE IF NOT EXISTS sent_to_erc
(
    "id"    SERIAL PRIMARY KEY,
    "snils" VARCHAR NOT NULL UNIQUE,
    "date"  DATE    NOT NULL
);

CREATE TABLE IF NOT EXISTS correct_person_data
(
    "snils"      VARCHAR(11) PRIMARY KEY,
    "family"     VARCHAR NOT NULL,
    "name"       VARCHAR NOT NULL,
    "patronymic" VARCHAR NOT NULL DEFAULT '',
    "birthdate"  DATE    NOT NULL
);

COMMIT;
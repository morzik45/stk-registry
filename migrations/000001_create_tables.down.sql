BEGIN;

DROP TABLE IF EXISTS correct_person_data;
DROP TABLE IF EXISTS sent_to_erc;
DROP TABLE IF EXISTS persons_from_erc;
DROP TABLE IF EXISTS persons_from_rstk;
DROP INDEX IF EXISTS erc_updates_datetime_received_idx;
ALTER TABLE erc_updates
    DROP CONSTRAINT IF EXISTS erc_updates_email_id_fkey;
DROP TABLE IF EXISTS erc_updates;
DROP INDEX IF EXISTS emails_datetime_received_idx;
ALTER TABLE emails
    DROP CONSTRAINT IF EXISTS emails_type_id_fkey;
DROP TABLE IF EXISTS emails;
DROP TABLE IF EXISTS email_types;
DROP TABLE IF EXISTS rstk_updates;
DROP TABLE IF EXISTS rstk_update_types;

COMMIT;
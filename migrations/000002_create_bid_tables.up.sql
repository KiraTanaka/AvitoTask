CREATE TYPE bid_author_type AS ENUM (
    'Organization',
    'User'
    );
CREATE TYPE bid_decision_type AS ENUM (
    'Approved',
    'Rejected'
    );

CREATE TYPE bid_status AS ENUM (
    'Created',
    'Published',
    'Canceled'
    );

CREATE TABLE bid
(
    id          uuid PRIMARY KEY             DEFAULT gen_random_uuid(),
    name        VARCHAR(100)                                           NOT NULL,
    description VARCHAR(500)                                           NOT NULL,
    status      bid_status                                             NOT NULL,
    tender_id   uuid                                                   NOT NULL REFERENCES tender (id) ON DELETE CASCADE,
    author_type bid_author_type                                        NOT NULL,
    author_id   uuid                                                   NOT NULL,
    version     INTEGER CHECK (version >= 1) DEFAULT 1                 NOT NULL,
    created_at  TIMESTAMP                    DEFAULT CURRENT_TIMESTAMP NOT NULL,
    decision    bid_decision_type
);

CREATE TABLE bid_version_hist
(
    id         SERIAL PRIMARY KEY,
    bid_id     uuid                                NOT NULL REFERENCES bid (id) ON DELETE CASCADE,
    version    INTEGER CHECK (version >= 1)        NOT NULL,
    params     jsonb                               NOT NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE OR REPLACE FUNCTION bid_version_hist_update_trigger_func()
    RETURNS TRIGGER
    LANGUAGE 'plpgsql' AS
$$
DECLARE
    params jsonb :='{}'::jsonb;
BEGIN
    IF new.name IS DISTINCT FROM old.name OR new.description IS DISTINCT FROM old.description THEN

        params = FORMAT('{"name":"%s"}', old.name)::jsonb ||
            FORMAT('{"description":"%s"}', old.description)::jsonb ||
            FORMAT('{"status":"%s"}', old.status)::jsonb;

        IF params IS DISTINCT FROM '{}'::jsonb THEN
            INSERT INTO bid_version_hist (bid_id, version, params)
            VALUES (old.id, new.version, params);
            new.version = new.version + 1;
        END IF;
    END IF;
    RETURN new;
END;
$$;

CREATE TRIGGER write_hist
    BEFORE UPDATE
    ON bid
    FOR EACH ROW
EXECUTE PROCEDURE bid_version_hist_update_trigger_func();

CREATE TABLE bid_decision
(
    id       SERIAL PRIMARY KEY,
    bid_id   uuid              NOT NULL REFERENCES bid (id) ON DELETE CASCADE,
    username VARCHAR(50)       NOT NULL,
    decision bid_decision_type NOT NULL
);
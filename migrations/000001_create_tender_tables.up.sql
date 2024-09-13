CREATE TYPE service_type AS ENUM (
    'Construction',
    'Delivery',
    'Manufacture'
    );

CREATE TYPE tender_status AS ENUM (
    'Created',
    'Published',
    'Closed'
    );

CREATE TABLE tender
(
    id              uuid PRIMARY KEY             DEFAULT gen_random_uuid(),
    name            VARCHAR(100)                                           NOT NULL,
    description     VARCHAR(500)                                           NOT NULL,
    service_type    service_type                                           NOT NULL,
    status          tender_status                                          NOT NULL,
    organization_id uuid                                                   NOT NULL REFERENCES organization (id) ON DELETE CASCADE,
    version         INTEGER CHECK (version >= 1) DEFAULT 1                 NOT NULL,
    created_at      TIMESTAMP                    DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE tender_version_hist
(
    id         SERIAL PRIMARY KEY,
    tender_id  uuid                                NOT NULL REFERENCES tender (id) ON DELETE CASCADE,
    version    INTEGER CHECK (version >= 1)        NOT NULL,
    params     jsonb                               NOT NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE OR REPLACE FUNCTION tender_version_hist_update_trigger_func()
    RETURNS TRIGGER
    LANGUAGE 'plpgsql' AS
$$
DECLARE
    params jsonb :='{}'::jsonb;
BEGIN
    IF new.name IS DISTINCT FROM old.name OR new.description IS DISTINCT FROM old.description
        OR new.service_type IS DISTINCT FROM old.service_type THEN

        params = FORMAT('{"name":"%s"}', old.name)::jsonb ||
            FORMAT('{"description":"%s"}', old.description)::jsonb ||
            FORMAT('{"serviceType":"%s"}', old.service_type)::jsonb ||
            FORMAT('{"status":"%s"}', old.status)::jsonb ||
            FORMAT('{"organizationId":"%s"}', old.organization_id)::jsonb;

        IF params IS DISTINCT FROM '{}'::jsonb THEN
            INSERT INTO tender_version_hist (tender_id, version, params)
            VALUES (old.id, new.version, params);
            new.version = new.version + 1;
        END IF;
    END IF;
    RETURN new;
END;
$$;

CREATE TRIGGER write_hist
    BEFORE UPDATE
    ON tender
    FOR EACH ROW
EXECUTE PROCEDURE tender_version_hist_update_trigger_func();
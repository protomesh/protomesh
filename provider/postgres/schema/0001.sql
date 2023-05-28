-- +migrate Up
CREATE TABLE resource_events (
    namespace TEXT NOT NULL,
    id UUID NOT NULL,
    version_index BIGINT NOT NULL,
    status VARCHAR(25) NOT NULL, 
    name TEXT NOT NULL,
    spec_type_url TEXT NOT NULL,
    spec_value BYTEA DEFAULT NULL,
    sha256_hash CHAR(43) NOT NULL,
    PRIMARY KEY (namespace, version_index, id)
);
CREATE INDEX idx_resource_events_version ON resource_events (namespace, version_index);

-- +migrate Down
DROP INDEX idx_resource_events_version;
DROP TABLE resource_events;
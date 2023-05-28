-- +migrate Up
CREATE TABLE resource_cache (
    namespace TEXT NOT NULL,
    id UUID NOT NULL,
    version_index BIGINT NOT NULL,
    name TEXT NOT NULL,
    spec_type_url TEXT NOT NULL,
    spec_value BYTEA NOT NULL,
    sha256_hash CHAR(43) NOT NULL,
    PRIMARY KEY (namespace, id)
);
CREATE INDEX idx_resource_cache_version ON resource_cache (namespace, version_index);

-- +migrate Down
DROP INDEX idx_resource_cache_version;
DROP TABLE resource_cache;

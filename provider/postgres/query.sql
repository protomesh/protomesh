-- name: UpsertResourceCache :exec
INSERT INTO resource_cache (
    namespace,
    id,
    version_index,
    name,
    spec_type_url,
    spec_value,
    sha256_hash
) VALUES (
    @namespace::TEXT,
    @id::UUID,
    @version_index::BIGINT,
    @name::TEXT,
    @spec_type_url::TEXT,
    @spec_value::BYTEA,
    @sha256_hash::TEXT
) ON CONFLICT (namespace, id)
DO UPDATE SET 
    version_index = EXCLUDED.version_index,
    name = EXCLUDED.name,
    spec_type_url = EXCLUDED.spec_type_url,
    spec_value = EXCLUDED.spec_value,
    sha256_hash = EXCLUDED.sha256_hash;

-- name: InsertActiveResourceEvent :exec
INSERT INTO resource_events (
    namespace,
    id,
    version_index,
    status,
    name,
    spec_type_url,
    spec_value,
    sha256_hash
) SELECT
    rc.namespace,
    rc.id,
    rc.version_index,
    'ACTIVE'::text,
    rc.name,
    rc.spec_type_url,
    rc.spec_value,
    rc.sha256_hash
FROM resource_cache AS rc
WHERE rc.namespace = @namespace::TEXT AND
rc.id = @id::UUID;

-- name: DropResourceCache :exec
DELETE FROM resource_cache WHERE namespace = @namespace::TEXT AND id = @id::UUID;

-- name: InsertDroppedResourceEvent :exec
INSERT INTO resource_events (
    namespace,
    id,
    version_index,
    status,
    name,
    spec_type_url,
    spec_value,
    sha256_hash
) SELECT
    rc.namespace,
    rc.id,
    @version_index::BIGINT,
    'DROPPED'::TEXT,
    rc.name,
    rc.spec_type_url,
    rc.spec_value,
    rc.sha256_hash
FROM resource_cache AS rc
WHERE rc.namespace = @namespace::TEXT AND
rc.id = @id::UUID;

-- name: CountResourceCacheBefore :one
SELECT count(*) FROM resource_cache WHERE namespace = @namespace::TEXT AND version_index < @before_version_index::BIGINT;

-- name: DropResourceCacheBefore :exec
DELETE FROM resource_cache WHERE namespace = @namespace::TEXT AND version_index < @before_version_index::BIGINT;

-- name: InsertDroppedResourceBeforeEvent :exec
INSERT INTO resource_events (
    namespace,
    id,
    version_index,
    status,
    name,
    spec_type_url,
    spec_value,
    sha256_hash
) SELECT
    rc.namespace,
    rc.id,
    @version_index::BIGINT,
    'DROPPED'::text,
    rc.name,
    rc.spec_type_url,
    rc.spec_value,
    rc.sha256_hash
FROM resource_cache AS rc
WHERE rc.namespace = @namespace::TEXT AND
rc.version_index < @before_version_index::BIGINT;

-- name: GetResourceCacheSummary :one
SELECT
    version_index,
    sha256_hash
FROM resource_cache AS rc WHERE rc.namespace = @namespace::TEXT AND rc.id = @id::UUID;

-- name: GetResourceCache :one
SELECT
    rc.version_index,
    rc.name,
    rc.spec_type_url,
    rc.spec_value,
    rc.sha256_hash
FROM resource_cache AS rc
WHERE rc.namespace = @namespace::TEXT AND rc.id = @id::UUID;

-- name: MaxVersionIndexForNamespace :one
SELECT
    id,
    version_index
FROM
    resource_events AS re
WHERE
    re.namespace = @namespace::TEXT
ORDER BY
    re.namespace DESC,
    re.version_index DESC,
    re.id DESC
LIMIT 1;

-- name: ListResourcesEventsFromNamespace :many
SELECT
    id,
    version_index,
    status,
    name,
    spec_type_url,
    spec_value,
    sha256_hash
FROM
    resource_events AS re
WHERE
    (re.namespace, re.version_index, re.id) > (@namespace::TEXT, @from_version_index::BIGINT, @from_id::UUID)
    AND rc.namespace = @namespace::TEXT
ORDER BY
    re.namespace ASC,
    re.version_index ASC,
    re.id ASC
LIMIT @max_rows OFFSET 0;

-- name: ListResourcesCachedFromNamespace :many
SELECT
    id,
    version_index,
    name,
    spec_type_url,
    spec_value,
    sha256_hash
FROM
    resource_cache AS re
WHERE
    re.namespace = @namespace::TEXT
ORDER BY
    re.namespace ASC,
    re.version_index ASC
LIMIT @max_rows OFFSET @rows_offset;
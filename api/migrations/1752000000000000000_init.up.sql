BEGIN;

CREATE TABLE posts (
    id              BIGINT       NOT NULL PRIMARY KEY,
    group_id        BIGINT       NOT NULL,
    title           TEXT,
    url             TEXT UNIQUE,
    text            TEXT,
    seo_title       TEXT,
    seo_description TEXT,
    seo_keywords    TEXT,
    created_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP
);

CREATE TABLE media (
    tg_message_id BIGINT           NOT NULL PRIMARY KEY,
    post_id       BIGINT           NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
    file_id       BIGINT           NOT NULL,
    type          VARCHAR(16)      NOT NULL,
    url           TEXT             NOT NULL,
    mime_type     VARCHAR(100),
    file_name     TEXT,
    size_bytes    BIGINT,
    width         INT,
    height        INT,
    duration      DOUBLE PRECISION,
    created_at    TIMESTAMP        NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP
);

CREATE INDEX media_post_id_idx ON media (post_id);

CREATE TABLE channels (
    id              BIGINT       NOT NULL PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    avatar          TEXT,
    subscriptions   BIGINT       NOT NULL DEFAULT 0,
    last_message_id BIGINT       NOT NULL DEFAULT 0,
    created_at      TIMESTAMP    NOT NULL DEFAULT NOW()
);

COMMIT;

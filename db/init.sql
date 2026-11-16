CREATE TABLE IF NOT EXISTS subscriptions (
                                             id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                                             chat_id BIGINT NOT NULL,
                                             repository TEXT NOT NULL,
                                             created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                             UNIQUE (chat_id, repository)
);

CREATE TABLE IF NOT EXISTS releases (
                                        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                                        repository TEXT NOT NULL,
                                        tag TEXT NOT NULL,
                                        url TEXT NOT NULL,
                                        published_at TIMESTAMPTZ NOT NULL,

                                        UNIQUE (repository, tag)
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_repository
    ON subscriptions (repository);

CREATE TABLE IF NOT EXISTS notifications (
                                             id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                                             release_id BIGINT NOT NULL REFERENCES releases(id),
                                             chat_id BIGINT NOT NULL,
                                             created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                             sent_at TIMESTAMPTZ,

                                             UNIQUE (release_id, chat_id)
);
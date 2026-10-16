-- +goose Up
CREATE TABLE categories (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent_id  bigint REFERENCES categories (id) ON DELETE RESTRICT,
    slug       text NOT NULL,
    name       text NOT NULL,
    position   integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    -- slug harus unik di antara kategori yang satu induk; NULLS NOT DISTINCT
    -- membuat aturan ini berlaku juga untuk kategori tingkat atas (parent_id NULL)
    UNIQUE NULLS NOT DISTINCT (parent_id, slug)
);

CREATE INDEX categories_parent_id_idx ON categories (parent_id);

-- +goose Down
DROP TABLE categories;

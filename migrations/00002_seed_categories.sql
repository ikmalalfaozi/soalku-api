-- +goose Up
INSERT INTO categories (slug, name, position) VALUES ('cpns', 'CPNS', 1);

INSERT INTO categories (parent_id, slug, name, position)
SELECT id, v.slug, v.name, v.position
FROM categories,
     (VALUES
        ('twk', 'Tes Wawasan Kebangsaan (TWK)', 1),
        ('tiu', 'Tes Intelegensia Umum (TIU)', 2),
        ('tkp', 'Tes Karakteristik Pribadi (TKP)', 3)
     ) AS v (slug, name, position)
WHERE categories.parent_id IS NULL AND categories.slug = 'cpns';

-- +goose Down
-- Anak harus dihapus lebih dulu karena parent_id memakai ON DELETE RESTRICT.
DELETE FROM categories
WHERE parent_id = (SELECT id FROM categories WHERE parent_id IS NULL AND slug = 'cpns');

DELETE FROM categories WHERE parent_id IS NULL AND slug = 'cpns';

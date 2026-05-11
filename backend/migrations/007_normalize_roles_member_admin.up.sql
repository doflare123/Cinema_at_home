UPDATE roles SET name = 'member' WHERE name = 'user' AND NOT EXISTS (SELECT 1 FROM roles WHERE name = 'member');
DELETE FROM roles WHERE name = 'user' AND EXISTS (SELECT 1 FROM roles WHERE name = 'member');

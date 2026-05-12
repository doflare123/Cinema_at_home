UPDATE users
SET role_id = member_role.id
FROM roles AS legacy_role
JOIN roles AS member_role ON member_role.name = 'member'
WHERE legacy_role.name = 'user'
  AND users.role_id = legacy_role.id;

UPDATE roles
SET name = 'member'
WHERE name = 'user'
  AND NOT EXISTS (SELECT 1 FROM roles WHERE name = 'member');

DELETE FROM roles
WHERE name = 'user'
  AND EXISTS (SELECT 1 FROM roles WHERE name = 'member');

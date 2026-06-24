UPDATE users SET role = 'admin' WHERE role = 'internal_admin';
UPDATE users SET role = 'user' WHERE role = 'restaurant_owner';

DROP TABLE IF EXISTS restaurant_members;
DROP TABLE IF EXISTS restaurants;

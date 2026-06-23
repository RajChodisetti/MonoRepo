-- Role model (P1-008): restaurants + membership for tenant-scoped access.

CREATE TABLE IF NOT EXISTS restaurants (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS restaurant_members (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  restaurant_id uuid NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  member_role text NOT NULL DEFAULT 'owner',
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT restaurant_members_unique UNIQUE (restaurant_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_restaurant_members_user_id
  ON restaurant_members (user_id);

CREATE INDEX IF NOT EXISTS idx_restaurant_members_restaurant_id
  ON restaurant_members (restaurant_id);

-- Align existing user roles with the P1-008 role model.
UPDATE users SET role = 'internal_admin' WHERE role = 'admin';
UPDATE users SET role = 'restaurant_owner' WHERE role = 'user';

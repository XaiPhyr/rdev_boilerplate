-- User Roles table
CREATE TABLE IF NOT EXISTS user_roles (
  id bigserial primary key,
  user_id bigint not null references users(id) on delete cascade,
  role_id bigint not null references roles(id) on delete cascade,
  active boolean not null default true,
  status varchar(1) not null default 'O',
  flag varchar(45),
  uuid uuid not null default gen_random_uuid() unique,
  created_at timestamptz not null default now(),
  created_by bigint default 0,
  updated_at timestamptz not null default now(),
  updated_by bigint default 0,
  deleted_at timestamptz,
  deleted_by bigint default 0,
  CONSTRAINT unique_user_role UNIQUE (user_id, role_id, deleted_at)
);

CREATE UNIQUE INDEX unique_user_role_active ON user_roles(user_id, role_id)
WHERE deleted_at IS NULL;
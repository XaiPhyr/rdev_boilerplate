-- Groups table
CREATE TABLE IF NOT EXISTS groups (
  id bigserial primary key,
  name varchar(255) not null unique,
  description text,
  active boolean not null default true,
  status varchar(1) not null default 'O',
  flag varchar(45),
  uuid uuid not null default gen_random_uuid() unique,
  created_at timestamptz not null default now(),
  created_by bigint default 0,
  updated_at timestamptz not null default now(),
  updated_by bigint default 0,
  deleted_at timestamptz,
  deleted_by bigint default 0
);
create table "households" (
  "id" character varying not null,
  "created_at" timestamptz not null,
  "updated_at" timestamptz not null,
  "name" character varying not null,
  "slug" character varying not null,
  "description" character varying not null default '',
  "active" boolean not null default true,
  primary key ("id")
);

create unique index "households_slug_key" on "households" ("slug");

create table "users" (
  "id" character varying not null,
  "created_at" timestamptz not null,
  "updated_at" timestamptz not null,
  "display_name" character varying not null,
  "email" character varying not null,
  "email_verified" boolean not null default false,
  "password_hash" character varying not null,
  "role" character varying not null default 'user',
  "last_login_at" timestamptz null,
  primary key ("id")
);

create unique index "users_email_key" on "users" ("email");

create table "household_members" (
  "id" character varying not null,
  "created_at" timestamptz not null,
  "updated_at" timestamptz not null,
  "role" character varying not null default 'viewer',
  "household_id" character varying not null,
  "user_id" character varying not null,
  primary key ("id"),
  constraint "household_members_households_members" foreign KEY ("household_id") references "households" ("id") on delete no action,
  constraint "household_members_users_memberships" foreign KEY ("user_id") references "users" ("id") on delete no action
);

create unique index "householdmember_household_id_user_id" on "household_members" ("household_id", "user_id");

create table "recipes" (
  "id" character varying not null,
  "created_at" timestamptz not null,
  "updated_at" timestamptz not null,
  "title" character varying not null,
  "description" character varying not null default '',
  "source_url" character varying not null default '',
  "source_captured_at" timestamptz null,
  "primary_media_id" character varying null,
  "gallery_media_ids" jsonb null,
  "prep_minutes" bigint null,
  "cook_minutes" bigint null,
  "servings" bigint null,
  "region" character varying null,
  "meal_type" character varying null,
  "difficulty" character varying null,
  "cuisine" character varying null,
  "popularity_score" double precision null default 0,
  "allergens" jsonb null,
  "aggregated_at" timestamptz null,
  "version" bigint not null default 1,
  "household_id" character varying not null,
  primary key ("id"),
  constraint "recipes_households_recipes" foreign KEY ("household_id") references "households" ("id") on delete no action
);

create index "recipe_household_id_title" on "recipes" ("household_id", "title");

create table "media_assets" (
  "id" character varying not null,
  "created_at" timestamptz not null,
  "updated_at" timestamptz not null,
  "original_filename" character varying not null,
  "mime_type" character varying not null,
  "media_type" character varying not null default 'image',
  "size_bytes" bigint not null,
  "checksum" character varying not null,
  "stored_at" timestamptz not null,
  "household_id" character varying not null,
  "recipe_id" character varying null,
  primary key ("id"),
  constraint "media_assets_households_media_assets" foreign KEY ("household_id") references "households" ("id") on delete no action,
  constraint "media_assets_recipes_media_assets" foreign KEY ("recipe_id") references "recipes" ("id") on delete set null
);

create unique index "mediaasset_household_id_checksum" on "media_assets" ("household_id", "checksum");

create table "recipe_ingredients" (
  "id" character varying not null,
  "created_at" timestamptz not null,
  "updated_at" timestamptz not null,
  "name" character varying not null,
  "quantity" double precision null,
  "unit" character varying null,
  "preparation" character varying null,
  "sort_order" bigint not null,
  "recipe_id" character varying not null,
  primary key ("id"),
  constraint "recipe_ingredients_recipes_ingredients" foreign KEY ("recipe_id") references "recipes" ("id") on delete cascade
);

create index "recipeingredient_recipe_id_sort_order" on "recipe_ingredients" ("recipe_id", "sort_order");

create table "recipe_nutrition" (
  "id" character varying not null,
  "created_at" timestamptz not null,
  "updated_at" timestamptz not null,
  "reference_quantity" character varying null,
  "energy_kj" bigint null,
  "energy_kcal" bigint null,
  "protein" double precision null,
  "carbohydrates" double precision null,
  "fat" double precision null,
  "saturated_fat" double precision null,
  "fiber" double precision null,
  "sugars" double precision null,
  "sodium" double precision null,
  "salt" double precision null,
  "breakdown" jsonb null,
  "recipe_id" character varying not null,
  primary key ("id"),
  constraint "recipe_nutrition_recipes_nutrition_entries" foreign KEY ("recipe_id") references "recipes" ("id") on delete cascade
);

create unique index "recipenutrition_recipe_id_reference_quantity" on "recipe_nutrition" ("recipe_id", "reference_quantity");

create table "recipe_shares" (
  "id" character varying not null,
  "created_at" timestamptz not null,
  "updated_at" timestamptz not null,
  "token_hash" character varying not null,
  "expires_at" timestamptz null,
  "revoked_at" timestamptz null,
  "last_accessed_at" timestamptz null,
  "access_count" bigint not null default 0,
  "recipe_id" character varying not null,
  "created_by_user_id" character varying not null,
  primary key ("id"),
  constraint "recipe_shares_recipes_shares" foreign KEY ("recipe_id") references "recipes" ("id") on delete cascade,
  constraint "recipe_shares_users_recipe_shares" foreign KEY ("created_by_user_id") references "users" ("id") on delete no action
);

create unique index "recipe_shares_token_hash_key" on "recipe_shares" ("token_hash");

create index "recipeshare_recipe_id" on "recipe_shares" ("recipe_id");

create index "recipeshare_created_by_user_id" on "recipe_shares" ("created_by_user_id");

create index "recipeshare_expires_at" on "recipe_shares" ("expires_at");

create table "recipe_steps" (
  "id" character varying not null,
  "created_at" timestamptz not null,
  "updated_at" timestamptz not null,
  "sort_order" bigint not null,
  "instruction" character varying not null,
  "duration_minutes" bigint null,
  "tip" character varying null,
  "recipe_id" character varying not null,
  primary key ("id"),
  constraint "recipe_steps_recipes_steps" foreign KEY ("recipe_id") references "recipes" ("id") on delete cascade
);

create index "recipestep_recipe_id_sort_order" on "recipe_steps" ("recipe_id", "sort_order");

create table "tags" (
  "id" character varying not null,
  "created_at" timestamptz not null,
  "updated_at" timestamptz not null,
  "name" character varying not null,
  "slug" character varying not null,
  "color" character varying not null default '',
  "system" boolean not null default false,
  "household_id" character varying not null,
  primary key ("id"),
  constraint "tags_households_tags" foreign KEY ("household_id") references "households" ("id") on delete no action
);

create unique index "tag_household_id_slug" on "tags" ("household_id", "slug");

create table "recipe_tags" (
  "recipe_id" character varying not null,
  "tag_id" character varying not null,
  primary key ("recipe_id", "tag_id"),
  constraint "recipe_tags_recipe_id" foreign KEY ("recipe_id") references "recipes" ("id") on delete cascade,
  constraint "recipe_tags_tag_id" foreign KEY ("tag_id") references "tags" ("id") on delete cascade
);

create table "refresh_sessions" (
  "id" character varying not null,
  "created_at" timestamptz not null,
  "updated_at" timestamptz not null,
  "token_hash" character varying not null,
  "expires_at" timestamptz not null,
  "revoked" boolean not null default false,
  "device_info" character varying null,
  "ip_address" character varying null,
  "last_used_at" timestamptz null,
  "user_id" character varying not null,
  primary key ("id"),
  constraint "refresh_sessions_users_refresh_sessions" foreign KEY ("user_id") references "users" ("id") on delete no action
);

create unique index "refresh_sessions_token_hash_key" on "refresh_sessions" ("token_hash");

create index "refreshsession_user_id_revoked" on "refresh_sessions" ("user_id", "revoked");

create index "refreshsession_expires_at" on "refresh_sessions" ("expires_at");

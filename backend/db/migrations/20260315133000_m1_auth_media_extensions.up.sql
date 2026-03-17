alter table "refresh_sessions"
  add column "active_household_id" character varying;

update "refresh_sessions" as rs
set "active_household_id" = membership."household_id"
from (
  select distinct on ("user_id")
    "user_id",
    "household_id"
  from "household_members"
  order by "user_id", "created_at", "id"
) as membership
where membership."user_id" = rs."user_id"
  and rs."active_household_id" is null;

alter table "refresh_sessions"
  alter column "active_household_id" set not null;

alter table "refresh_sessions"
  add constraint "refresh_sessions_households_refresh_sessions"
  foreign key ("active_household_id") references "households" ("id") on delete no action;

create index "refreshsession_active_household_id" on "refresh_sessions" ("active_household_id");

drop index if exists "mediaasset_household_id_checksum";

create table "stored_objects" (
  "id" character varying not null,
  "created_at" timestamptz not null,
  "updated_at" timestamptz not null,
  "original_filename" character varying not null,
  "mime_type" character varying not null,
  "size_bytes" bigint not null,
  "checksum" character varying not null,
  "content" bytea not null,
  "household_id" character varying not null,
  primary key ("id"),
  constraint "stored_objects_households_stored_objects"
    foreign key ("household_id") references "households" ("id") on delete no action
);

create unique index "storedobject_household_id_checksum" on "stored_objects" ("household_id", "checksum");

insert into "stored_objects" (
  "id",
  "created_at",
  "updated_at",
  "original_filename",
  "mime_type",
  "size_bytes",
  "checksum",
  "content",
  "household_id"
)
select distinct on ("household_id", "checksum")
  "id",
  "created_at",
  "updated_at",
  coalesce(
    nullif(
      btrim(
        regexp_replace("original_filename", '[[:cntrl:]/\\"]', '_', 'g'),
        ' .'
      ),
      ''
    ),
    'download'
  ),
  "mime_type",
  "size_bytes",
  "checksum",
  decode('', 'hex'),
  "household_id"
from "media_assets"
order by "household_id", "checksum", "id";

alter table "media_assets"
  add column "storage_object_id" character varying,
  add column "alt_text" character varying not null default '',
  add column "sort_order" bigint not null default 1;

with ranked_media as (
  select
    "id",
    row_number() over (partition by "recipe_id" order by "created_at", "id") as "row_num"
  from "media_assets"
  where "recipe_id" is not null
)
update "media_assets" as ma
set "sort_order" = ranked_media."row_num"
from ranked_media
where ranked_media."id" = ma."id";

update "media_assets" as ma
set "storage_object_id" = so."id"
from "stored_objects" as so
where so."household_id" = ma."household_id"
  and so."checksum" = ma."checksum"
  and ma."storage_object_id" is null;

alter table "media_assets"
  add constraint "media_assets_stored_objects_media_assets"
  foreign key ("storage_object_id") references "stored_objects" ("id") on delete no action;

alter table "media_assets"
  alter column "storage_object_id" set not null;

create index "mediaasset_recipe_id_sort_order" on "media_assets" ("recipe_id", "sort_order");

drop index if exists "mediaasset_recipe_id_sort_order";

alter table "media_assets"
  drop constraint if exists "media_assets_stored_objects_media_assets";

alter table "media_assets"
  drop column if exists "sort_order",
  drop column if exists "alt_text",
  drop column if exists "storage_object_id";

create unique index "mediaasset_household_id_checksum" on "media_assets" ("household_id", "checksum");

drop index if exists "storedobject_household_id_checksum";
drop table if exists "stored_objects";

drop index if exists "refreshsession_active_household_id";

alter table "refresh_sessions"
  drop constraint if exists "refresh_sessions_households_refresh_sessions";

alter table "refresh_sessions"
  drop column if exists "active_household_id";

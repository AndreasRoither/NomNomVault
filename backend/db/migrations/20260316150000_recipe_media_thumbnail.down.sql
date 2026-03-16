alter table "media_assets"
  drop constraint if exists "media_assets_stored_objects_thumbnail_media_assets";

alter table "media_assets"
  drop column if exists "thumbnail_storage_object_id";

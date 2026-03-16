alter table "media_assets"
  add column "thumbnail_storage_object_id" character varying;

alter table "media_assets"
  add constraint "media_assets_stored_objects_thumbnail_media_assets"
  foreign key ("thumbnail_storage_object_id") references "stored_objects" ("id") on delete no action;

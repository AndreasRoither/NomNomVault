alter table "recipes"
  add column "status" character varying not null default 'published';

update "recipes"
set "status" = case
  when "archived_at" is not null then 'archived'
  else 'published'
end;

alter table "recipes"
  drop column if exists "archived_at";

create table "source_records" (
  "id" character varying not null,
  "created_at" timestamptz not null,
  "updated_at" timestamptz not null,
  "source_type" character varying not null,
  "import_kind" character varying not null,
  "submitted_url" character varying null,
  "normalized_url" character varying null,
  "canonical_url" character varying null,
  "title_hint" character varying null,
  "content_hash" character varying null,
  "metadata_json" jsonb null,
  "retention_state" character varying not null default 'retained',
  "household_id" character varying not null,
  "raw_snapshot_storage_object_id" character varying null,
  primary key ("id"),
  constraint "source_records_households_source_records"
    foreign key ("household_id") references "households" ("id") on delete no action,
  constraint "source_records_stored_objects_source_records"
    foreign key ("raw_snapshot_storage_object_id") references "stored_objects" ("id") on delete set null
);

create index "sourcerecord_household_id_import_kind_normalized_url" on "source_records" ("household_id", "import_kind", "normalized_url");
create index "sourcerecord_household_id_import_kind_content_hash" on "source_records" ("household_id", "import_kind", "content_hash");

create table "import_jobs" (
  "id" character varying not null,
  "created_at" timestamptz not null,
  "updated_at" timestamptz not null,
  "import_kind" character varying not null,
  "status" character varying not null default 'queued',
  "idempotency_key" character varying null,
  "active_idempotency_key" character varying null,
  "active_fingerprint_key" character varying null,
  "fallback_fingerprint" character varying not null,
  "normalized_payload_json" jsonb null,
  "conflict_state" character varying not null default 'none',
  "warnings_json" jsonb null,
  "confidence_score" double precision null,
  "error_code" character varying null,
  "error_message" character varying null,
  "attempt_count" bigint not null default 1,
  "started_at" timestamptz null,
  "finished_at" timestamptz null,
  "household_id" character varying not null,
  "draft_recipe_id" character varying null,
  "match_recipe_id" character varying null,
  "source_record_id" character varying not null,
  "requested_by_user_id" character varying not null,
  primary key ("id"),
  constraint "import_jobs_households_import_jobs"
    foreign key ("household_id") references "households" ("id") on delete no action,
  constraint "import_jobs_recipes_draft_import_jobs"
    foreign key ("draft_recipe_id") references "recipes" ("id") on delete set null,
  constraint "import_jobs_recipes_matched_import_jobs"
    foreign key ("match_recipe_id") references "recipes" ("id") on delete set null,
  constraint "import_jobs_source_records_import_jobs"
    foreign key ("source_record_id") references "source_records" ("id") on delete cascade,
  constraint "import_jobs_users_requested_import_jobs"
    foreign key ("requested_by_user_id") references "users" ("id") on delete no action
);

create index "importjob_household_id_request_d1fdc51a9411be21844c964f037857c1" on "import_jobs" ("household_id", "requested_by_user_id", "import_kind", "idempotency_key");
create index "importjob_household_id_import_kind_fallback_fingerprint" on "import_jobs" ("household_id", "import_kind", "fallback_fingerprint");
create index "importjob_household_id_status_created_at" on "import_jobs" ("household_id", "status", "created_at");
create index "importjob_source_record_id_created_at" on "import_jobs" ("source_record_id", "created_at");
create unique index "import_jobs_active_idempotency_key_key" on "import_jobs" ("active_idempotency_key");
create unique index "import_jobs_active_fingerprint_key_key" on "import_jobs" ("active_fingerprint_key");

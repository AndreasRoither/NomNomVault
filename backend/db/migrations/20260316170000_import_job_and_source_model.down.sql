drop table if exists "import_jobs";

drop table if exists "source_records";

alter table "recipes"
  add column "archived_at" timestamp with time zone null;

update "recipes"
set "archived_at" = case
  when "status" = 'archived' then now()
  else null
end;

alter table "recipes"
  drop column if exists "status";

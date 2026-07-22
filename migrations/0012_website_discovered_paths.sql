ALTER TABLE websites
ADD COLUMN discovered_paths JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE websites
ADD CONSTRAINT websites_discovered_paths_array CHECK (jsonb_typeof(discovered_paths) = 'array');

ALTER TABLE websites
ADD COLUMN fingerprints JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE websites
ADD CONSTRAINT websites_fingerprints_array CHECK (jsonb_typeof(fingerprints) = 'array');

ALTER TABLE websites DROP CONSTRAINT websites_service_id_key;

CREATE UNIQUE INDEX websites_service_url_idx ON websites (service_id, url);

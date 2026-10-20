UPDATE connectors
SET manifest=jsonb_set(manifest,'{publisher}','"SINet"'::jsonb)
WHERE id='echo' AND version='1.0.0' AND publisher_id IS NULL;

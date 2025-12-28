-- Add event_name column to events table
-- event_name is the specific, detailed event (e.g., "libpulse_upload_to_aws")
-- props.op is the generic operation type (e.g., "upload")

ALTER TABLE public.events
ADD COLUMN event_name TEXT NOT NULL DEFAULT 'unknown';

-- Remove default after adding the column (for future inserts to require it)
ALTER TABLE public.events
ALTER COLUMN event_name DROP DEFAULT;

-- Add index for common queries by event_name
CREATE INDEX idx_events_event_name ON public.events(event_name);

-- Add comment for documentation
COMMENT ON COLUMN public.events.event_name IS 'Specific event name (e.g., libpulse_upload_to_aws).';

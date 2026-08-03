-- destinations.partners is free-text tourism-affiliate data, unrelated to the
-- Partner entity. Rename it now (before the Business migration makes this
-- ambiguous) so ownership-related lookups never confuse the two.
ALTER TABLE destinations RENAME COLUMN partners TO affiliate_partners;

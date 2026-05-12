-- Introduce ChannelStatusClosing (3) between Challenged (2) and Closed (4).
-- Prior to this migration Closed was encoded as 3; shift it to 4 first,
-- then the gap at 3 becomes the new Closing value.
--
-- Status enum mapping after this migration:
--   0 = void
--   1 = open
--   2 = challenged
--   3 = closing  (co-signed Finalize stored off-chain; on-chain close pending)
--   4 = closed

UPDATE channels SET status = 4 WHERE status = 3;

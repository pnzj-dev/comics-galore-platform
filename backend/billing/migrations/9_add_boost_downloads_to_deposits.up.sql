-- Track how many downloads a deposit is purchasing (quota boost) and whether
-- the boost has been granted. 0 / false for ordinary top-up deposits.
ALTER TABLE deposits ADD COLUMN boost_downloads INT NOT NULL DEFAULT 0;
ALTER TABLE deposits ADD COLUMN boost_granted BOOLEAN NOT NULL DEFAULT false;

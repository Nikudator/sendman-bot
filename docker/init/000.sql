CREATE TABLE "botusers" (
    id       serial PRIMARY KEY,
    tid      bigint, -- telegram ID
    uname    varchar(64) NOT NULL DEFAULT '', -- @username
    lname    varchar(64) NOT NULL DEFAULT '', -- last name
    fname    varchar(64) NOT NULL DEFAULT '', -- first name
    rdate    timestamp DEFAULT CURRENT_TIMESTAMP,   -- registration date
    sender   boolean DEFAULT false,    -- can send message to user
    reciver  boolean DEFAULT false,     -- can recive message from user
    uadmin   boolean DEFAULT false      -- user is admin
);
CREATE UNIQUE INDEX tid ON "botusers" (tid);
CREATE INDEX uname ON "botusers" (uname);
CREATE INDEX lname ON "botusers" (lname);
CREATE INDEX fname ON "botusers" (fname);
CREATE INDEX rdate ON "botusers" (rdate);
CREATE INDEX sender ON "botusers" (sender);
CREATE INDEX reciver ON "botusers" (reciver);
CREATE INDEX uadmin ON "botusers" (uadmin);
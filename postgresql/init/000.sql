CREATE TABLE "botusers" (
    id       serial PRIMARY KEY,
    tid      bigint, -- telegram ID
    uname    varchar(64) NOT NULL, -- @username
    lname    varchar(64) NOT NULL, -- last name
    fname    varchar(64) NOT NULL, -- first name
    rdate    timestamp,   -- registration date
    sender   boolean,     -- can send message to user
    reciver  boolean,     -- can recive message from user
    uadmin   boolean      -- user is admin
);
CREATE UNIQUE INDEX tid ON "botusers" (tid);
CREATE INDEX uname ON "botusers" (uname);
CREATE INDEX lname ON "botusers" (lname);
CREATE INDEX fname ON "botusers" (fname);
CREATE INDEX rdate ON "botusers" (rdate);
CREATE INDEX sender ON "botusers" (sender);
CREATE INDEX reciver ON "botusers" (reciver);
CREATE INDEX uadmin ON "botusers" (uadmin);
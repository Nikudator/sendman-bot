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

CREATE TABLE "petitons" (
    id       serial PRIMARY KEY,
    link    varchar(6) NOT NULL DEFAULT '', -- 6 character for link
    vote   boolean DEFAULT false,    -- for group petition (yes/no)
    title text  NOT NULL DEFAULT '', -- title for link
    added_u_id bigint NOT NULL DEFAULT  0, -- who added
    add_date timestamp DEFAULT CURRENT_TIMESTAMP,   -- add date
    deleted_u_id bigint NOT NULL DEFAULT 0, --who deleted
    delete_date timestamp -- delete date
);
CREATE UNIQUE INDEX link ON "petitons" (link);
CREATE INDEX added_u_id ON "petitons" (added_u_id);
CREATE INDEX add_date ON "petitons" (add_date);
CREATE INDEX deleted_u_id ON "petitons" (deleted_u_id);
CREATE INDEX delete_date ON "petitons" (delete_date);

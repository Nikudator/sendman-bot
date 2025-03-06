CREATE TABLE user (
    id       int,
    tid      bigint, -- telegram ID
    uname    varchar(64), -- @username
    lname    varchar(64), -- last name
    fname    varchar(64), -- first name
    rdate    timestamp,   -- registration date
    sender   boolean,     -- can send message to user
    reciver  boolean,     -- can recive message from user
    uadmin   boolean      -- user is admin
);
CREATE UNIQUE INDEX id ON user (id);
CREATE UNIQUE INDEX tid ON user (tid);
CREATE INDEX uname ON user (uname);
CREATE INDEX lname ON user (lname);
CREATE INDEX fname ON user (fname);
CREATE INDEX rdate ON user (rdate);
CREATE INDEX sender ON user (sender);
CREATE INDEX reciver ON user (reciver);
CREATE INDEX uadmin ON user (uadmin);
CREATE TABLE ping (
    device_id UUID NOT NULL,
    ip INET NOT NULL,
    sent_time TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    rtt INTEGER,
    PRIMARY KEY (device_id, sent_time),
    FOREIGN KEY (device_id) REFERENCES device(id) ON DELETE CASCADE
);

CREATE INDEX ping_ip ON ping (ip);

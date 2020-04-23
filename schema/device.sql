CREATE TABLE device_type (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR NOT NULL UNIQUE CHECK (0 < char_length(name) AND char_length(name) < 256)
);

INSERT INTO device_type (name) VALUES ('Server'), ('Switch');

CREATE TABLE device (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_type_id UUID NOT NULL,
    hostname VARCHAR NOT NULL UNIQUE CHECK (0 < char_length(hostname) AND char_length(hostname) < 256),
    FOREIGN KEY (device_type_id) REFERENCES device_type(id)
);

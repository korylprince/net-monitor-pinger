CREATE TABLE ping_aggregate_template (
    device_id UUID NOT NULL,
    ip INET NOT NULL,
    total BIGINT NOT NULL,
    lost BIGINT NOT NULL,
    loss_pct NUMERIC(5, 2) NOT NULL,
    max NUMERIC(6, 2) NOT NULL,
    min NUMERIC(6, 2) NOT NULL,
    avg NUMERIC(6, 2) NOT NULL,
    stddev NUMERIC(6, 2) NOT NULL,
    FOREIGN KEY (device_id) REFERENCES device(id)
);

CREATE FUNCTION ping_aggregate_over(duration INTERVAL)
RETURNS SETOF ping_aggregate_template AS $$
    SELECT
        device_id,
        ip,
        COUNT(*) AS total,
        SUM(CASE WHEN rtt IS NULL THEN 1 ELSE 0 END) AS lost,
        CAST(SUM(CASE WHEN rtt IS NULL THEN 1 ELSE 0 END) * 100 / CAST(COUNT(*) AS NUMERIC(5, 2)) AS NUMERIC(5, 2)) AS loss_pct,
        CAST(MAX(rtt) AS NUMERIC(6, 2)) AS max,
        CAST(MIN(rtt) AS NUMERIC(6, 2)) AS min,
        CAST(AVG(rtt) AS NUMERIC(6, 2)) AS avg,
        CAST(STDDEV(rtt) AS NUMERIC(6, 2)) AS stddev
    FROM ping
    WHERE sent_time > (NOW() AT TIME ZONE 'UTC') - duration
    GROUP BY device_id, ip;
$$ LANGUAGE sql STABLE;

CREATE VIEW ip_status AS
SELECT
    ping.device_id,
	ping.ip,
	ping.sent_time,
	ping.rtt
FROM ping INNER JOIN (
	SELECT
		device_id,
		ip,
		MAX(sent_time) AS sent_time
	FROM ping
	GROUP BY device_id, ip
) AS pings ON
	pings.device_id = ping.device_id AND
	pings.ip = ping.ip AND
	pings.sent_time = ping.sent_time;

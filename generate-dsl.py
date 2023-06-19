#!/usr/bin/env python3

raw_dsl_template = """set 'pipeline.name' = '{dir_name}';

CREATE CATALOG myhive WITH (
    'type' = 'hive',
    'default-database' = 'kodofs2',
    'hive-conf-dir' = '/home/qiniu/conf/hive'
);
-- set the HiveCatalog as the current catalog of the session
USE CATALOG myhive;

SET 'table.sql-dialect' = 'hive';

CREATE EXTERNAL TABLE IF NOT EXISTS dwd_{kafka_topic}(
  {sink_fields}
) PARTITIONED BY (`day` STRING, `hour` STRING) STORED AS parquet TBLPROPERTIES  (
  'parquet.compression' = 'SNAPPY',
  'auto-compaction' = 'true',
  'sink.rolling-policy.file-size' = '128MB',
  'sink.rolling-policy.rollover-interval' = '30 min',
  'sink.rolling-policy.check-interval' = '1 min',
  'sink.partition-commit.delay' = '1 min',
  'sink.partition-commit.policy.kind' = 'success-file,metastore',
  'sink.partition-commit.success-file.name' = '_SUCCESS'
);

SET 'table.sql-dialect' = 'default';

CREATE TEMPORARY FUNCTION ipInfo AS 'com.qiniu.defy.flink.sqlsubmit.udf.IpInfo';
CREATE TEMPORARY FUNCTION formatClientRegion AS 'com.qiniu.defy.flink.sqlsubmit.udf.FormatClientRegion';
CREATE TEMPORARY FUNCTION formatClientIsp AS 'com.qiniu.defy.flink.sqlsubmit.udf.FormatClientIsp';

CREATE TEMPORARY TABLE {kafka_topic}_source(
	headers ROW<
        `topic` VARCHAR,
	    `timestamp` BIGINT
	>,
    `body` ARRAY<
        ROW<
          {source_fields}
        >
    >
) WITH (
    'connector' = 'kafka',
    'topic' = '{kafka_topic}',
    'properties.bootstrap.servers' = 'jjh645:19092,jjh646:19092,jjh655:19092',
    'properties.group.id' = '{dir_name}',
    'scan.startup.mode' = 'latest-offset',
    'format' = 'jsonx',
    'jsonx.parse-as-flume-event' = 'true',
    'jsonx.ignore-parse-errors' = 'true'
);

CREATE TEMPORARY VIEW {kafka_topic}_view1 AS
SELECT
    {view1_keys}
FROM  netrd_qos_qiniupageloadlog_source
CROSS JOIN UNNEST(`body`) AS t({source_keys});

CREATE TEMPORARY VIEW {kafka_topic}_view2 AS
SELECT
    {view2_keys}    
    FROM_UNIXTIME(`ts`/1000, 'yyyyMMdd') as `day`,
    FROM_UNIXTIME(`ts`/1000, 'HH') as `hour`
FROM {kafka_topic}_view1;


INSERT INTO dwd_{kafka_topic}
SELECT
   *
FROM {kafka_topic}_view2;
"""

BIGINT = 'BIGINT'
FLOAT = 'FLOAT'
VARCHAR = 'VARCHAR'
STRING = 'STRING'

type_alias_map = {
    STRING: VARCHAR,
    VARCHAR:VARCHAR,
    BIGINT:BIGINT,
    FLOAT:FLOAT
}

def format_field(kvalues):
    field_descs = []
    for key in kvalues:
        field_descs.append('  `{}` {}'.format(key, kvalues[key]))
    return field_descs

ipinfo_suffix_map = {
    '_country': 'country',
    '_province': 'region',
    '_city': 'city',
    '_isp': 'isp'
}

schema = [
    {
        "name": "ts",
        "type": BIGINT
    },
    {
        "name": "url",
        "type": STRING
    },
    {
        "name": "tag",
        "type": STRING
    },
    {
        "name": "t_page_load",
        "type": FLOAT
    },
    {
        "name": "t_res_load",
        "type": FLOAT
    },
    {
        "name": "t_full_load",
        "type": FLOAT
    },
    {
        "name": "r_id",
        "type": STRING
    },
    {
        "name": "t_window_load",
        "type": FLOAT
    },
    {
        "name": "_loc",
        "type": STRING
    },
    {
        "name": "_loc_country",
        "type":STRING,
        "extended": True
    },
    {
        "name": "_loc_province",
        "type": STRING,
        "extended": True
    },
    {
        "name": "_loc_city",
        "type": STRING,
        "extended": True
    },
    {
        "name": "_loc_isp",
        "type": STRING,
        "extended": True
    },
    {
        "name": "os",
        "type": STRING
    },
    {
        "name": "app",
        "type": STRING
    },
    {
        "name": "dev_model",
        "type": STRING
    },
    {
        "name": "dev_id",
        "type": STRING
    }
]


def main():
    sink_fields = {}
    source_fields = {}
    fields_extended = {}

    for field in schema:
        sink_fields[field["name"]] = field["type"]
        
        extended = field.get("extended")
        if extended is not None:
            fields_extended[field["name"]] = extended
        
        if extended is None or not extended:
            source_fields[field["name"]] = type_alias_map[field["type"]]

    source_keys = []
    for field in schema:
        extended = field.get("extended")
        if extended is None or not extended:
            source_keys.append("`{}`".format(field["name"]))

    view1_keys = []
    for key in source_keys:
        view1_keys.append('  {}'.format(key))
    view1_keys.append('    ipInfo(_loc) as ipInfo')

    view2_keys = []
    for key in sink_fields:
        if fields_extended.get(key) is not None:
            extended = fields_extended[key]
            
            if key.rindex('_') != -1:
                rid = key.rindex('_')
                key_ext = key[rid:]
                
                if key_ext.find('country') != -1:
                    view2_keys.append("    ipInfo.country as `{}`".format(key))
                elif key_ext.find('city') != -1:
                    view2_keys.append("    '-' as `{}`".format(key))
                elif key_ext.find('province') != -1:
                    view2_keys.append("    formatClientRegion(ipInfo.region,'-') as `{}`".format(key))
                elif key_ext.find('isp') != -1:
                    view2_keys.append("    formatClientIsp(ipInfo.isp,'-') as `{}`".format(key))            
        
        else:
            view2_keys.append("    `{}`".format(key))

    format_view2_keys = ',\n'.join(view2_keys)

    if len(view2_keys) > 0:
        format_view2_keys += ','

    template_fields = {
        'kafka_topic': 'netrd_qos_qiniupageloadlog',
        'dir_name': """{{ .dir }}""",
        'sink_fields': ',\n'.join(format_field(sink_fields)),
        'source_fields': ',\n'.join(format_field(source_fields)),
        'source_keys': ','.join(source_keys),
        'view1_keys': ',\n'.join(view1_keys),
        'view2_keys':  format_view2_keys
    }

    with open("/tmp/job.dsl", 'w') as f:
        f.write(raw_dsl_template.format(**template_fields))

if __name__ == "__main__":
    main()

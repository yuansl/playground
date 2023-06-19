mysqldump --databases traffic --tables raw_day_traffic_2021_04 --where="domain in('line1.qngslb.com')" -h xs2198 -P 3359 -u traffic_admin -p --skip-opt >cloudflare.fusion.traffic.dump.202104.sql


delete from raw_day_traffic_2021_04 where domain='line1.qngslb.com';



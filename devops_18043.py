#!/usr/bin/python3

import pymysql
import datetime
import sys


def list_dcdn_domains(db, start, end):
    domains = []

    try:
        with db.cursor() as cur:
            cur.execute('SELECT distinct(domain) from raw_day_traffic_{} where cdn=23 and data_type=1 and source_type=2'.format(start.strftime("%Y_%m")))

            rows = cur.fetchall()
            for row in rows:
                domains.append(row[0])

    except Exception as e:
        print(e)
        return None
    finally:
        db.close()
    return domains


def fetch_dup_domains(db, start, end):
    try:
        with db.cursor() as cur:
            day = start
            while day < end:
                domains = []

                cur.execute('SELECT distinct(domain) from raw_day_traffic_{} where day={} and cdn=23 and data_type=0 and source_type=2'.format(day.strftime("%Y_%m"), int(day.strftime("%Y%m%d"))))
                rows = cur.fetchall()
                for row in rows:
                    domains.append(row[0])

                print("""delete from raw_day_traffic_{} where day={} and cdn=17 and data_type=0 and domain in({});\n""".format(day.strftime("%Y_%m"), int(day.strftime("%Y%m%d")), ','.join(["'{}'".format(domain) for domain in domains])))
                
                day = day + datetime.timedelta(days=1)
            
    except Exception as e:
        print(e)
    finally:
        db.close()


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: {} <start> <end> <mode>".format(sys.argv[0]))
        exit(0)

    start_date = datetime.datetime.strptime(sys.argv[1], "%Y-%m-%d").date()
    end_date = datetime.datetime.strptime(sys.argv[2], "%Y-%m-%d").date()

    try:
        db = pymysql.connect(host='10.34.46.62',
                             port=3359,
                             user='traffic_admin',
                             password='kdK26Ws824Q9ivvfao2ns',
                             database='traffic')
    except Exception as e:
        print('exception: {}'.format(e))

    if sys.argv[3] == "dup":
        fetch_dup_domains(db, start_date, end_date)
    else:
        domains = list_dcdn_domains(db, start_date, end_date)
        if domains is not None:
            print(','.join(domains))

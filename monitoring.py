#!/usr/bin/python3

import datetime
import requests

header = {
    "Authorization": "QBox 557TpseUM8ovpfUhaw8gfa2DQ0104ZScM-BTIcBx:VxxxPblLsXIET0XxChFJcKt1frw="
}
FUSION_TRAFFIC_ENDPOINT = "http://fusion.qiniuapi.com/v2/tune/bandwidth"


def fetch_fusion_traffic(domains, begin, end):
    try:
        data = {
            "domains": domains,
            "granularity": "5min",
            "startDate": begin.strftime('%Y-%m-%d %H:%M'),
            "endDate": end.strftime('%Y-%m-%d %H:%M'),
            "type": "302bandwidth",
        }
        r = requests.post(FUSION_TRAFFIC_ENDPOINT, headers=header, json=data, timeout=30, verify=False)
        x = r.json()
        return x
    except Exception:
        url0 = FUSION_TRAFFIC_ENDPOINT.split("com", 1)[1]
        print('http_302delay_code{{pattern="{}",code="0"}} 1'.format(url0))
        exit()


def inspect_reponse_and_report(payload, url, time, domain, minutes):
    sum = 0
    if "code" in payload and payload["code"] == 200:
        n = len(payload["data"][domain]["china"])
        while n > 0:
            if str(payload["time"][n - 1]) == str(time.strftime("%Y-%m-%d %H:%M:%S")):
                if payload["data"][domain]["china"][n - 1] == 0:
                    print(
                        'blackbox_http_302delay{{pattern="{}",domain="{}",area="china" {}}}'.format(
                            url, domain, minutes))
                    return sum + 1
                else:
                    break
            else:
                n = n - 1
        return sum
    else:
        if "code" in payload:
            print(
                'http_302delay_code{{pattern="{}",code="{}"}} 1'.format(url,payload["code"]))
        else:
            print(
                'http_302delay_code{{pattern="{}",code="0"}} 1'.format(url))

        exit(1)


def get_5min_aligned_time(minutes):
    now = datetime.datetime.today() - datetime.timedelta(minutes=minutes)

    if now.minute % 5 == 0:
        min = now.minute
    else:
        min = now.minute - now.minute % 5

    now = datetime.datetime(
        now.year, now.month, now.day, now.hour, min
    )
    return now


def detect_traffic_api_delay(delay_minutes, domains):
    monitorat = get_5min_aligned_time(delay_minutes)

    for domain in domains:
        payload = fetch_fusion_traffic(domain, monitorat, monitorat)
        inspect_reponse_and_report(payload, "/v2/tune/bandwidth", monitorat, domain, delay_minutes)


def detect_traffic_exception_of_domain(domain, at):
    stats = []
    for t in [at - datetime.timedelta(hours=24), at - datetime.timedelta(hours=48)]:
        payload = fetch_fusion_traffic(domain, t, t)
        stats.append(payload["data"][domain]["china"][0])

    payload = fetch_fusion_traffic(domain, at, at)
    current = payload["data"][domain]["china"][0]
    avgbandwidth = sum(stats)/len(stats)

    try:
        if abs(avgbandwidth-current)/avgbandwidth > 0.4:
            print('http_302_unexpected_traffic{{pattern="/v2/tune/bandwidth",avgbandwidth="{}",current="{}"}} 1'.format(avgbandwidth, current))
    except Exception:
        if avgbandwidth == 0:
            print('http_302_unexpected_traffic{{pattern="/v2/tune/bandwidth",avgbandwidth="0",current="{}"}} 1'.format(current))
            exit(1)

def detect_traffic_exception(domains, delay_minutes):
    at = get_5min_aligned_time(delay_minutes)

    for domain in domains:
        detect_traffic_exception_of_domain(domain, at)

if __name__ == "__main__":
    domains = ["dl.steam.clngaa.com"]
    delay_minutes = 20

    detect_traffic_api_delay(delay_minutes, domains)
    detect_traffic_exception(domains, delay_minutes)

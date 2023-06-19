#!/bin/bash

cdns=aliyun alidcdn alcmcc baidu baishanyun cloudflare dpco dpco2 googlecloudcdn huaweicloud pingan ss tencent upyun yunfan

fscdn_host=fscdn.defy.internal.qiniu.io

if [ -n "$1" ]; then
	date=$1
else
	echo "Usage: $0 date"
	exit 0
fi

FSHOST=xs321
for cdn in pingan ; do
    fsrobot_host=$FSHOST:18214
    
    begin=$(date --date=$date +%Y-%m-%d)
    end=$(date --date="$begin next day" +%Y-%m-%d)

    log_file=time-range-bandwidth.$cdn.$(date --date=$begin +%Y%m%d).log
    echo "$(date): processing cdn $cdn ..." >>$log_file

    declare -a domains=($(curl http://$fscdn_host/v1/cdn/domains?cdn=$cdn | jq . |grep Domain | awk -F: '{print $2}'|tr -d '"' |tr -d '[:blank:]'))
    #declare -a domains=($(./onlinedomains -cdn $cdn -day $begin -hour 1))
    # declare -a domains=(baiducdncmn2.inter.iqiyi.com baiducdncmn2.inter.ptqy.gitv.tv baiducdncmn3.inter.iqiyi.com baiducdncmn3.inter.ptqy.gitv.tv baiducdncnc.inter.iqiyi.com baiducdncnc.inter.ptqy.gitv.tv baiducdnct-gd.inter.iqiyi.com baiducdnct-gd.inter.ptqy.gitv.tv baiducdnct.inter.iqiyi.com baiducdnct.inter.ptqy.gitv.tv baiducdntest.inter.iqiyi.com baiducdntest.inter.ptqy.gitv.tv bdcdn.inter.iqiyi.com bdcdncmn2.inter.71edge.com bdcdncmn2.inter.ptqy.gitv.tv bdcdncmn3.inter.71edge.com bdcdncmn3.inter.ptqy.gitv.tv bdcdncnc.inter.71edge.com bdcdncnc.inter.ptqy.gitv.tv bdcdnct-gd.inter.71edge.com bdcdnct-gd.inter.ptqy.gitv.tv bdcdnct.inter.71edge.com bdcdnct.inter.ptqy.gitv.tv)

    for domain in ${domains[@]} ; do
	echo processing domain $domain ...

	curl -v "http://$fsrobot_host/v1/cdn/time/range/bandwidth?cdn=$cdn&from=$begin&to=$end&domains=$domain" >>$log_file 2>&1
    done
done

#!/bin/bash

declare -A tasks

tasks[202101]=bilibli.domains.202101.csv
tasks[202102]=bilibli.domains.202102.csv
tasks[202103]=bilibli.domains.202103.csv
tasks[202104]=bilibli.domains.202104.csv
tasks[202105]=bilibli.domains.202105.csv
tasks[202106]=bilibli.domains.202106.csv

for key in ${!tasks[*]} ; do
    month=$key

    csvfile=${tasks[$key]}
    domains=
    for domain in $(awk -F, 'NR>1{print $1}' $csvfile); do if [ -z $domains ]; then domains=$domain; else domains=$domain,$domains; fi done

    begin=$(date --date=$month"01" +%Y-%m-%d)
    end=$(date --date="$begin next month" +%Y-%m-%d)
    ./rawbandwidth -cdn aliyun -begin $begin -end $end -excludes $domains -host xs615 -passwd kdK26Ws824Q9ivvfao2ns >aliyun.withoutbilibili.traffic.$month.csv 2>&1 
    
done


iqiyiundomains=baiducdncnc.inter.iqiyi.com,baiducdncnc.inter.ptqy.gitv.tv,baiducdnct-gd.inter.iqiyi.com,baiducdnct-gd.inter.ptqy.gitv.tv,baiducdnct.inter.iqiyi.com,baiducdnct.inter.ptqy.gitv.tv,baiducdntest.inter.iqiyi.com,baiducdntest.inter.ptqy.gitv.tv,bdcdn.inter.iqiyi.com,bdcdncnc.inter.71edge.com,bdcdncnc.inter.ptqy.gitv.tv,bdcdnct-gd.inter.71edge.com,bdcdnct-gd.inter.ptqy.gitv.tv,bdcdnct.inter.71edge.com,bdcdnct.inter.ptqy.gitv.tv


iqiyicmccdomains=baiducdncmn2.inter.iqiyi.com,baiducdncmn2.inter.ptqy.gitv.tv,baiducdncmn3.inter.iqiyi.com,baiducdncmn3.inter.ptqy.gitv.tv,bdcdncmn2.inter.71edge.com,bdcdncmn2.inter.ptqy.gitv.tv,bdcdncmn3.inter.71edge.com,bdcdncmn3.inter.ptqy.gitv.tv,baiducdntest.inter.iqiyi.com,baiducdntest.inter.ptqy.gitv.tv

domains=$iqiyiundomains

for month in 202104 202105 202106  ; do

    begin=$(date --date=$month"01" +%Y-%m-%d)
    end=$(date --date="$begin next month" +%Y-%m-%d)
    
    ./rawbandwidth -cdn baishanyun -begin $begin -end $end -excludes $domains -host xs615 -passwd kdK26Ws824Q9ivvfao2ns > baishanyun.traffic.exclude-iqiyicun.$month.csv -source 1  2>&1 &
done

# 爱奇艺电信联通:
# 2021-06:
# peak95Average: 822.391910 Mbps, peak95Month: 844.855735 Mbps

# 去掉爱奇艺电联白山:
# 2021-06
# peak95Average: 112.588513641 Gbps, peak95Month: 112.753716445 Gbps

# 去掉移动爱奇艺后白山：

# 2021-04:
# peak95Average: 167.173709541 Gbps, peak95Month: 178.989668423 Gbps

# 2021-05:
# peak95Average: 50.210035173 Gbps, peak95Month: 50.270341831 Gbps

# 2021-06:
# peak95Average: 63.730257220 Gbps, peak95Month: 64.918297679 Gbps


# 移动爱奇艺在白山量:
# 2021-04:
# peak95Average: 45.113006069 Gbps, peak95Month: 47.064305951 Gbps

# 2021-05:
# peak95Average: 43.251349770 Gbps, peak95Month: 49.103187223 Gbps

# 2021-06:
# peak95Average: 51.095101329 Gbps, peak95Month: 51.484183835 Gbps


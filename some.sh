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

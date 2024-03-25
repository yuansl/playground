FROM ubuntu:latest
RUN sed -i 's/archive.ubuntu.com/mirrors.ustc.edu.cn/g' /etc/apt/sources.list && apt update && apt install -y curl
COPY ./myraft /usr/local/bin/myraft
WORKDIR /var/local/lib/myraft
EXPOSE 1234

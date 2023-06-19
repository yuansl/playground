FROM ubuntu:22.04
COPY gomemlimit /usr/local/bin
WORKDIR /var/lib/gomemlimit
EXPOSE 9090
CMD ["gomemlimit", "-n", "1000_000_00"]

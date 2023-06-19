#!/bin/bash

# generate CA's private key & CA's self-signed certificate
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=CN/ST=Shanghai/L=Shanghai/O=yuansl.io/OU=Technology/CN=*.yuansl.io/emailAddress=yuanshenglong@126.com"

# generate server's private key & server's certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=CN/ST=Shanghai/L=Shanghai/O=yuansl.io/OU=IT/CN=*.yuansl.io/emailAddress=yuanshenglong@126.com" 

# signing server's certificate request by CA's private key
openssl x509 -req -in server-req.pem -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf 

# verify the signed server certificate
openssl x509 -in server-cert.pem  -noout -text

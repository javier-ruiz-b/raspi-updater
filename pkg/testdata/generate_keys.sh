#!/bin/bash
set -eux

hostname="$1"

echo "Generating CA key and certificate:"
openssl req -x509 -sha256 -nodes -days 3650 -newkey rsa:4096 \
  -keyout ca.key -out ca.crt \
  -subj '/O=quic-go Certificate Authority/'

echo "Generating CSR"
openssl req -out "$hostname".csr -new -newkey rsa:4096 -nodes -keyout "$hostname".key \
  -subj '/O=quic-go/'

echo "Sign certificate:"
openssl x509 -req -sha256 -days 3650 -in "$hostname".csr  -out "$hostname".crt \
  -CA ca.crt -CAkey ca.key -CAcreateserial \
  -extfile <(printf "subjectAltName=DNS:%s" "$hostname")

# debug output the certificate
openssl x509 -noout -text -in "$hostname".crt

# we don't need the CA key, the serial number and the CSR anymore
rm ca.key "$hostname".csr ca.srl


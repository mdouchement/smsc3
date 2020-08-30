#!/bin/bash

# Parameter $1 is the smsc (cm / sinch)
dlrurl="http%3A%2F%2Fhttpcat%3A8888%2Fsms-dlr%3Fdlr%3D%25d%26smsid%3D%25I%26sender%3D%25p%26receiver%3D%25P%26time%3D%25t%26text%3D%25b%26smscid%3D%25i%26charset%3D%25C%26coding%3D%25c%26meta%3D%25D"

set -o xtrace
curl -i "http://localhost:13013/cgi-bin/sendsms?username=foo&password=bar&from=GOPHER&to=0033600000001&charset=UTF-8&text=Hello+world&dlr-mask=31&dlr-url=${dlrurl}&smsc=$1"
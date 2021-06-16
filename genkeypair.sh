#!/bin/bash
openssl req -x509 -nodes -newkey rsa:4096 -keyout server.key -out server.crt -days 365
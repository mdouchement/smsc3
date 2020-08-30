#!/bin/bash

GOOS=linux GOARCH=amd64 go build && docker-compose kill && docker-compose up
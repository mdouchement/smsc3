# Fake kannel

This directory contains all configuration files to create a fake kannel server.
Base code come from: https://github.com/onlinecity/kannel-docker

## How to use my own kannel config ?

If you don't want to use the templating system to generare config from env but your own config file:
- Mount you config file on `/etc/kannel/kannel.conf`

#!/bin/bash
set -e

wget https://www.antlr.org/download/antlr-${ANTLR_VERSION}-complete.jar -O /tmp/antlr.jar
sudo mv /tmp/antlr.jar /usr/local/lib/

#!/bin/bash
set -e
base=`dirname "$0"`"/.."

echo building server...
cd $base
go build -v

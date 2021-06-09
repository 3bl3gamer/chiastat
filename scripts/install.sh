#!/bin/bash
set -e

repo_dir=`dirname "$0"`"/.."
repo_dir=`realpath "$repo_dir"`

function usage {
    echo "usage: $0 [--migrate] [--restart]"
}
if [ "$1" == "-h" ] || [ "$1" == "--help" ]; then usage; exit 2; fi

migrate=false
restart=false
for i in "$@"; do
    case $i in
        --migrate) migrate=true;;
        --restart) restart=true;;
    esac
done

if [ $migrate = true ]; then
    go run "$repo_dir"/migrations/*.go
fi

if [ $restart = true ]; then
    for name in chia chiastat-update; do
        systemctl restart $name
    done
fi
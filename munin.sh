#!/bin/sh

# copy to /etc/munin/plugins/phone

SWITCHD_HOST=${SWITCHD_HOST:-localhost:8000}

output_config() {
    echo "multigraph battery_charge"
    echo "graph_title Battery Charge"
    echo "graph_args -l 15 -u 81"
    echo "graph_vlabel level (%)"
    echo "batt.label Charge"
    echo "\nmultigraph data_age"
    echo "graph_title Data Age"
    echo "graph_vlabel time (sec)"
    echo "graph_args -l 0 -u 43200 -r"
    echo "age.label Age"
}

output_values() {
    echo "multigraph battery_charge"
    echo "batt.value $(charge_level)"
    echo "\nmultigraph data_age"
    echo "age.value $(data_age)"
}

charge_level() {
    curl -qs "$SWITCHD_HOST/dbg/level"
}

data_age(){
    curl -qs "$SWITCHD_HOST/dbg/age"
}

output_usage() {
    printf >&2 "%s - munin plugin for switchd\n" ${0##*/}
    printf >&2 "Usage: %s [config]\n" ${0##*/}
}

case $# in
    0)
        output_values
        ;;
    1)
        case $1 in
            config)
                output_config
                ;;
            *)
                output_usage
                exit 1
                ;;
        esac
        ;;
    *)
        output_usage
        exit 1
        ;;
esac

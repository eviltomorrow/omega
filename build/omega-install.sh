#!/bin/bash

function output() {
    echo -e "{\"omega-watchdog\":\"${1}\", \"omega\":\"${2}\"}"
}

omega_home=$(pwd)/omega

while getopts i:o:e:g: OPT; do
    case ${OPT} in
        i) inner_ip=${OPTARG}
        ;;
        o) outer_ip=${OPTARG}
        ;;
        e) endpoints=${OPTARG}
        ;;
        g) group=${OPTARG}
        ;;
        \?)
           echo "unknown [${OPT}], check input parameter"
           exit 0
    esac
done

if [ -z "${inner_ip}" -o -z "${endpoints}" -o -z "${group}" ]; then
    echo "[inner_ip:${inner_ip}, endpoints: ${endpoints}, group: ${group}], check input parameter"
    exit 0
fi


function prepareEnv() {
    if [ ! -n "${inner_ip}" ]; then
        echo "inner_ip parameter is missing"
        exit 0
    fi
    if [ ! -f "omega.tar.gz" ]; then
        echo "image file omega.tar.gz not exist"
        exit 0
    fi

    if [ -d ${omega_home} ]; then
        if [ ! -d "omega/bin" ]; then
            echo "omega/bin not exist[systemd], please change install dir"
            exit 0
        fi
        if [ ! -e "omega/bin/omega-watchdog" ]; then
            echo "omega/bin/omega-watchdog not exist, please change install dir"
            exit 0
        fi
        if [ ! -d "omega/etc" ]; then
            echo "omega/etc not exist[systemd], please change install dir"
            exit 0
        fi
        if [ ! -e "omega/etc/omega.conf" ]; then
            echo "omega/etc/omega.conf not exist[systemd], please change install dir"
            exit 0
        fi
        if [ -e "omega/var/run/omega-watchdog.pid" ]; then
            kill -3 $(cat omega/var/run/omega-watchdog.pid)
        fi
        if [ -e "omega/var/run/omega.pid" ]; then
            kill -3  $(cat omega/var/run/omega.pid)
        fi

        rm -rf omega/bin
        rm -rf omega/var
    fi 
}

function doInstall() {
    tar -zxvf omega.tar.gz > /dev/null
    check0 "tar -zxvf omega.tar.gz"

    chown -R ${USER}:${GROUPS} ${omega_home}
    cd ${omega_home}
    check0 "cd ${omega_home}"

    chmod a+x bin/omega-watchdog
    check0 "chmod a+x bin/omega-watchdog"
}

function genConfig() {
    cd ${omega_home}
    check0 "cd ${omega_home}"

    attrs="--inner_ip ${inner_ip} --endpoints ${endpoints} --group ${group}"
    if [ -n ${outer_ip} ]; then
        attrs="${attrs} --outer_ip ${outer_ip}"
    else
        attrs="${attrs} --outer_ip ${inner_ip}"
    fi

    if [ ! -e ../etc/omega.conf ]; then
        touch etc/omega.conf
    fi
    ./bin/omega-watchdog config ${attrs} > etc/omega.conf
    check0 "./bin/omega-watchdog config ${attrs} > /etc/omega.conf"
}

function startupOmega() {
    cd ${omega_home}/bin
    ./omega-watchdog --daemon
    check0 " ./omega-watchdog --daemon"

    while :
    do
        if [ -e ../log/error.log ]; then
            errlog=$(tail -n 1 ../log/error.log)
            echo ${errlog}
            exit 0
        fi
        if [ -e ../var/run/omega-watchdog.pid ]; then
            if [ -e ../var/run/omega.pid ]; then
                omega_watchdog_pid=$(cat ../var/run/omega-watchdog.pid)
                omega_pid=$(cat ../var/run/omega.pid)
                output ${omega_watchdog_pid} ${omega_pid}
                exit 0  
            fi
        fi
        sleep 1
    done
}

function check0() {
    if [ $? -ne 0 ]; then
        echo $1
        exit 0
    fi
}


prepareEnv
doInstall
genConfig
startupOmega

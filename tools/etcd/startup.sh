#!/bin/bash

./etcd --name 'default' \
--data-dir 'data/' \
--listen-client-urls 'http://192.168.12.140:2379,http://127.0.0.1:2379' \
--advertise-client-urls 'http://192.168.12.140:2379,http://127.0.0.1:2379' \
--listen-peer-urls 'http://192.168.12.140:2380,http://127.0.0.1:2380' \
--initial-advertise-peer-urls 'http://192.168.12.140:2380,http://127.0.0.1:2380' \
# --initial-cluster default=http://0.0.0.0:2380

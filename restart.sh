#!/bin/bash
# 清空数据并重启

killall -9 ssbcV2

./ssbcV2 clear

sh ./start.sh

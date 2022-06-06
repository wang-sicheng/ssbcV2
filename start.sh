#!/bin/bash
# 启动所有节点

count=`ps -ef | grep ssbcV2 | grep -v "grep" | wc -l`
if [ $count -gt 0 ]; then
  echo "ssbcV2 service is running. Please stop it first."
  exit
fi

if [ ! -d "log" ]; then
  mkdir log
fi
echo "Waiting..."

./ssbcV2 N0 > log/N0.log 2>&1 &
./ssbcV2 N4 > log/N4.log 2>&1 &
sleep 2

./ssbcV2 N1 > log/N1.log 2>&1 &
./ssbcV2 N2 > log/N2.log 2>&1 &
./ssbcV2 N3 > log/N3.log 2>&1 &
./ssbcV2 client1 > log/client1.log 2>&1 &

./ssbcV2 N5 > log/N5.log 2>&1 &
./ssbcV2 N6 > log/N6.log 2>&1 &
./ssbcV2 N7 > log/N7.log 2>&1 &
./ssbcV2 client2 > log/client2.log 2>&1 &

sleep 4
echo "You've started ssbcV2 service."

@echo off
set /p nbSlave=Number of slaves: 

echo "running master"
start cmd /k go run master/master.go

echo "running slave"
FOR /L %%x IN (1, 1, %nbSlave%) DO start cmd /k go run slave/slave.go
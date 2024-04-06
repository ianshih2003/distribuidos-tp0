#!/bin/bash


res=$(echo $MSG | nc -w $TIMEOUT $SERVER_IP $PORT)

if [ "$MSG" = "$res" ]
then
    echo "Servidor respondio correctamente"
    return 0
else
    echo "Servidor respondio incorrectamente"
    return 1
fi
#!/bin/bash

MSG="Hola mundo"
SERVER_IP=server
TIMEOUT=3
PORT=12345

res=$(nc -w $TIMEOUT $SERVER_IP $PORT <<< $MSG)

if [ "$MSG" = "$res" ]
then
    echo "Servidor respondio correctamente"
else
    echo "Servidor respondio incorrectamente"
fi
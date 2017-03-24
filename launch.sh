#!/bin/bash

# a script to launch a go project after a tar has been transfered to the server
# it will assume somethings like the name and location of the tar
# and the project name in supervisor
# this should only be run on the server where the project lives
# It will stop supervisor, remove all files that are contained in the new .tar
# then restart supervisor

clear

DIR=`echo ${PWD##*/}`

echo "Stopping ${DIR} with supervisor..."
sudo supervisorctl stop ${DIR}

echo "Removing old files..."
sudo rm -rf static/ templates/ fonts/ ${DIR} ${DIR}.tar

echo "Moving new tar to current location..."
sudo mv ~/${DIR}.tar .
if [ ! -f $DIR.tar ]; then
    echo "Move tar $DIR.tar failed."
    exit 1
fi

echo "Extracting contents of new tar..."
sudo tar xf ${DIR}.tar
if [ ! -f $DIR ]; then
	echo "No binary found"
	exit 1
fi

echo "Starting ${DIR} with supervisor..."
sudo supervisorctl start ${DIR}

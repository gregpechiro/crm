#!/bin/bash

clear
NODE="$1"
DIR=`echo ${PWD##*/}`
#EXCLUDE=(".git" ".gitignore" "db" "upload" "deploy.sh" "*.go" "launch.sh" ".tern-project" "README.md" ".floo" ".flooignore")
INCLUDE=("${DIR} static/ templates/")

if [ "$NODE" == "" ]; then
    echo "No node specified!"
    exit 1
fi

if [ -f "${DIR}.tar" ]; then
    echo "Removing old tar ${DIR}.tar..."
    rm $DIR.tar
fi

echo "Removing old binary ${DIR}..."
go clean
echo "Building ${DIR}..."
go build

if [ ! -f $DIR ]; then
    echo "Build $DIR failed."
    exit 1
fi

echo "Creating tar ${DIR}.tar..."

#for item in ${EXCLUDE[*]}; do
#    TOGETHER="$TOGETHER --exclude $item"
#done

tar cf $DIR.tar $INCLUDE
if [ ! -f "$DIR.tar" ]; then
    echo "Create $DIR.tar failed."
    exit 1
fi

echo "SCP to node${NODE}..."
scp $DIR.tar node$NODE@node$NODE.cagnosolutions.com:/home/node$NODE

#!/bin/bash


go build -o server main.go

export KUBE_CONFIG="$HOME/.kube/config"
export LOGLEVEL="debug"
export INDEX_HTML="gui/index.html"

./server

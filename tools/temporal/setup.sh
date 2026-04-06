#!/bin/bash

until temporal operator cluster health | grep -q SERVING; do
    echo "Waiting for Temporal server to start..."
    sleep 1
done
echo "create default namespace"
temporal operator namespace create --namespace default || echo "already exists"

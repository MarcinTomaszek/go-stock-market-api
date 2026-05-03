#!/bin/bash

if [ -z "$1" ]; then
  echo "Please enter the port, e.g.: ./start.sh 8080"
  exit 1
fi

export APP_PORT=$1
docker-compose up --build -d
echo "The application was launched on http://localhost:$APP_PORT"
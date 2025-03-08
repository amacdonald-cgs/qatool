#!/bin/bash
cd backend_api
go build -o backend_api
DEV_MODE=true ./backend_api

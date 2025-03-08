#!/bin/bash
cd backend_api
go build -o backend_api
JWT_SECRET="1234567890123456789012345678901234567890" DEV_MODE=true ./backend_api

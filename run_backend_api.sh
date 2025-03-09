#!/bin/bash
cd backend_api
go build -o backend_api
DATABASE_URL="postgresql://qa:password@localhost:5433/qa?schema=public" JWT_SECRET="1234567890123456789012345678901234567890" DEV_MODE=true ./backend_api

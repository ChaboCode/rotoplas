# Rotoplas
catbox.moe inspired upload server

This is an educational project for a file upload server. It runs on the
Gin Framework, and does nothing more than receive files from an HTML form
and registering them into a MongoDB database. It also displays the current
stored files in the main page.

## Pending features
- Display thumbnail instead of whole file
- Home page listing pagination
- Login features
- ~~Better styling~~
- MinIO storage?

## How to run
1. `go get ./...` > Install dependencies
2. `docker-compose up` > Build Docker image and deploy both Go server and MongoDB instance
3. `localhost:8080` > Open the uploader
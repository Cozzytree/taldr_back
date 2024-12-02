build:
	@echo "Building..."


	@go build -o main cmd/main.go

# Run the application
run:
   # Run the MongoDB container with a volume for data persistence
	# sudo docker run -d --name taldr -p 27017:27017 \
	# 	-e MONGO_INITDB_ROOT_USERNAME=root -e MONGO_INITDB_ROOT_PASSWORD=secret \
	# 	-v /home/cozzytree/Documents/taldr_db/db:/data/db \
	# 	mongo
	sudo docker start taldr

	# Run the Go application
	@go run cmd/main.go


.PHONY: build run

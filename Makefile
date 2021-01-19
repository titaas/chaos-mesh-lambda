build:
	GOOS=linux go build main.go
	zip function.zip main

clean:
	rm -f main function.zip

upload:
	@echo "You can using aws client to upload lambda function here"
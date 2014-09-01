run:
	PORT=8080 ./go-reload *.go

push:
	git push origin
	git push heroku master

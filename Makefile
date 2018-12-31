run: js
	go run main.go

js:
	gopherjs build -o static/index.js views/index/index.go

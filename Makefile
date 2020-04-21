build:
	go build ./...

run-tests:
	cd tests && go test -o ../tests.test -c -covermode=atomic -coverprofile=c.out -coverpkg=../...
	./tests.test -test.v=true -test.coverprofile=c.out
	sed -i -e '\:^go-ml.dev/pkg/iokit/:d' c.out
	sed -i -e '\:^go-ml.dev/pkg/zorros/:d' c.out
	cp c.out gocov.txt
	sed -i -e 's:go-ml.dev/pkg/base/::g' c.out

run-cover:
	go tool cover -html=gocov.txt

run-cover-tests: run-tests run-cover


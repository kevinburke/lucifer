install:
	go install ./...

watch:
	justrun -c 'make install' 'lucifer'

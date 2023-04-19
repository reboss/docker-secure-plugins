

.container: Dockerfile main.go plugin.go config.json
		docker build -t docker-shield .
		touch .container

rootfs: .container
		mkdir rootfs
		$(eval ID := $(shell docker run --rm -d docker-shield))
		docker export $(ID) | tar -x -C $@
		docker kill $(ID)

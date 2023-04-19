.PHONY: plugin

clean:
		rm -rf .container rootfs
		docker plugin disable docker-shield || true
		docker plugin rm docker-shield || true

.container: Dockerfile main.go plugin.go config.json
		docker build -t docker-shield .
		touch .container

rootfs: .container
		rm -rf $@
		mkdir $@
		$(eval ID := $(shell docker run --rm -d docker-shield))
		docker export $(ID) | tar -x -C rootfs
		docker kill $(ID)

plugin: rootfs
		docker plugin disable docker-shield || echo "Already disabled"
		docker plugin rm docker-shield || echo "Already removed"
		docker plugin create docker-shield .
		docker plugin enable docker-shield

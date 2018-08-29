.PHONY: all
all: install_task_runner run_tasks

.PHONY: install_task_runner
install_task_runner:
	@echo "==> Installing task runner ${PROJECT_NAME}..."
	go get github.com/go-task/task/cmd/task

.PHONY: run_tasks
run_tasks:
	@echo "==> Running tasks..."
	task build dist


.PHONY: build_docker_image
build_docker_image:
	@echo "==> Building docker image..."
	docker build -t hetzner-server-market-exporter .

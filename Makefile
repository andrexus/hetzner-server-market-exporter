.PHONY: all
all: task_runner_linux run_tasks


task_runner_mac: download_task_runner_mac install_task_runner
task_runner_linux: download_task_runner_linux install_task_runner

.PHONY: download_task_runner_mac
download_task_runner_mac:
	@echo "==> Downloading task runner (mac)..."
	wget -q -O task.tar.gz https://github.com/go-task/task/releases/download/v2.1.1/task_darwin_amd64.tar.gz

.PHONY: download_task_runner_linux
download_task_runner_linux:
	@echo "==> Downloading task runner (linux)..."
	wget -q -O task.tar.gz  https://github.com/go-task/task/releases/download/v2.1.1/task_linux_386.tar.gz

.PHONY: install_task_runner
install_task_runner:
	@echo "==> Installing task runner..."
	tar xzf task.tar.gz
	mv task ${GOPATH}/bin
	rm -f task.tar.gz

.PHONY: run_tasks
run_tasks:
	@echo "==> Running tasks..."
	task setup_build test build dist

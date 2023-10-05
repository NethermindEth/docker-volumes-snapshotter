.PHONY: format

format: ## Run gofumpt against code to format it
	@gofumpt -l -w .

.PHONY: release-dry release

release-dry: ## - dry-run goreleaser release
	goreleaser release --snapshot \
		--clean --skip=publish

release: ## - release via goreleaser (needs tag)
	goreleaser release --clean

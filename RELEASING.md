# Releasing

Barebones releasing instructions

- Update Makefile `VERSION_xxx` and submit changes
- Tag the new release (at the above commit)
- Tagging triggers a goreleaser build on github actions
- Artifacts are automatically added to the github release 
- Container images are published to ghcr.io/googlecontainertools/container-structure-test (:latest, :<version>)

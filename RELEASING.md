# Releasing

Barebones releasing instructions

- Update Makefile `VERSION_xxx` and submit changes
- Tag the new release (at the above commit)
- Tagging triggers a build on cloud build
- Go update the release on github with artifacts (from gcs) and release notes

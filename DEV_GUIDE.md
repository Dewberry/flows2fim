# DEV GUIDE

Notes and checklists for repository maintainers. See [CONTRIBUTING.md](CONTRIBUTING.md) for the general contribution process.

## Release Checklist

### Pre-release

- [ ] Run all tests
- [ ] Make the following changes in a separate branch
- [ ] Add a `## [x.y.z] - YYYY-MM-DD` entry to `CHANGELOG.md` for the version being released, the release workflow reads this section for the release notes/title and fails if it's missing
- [ ] Update the footnote links at the bottom of `CHANGELOG.md`: retarget `[Unreleased]`'s compare link to start from the new tag, and add a new `[x.y.z]: https://github.com/NGWPC/flows2fim/releases/tag/vx.y.z` line for the version being released
- [ ] Update sample/test data, if needed
- [ ] Update README.md and help docs for commands, if needed

### Tag and Push

- [ ] Create and push a semver tag (`v*.*.*`) from the release-prep branch, to trigger the `Release` GitHub Actions workflow, which:

  - builds release binaries for all platforms
  - builds, scans with Trivy, and pushes the multi-arch (`amd64`/`arm64`) container image
  - generates release notes and title from the `CHANGELOG.md` entry
  - creates the GitHub release as a **draft**, with all binaries attached
- [ ] If a workflow run fails partway, just fix the issue and re-run it. The release is created as a draft, so retries under the same tag are safe, nothing is locked until it's published

### Publish

- [ ] Review the draft release on GitHub (title, generated notes, attached assets)
- [ ] If everything looks good, merge release-prep branch into `master` with a merge commit (not squash or rebase), this keeps the tagged commit in history that the release will point to
- [ ] Manually click **Publish release** on the draft in the GitHub UI to finalize it

  - This is permanent: GitHub's immutable-releases feature locks the tag and assets once published. The tag can never be reused, even if the release is later deleted
# Changelog

## [3.0.1](https://github.com/woodleighschool/jamf-user-sync/compare/3.0.0...v3.0.1) (2026-10-03)


### Bug Fixes

* **go:** update module resty.dev/v3 (v3.0.0-rc.3 → v3.0.0-rc.4) ([#12](https://github.com/woodleighschool/jamf-user-sync/issues/12)) ([27e8136](https://github.com/woodleighschool/jamf-user-sync/commit/27e8136003ffd42cb2e572c57adac4a13d022c2f))
* skip missing directory accounts during user sync ([ec30117](https://github.com/woodleighschool/jamf-user-sync/commit/ec30117a3e7b02dae11acfa2a784225af65a4510))

## [3.0.0](https://github.com/woodleighschool/jamf-user-sync/compare/2.0.9...3.0.0) (2026-10-01)


### ⚠ BREAKING CHANGES

* replace Python user sync with a one-shot Go command

### Features

* **container:** update image docker.io/library/python (3.12 → 3.14) ([#1](https://github.com/woodleighschool/jamf-user-sync/issues/1)) ([5303586](https://github.com/woodleighschool/jamf-user-sync/commit/5303586641860b0d95e7aabe4f39e57f615264af))
* replace Python user sync with a one-shot Go command ([4f9e243](https://github.com/woodleighschool/jamf-user-sync/commit/4f9e243fb06b48a758fc7b0ae8701ebf8d11253d))


### Bug Fixes

* seed releases from published versions ([2985236](https://github.com/woodleighschool/jamf-user-sync/commit/29852366a0c6202671101818bef0931b7f80ae26))


### Build System

* adopt shared Go tooling and container releases ([6c2c7bf](https://github.com/woodleighschool/jamf-user-sync/commit/6c2c7bf93bee1dfc114b675d11b55cd4d1550208))


### Continuous Integration

* **github-action:** Update action actions/checkout (v5.0.0 → v5.0.1) ([94294b8](https://github.com/woodleighschool/jamf-user-sync/commit/94294b88232c6dd6b45aac69f4641ed1bb1d8255))
* **github-action:** Update action actions/checkout (v5.0.1 → v7.0.0) ([#2](https://github.com/woodleighschool/jamf-user-sync/issues/2)) ([614971d](https://github.com/woodleighschool/jamf-user-sync/commit/614971d6d73e60188147742d749469e94bc225d9))
* **github-action:** Update action actions/checkout (v7.0.0 → v7.0.1) ([929392b](https://github.com/woodleighschool/jamf-user-sync/commit/929392b56199c493090635a36ed67b37fc959eb7))
* **github-action:** Update action docker/build-push-action (v6.18.0 → v6.19.2) ([e9962dd](https://github.com/woodleighschool/jamf-user-sync/commit/e9962dd63ad2514597d9f842f356b7b4a5650915))
* **github-action:** Update action docker/build-push-action (v6.19.2 → v7.3.0) ([#3](https://github.com/woodleighschool/jamf-user-sync/issues/3)) ([400563b](https://github.com/woodleighschool/jamf-user-sync/commit/400563bbde331e63c0d50445c52b83b5382c8cba))
* **github-action:** Update action docker/login-action (v3.5.0 → v3.7.0) ([0e2dc55](https://github.com/woodleighschool/jamf-user-sync/commit/0e2dc553b72dc88cfc64a210d29078f78c6d2ff6))
* **github-action:** Update action docker/login-action (v3.7.0 → v4.4.0) ([#4](https://github.com/woodleighschool/jamf-user-sync/issues/4)) ([ab4cf9f](https://github.com/woodleighschool/jamf-user-sync/commit/ab4cf9fd587ec3ab1f7d29feb041fec327d6e8ed))
* **github-action:** Update action docker/login-action (v4.4.0 → v4.5.0) ([2aec902](https://github.com/woodleighschool/jamf-user-sync/commit/2aec90224deee6815ccdc3aacd23a3c380776a7b))
* **github-action:** Update action docker/login-action (v4.5.0 → v4.5.1) ([50ea33c](https://github.com/woodleighschool/jamf-user-sync/commit/50ea33cb852fc4448970be9de5176fc8ce6fa55f))
* **github-action:** Update action docker/login-action (v4.5.1 → v4.5.2) ([34a3c93](https://github.com/woodleighschool/jamf-user-sync/commit/34a3c9373eff650afa117d3a8120cf553ee4edc9))
* **github-action:** Update action docker/login-action (v4.5.2 → v4.6.0) ([4524923](https://github.com/woodleighschool/jamf-user-sync/commit/4524923f66067f9d96df3a0ee8b11d960b7e3ddc))
* **github-action:** Update action docker/metadata-action (v5.10.0 → v6.2.0) ([#5](https://github.com/woodleighschool/jamf-user-sync/issues/5)) ([ef949d4](https://github.com/woodleighschool/jamf-user-sync/commit/ef949d4f12fd674406c55d9c9dd58632350cf51e))
* **github-action:** Update action docker/metadata-action (v5.8.0 → v5.10.0) ([dbb6807](https://github.com/woodleighschool/jamf-user-sync/commit/dbb680788e66ff0ac13cc5e2d24590ad0b95d83c))
* **github-action:** Update action docker/setup-buildx-action (v3.11.1 → v3.12.0) ([cd49054](https://github.com/woodleighschool/jamf-user-sync/commit/cd4905411440a0f425ee693d1b14954011f0ac7b))
* **github-action:** Update action docker/setup-buildx-action (v3.12.0 → v4.2.0) ([#6](https://github.com/woodleighschool/jamf-user-sync/issues/6)) ([90ef29e](https://github.com/woodleighschool/jamf-user-sync/commit/90ef29e9bcca17b770a8659de43927a783c31c8c))
* **github-action:** update action docker/setup-buildx-action (v4.2.0 → v4.3.0) ([#8](https://github.com/woodleighschool/jamf-user-sync/issues/8)) ([c8c4572](https://github.com/woodleighschool/jamf-user-sync/commit/c8c45726441573fa42a6961b55a4addbeb79b08a))
* **renovate:** use shared configuration ([467a1b6](https://github.com/woodleighschool/jamf-user-sync/commit/467a1b60d082f0fff56f7731f42721e1a192f72b))
* use the dispatch app for renovate runs ([3f06e3c](https://github.com/woodleighschool/jamf-user-sync/commit/3f06e3cd70d8acc4ac1c5a5f8a3d2f16c876cd2c))


### Miscellaneous Chores

* **github-action:** update action ubuntu (24.04 → 26.04) ([#9](https://github.com/woodleighschool/jamf-user-sync/issues/9)) ([dcfa382](https://github.com/woodleighschool/jamf-user-sync/commit/dcfa382d5da4027d2c7fd411044d36fb46c65f68))
* **github-action:** update github-actions ([#10](https://github.com/woodleighschool/jamf-user-sync/issues/10)) ([755571c](https://github.com/woodleighschool/jamf-user-sync/commit/755571c1fb76482c9b991e12c7dcc12a7defd09b))

## 2.0.9

Existing release lineage. Earlier changes are recorded in Git history and releases.

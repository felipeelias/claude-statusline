# Changelog

## [0.11.0](https://github.com/felipeelias/claude-statusline/compare/v0.10.0...v0.11.0) (2026-09-20)


### Features

* add agent_name module ([#25](https://github.com/felipeelias/claude-statusline/issues/25)) ([e7118e9](https://github.com/felipeelias/claude-statusline/commit/e7118e952e4b4ccd60191261d0ba592bd8a8c82b))
* add ANSI style parsing system ([df2d8fc](https://github.com/felipeelias/claude-statusline/commit/df2d8fc619be5f48c11771aebcb79f6eaff7af29))
* add bar_style presets for progress bars ([#19](https://github.com/felipeelias/claude-statusline/issues/19)) ([9518c14](https://github.com/felipeelias/claude-statusline/commit/9518c14437d205f748eaa4605ffc78f0c9342fe7))
* add built-in themes with independent theme and palette axes ([#8](https://github.com/felipeelias/claude-statusline/issues/8)) ([c80cc6f](https://github.com/felipeelias/claude-statusline/commit/c80cc6f1aed62b663df9f3c005c3ec72d319c58a))
* add burn rate and API duration to cost module ([#18](https://github.com/felipeelias/claude-statusline/issues/18)) ([313e0df](https://github.com/felipeelias/claude-statusline/commit/313e0df0fea9d872ff601221bc63050dd0e34b9e))
* add CLI subcommands and built-in themes ([3c33f84](https://github.com/felipeelias/claude-statusline/commit/3c33f8467940da8985c911487171057d6d7ce8c1))
* add effort statusline module ([#64](https://github.com/felipeelias/claude-statusline/issues/64)) ([968da28](https://github.com/felipeelias/claude-statusline/commit/968da28ce46d9a63c55e4c9b4dcf74f264cededd))
* add format string renderer with module composition ([29d76c5](https://github.com/felipeelias/claude-statusline/commit/29d76c5c9238c7dfa4b663d93f5e87ec12ff8dd2))
* add git status indicators to git_branch module ([#13](https://github.com/felipeelias/claude-statusline/issues/13)) ([4579b1a](https://github.com/felipeelias/claude-statusline/commit/4579b1ac1d1006415f95e1718457f4177e232433))
* add git_branch, session_timer, lines_changed modules ([30fcce4](https://github.com/felipeelias/claude-statusline/commit/30fcce45c5afc818d4fa503821a651ae1840c536))
* add JSON input parsing from stdin ([62043ae](https://github.com/felipeelias/claude-statusline/commit/62043aef19a624bd07d5a93aca6248a62bd6f791))
* add model name formatting options ([#22](https://github.com/felipeelias/claude-statusline/issues/22)) ([cbae790](https://github.com/felipeelias/claude-statusline/commit/cbae7902cdab42e20bf1ac1b5d1217d57122817b))
* add model, directory, cost, context modules ([a28da37](https://github.com/felipeelias/claude-statusline/commit/a28da37c198dca8e19026d8e2366e61bbed88a41))
* add OSC 8 clickable hyperlinks for git_branch and directory ([#26](https://github.com/felipeelias/claude-statusline/issues/26)) ([9b7f3af](https://github.com/felipeelias/claude-statusline/commit/9b7f3aff19c30668d64c4676d9696ed8109cb8dd))
* add release infrastructure and example config ([64f32b5](https://github.com/felipeelias/claude-statusline/commit/64f32b5a9631329e91092d0cea8e31bf5663138f))
* add TOML config with defaults and palettes ([a4ae54d](https://github.com/felipeelias/claude-statusline/commit/a4ae54dc6271f30c37df5aa69b8b27c342d41a24))
* add usage limits module ([#14](https://github.com/felipeelias/claude-statusline/issues/14)) ([4c8b51d](https://github.com/felipeelias/claude-statusline/commit/4c8b51dcd4c625c7a62bd137430d292f88519384))
* add version module ([#20](https://github.com/felipeelias/claude-statusline/issues/20)) ([d6c7981](https://github.com/felipeelias/claude-statusline/commit/d6c7981faf46132ced8b0647881bfa087e3234a5))
* add vim_mode module ([#24](https://github.com/felipeelias/claude-statusline/issues/24)) ([b2e5807](https://github.com/felipeelias/claude-statusline/commit/b2e5807a4c9b7dd68b5d9c8f8ad5cecd98268962))
* add windows binaries to releases ([ea8f9a8](https://github.com/felipeelias/claude-statusline/commit/ea8f9a870c782b48a0b481f4a8cf339bf1311197))
* automate releases with release-please ([ff49222](https://github.com/felipeelias/claude-statusline/commit/ff492221da76cb41011b24b3771ee9cb7550cbbf))
* change default format to cwd | branch | model | cost | context ([6856228](https://github.com/felipeelias/claude-statusline/commit/68562287f5bdeadcd65176c71d01afa45c577b6c))
* expand input payload to match full Claude Code schema ([#16](https://github.com/felipeelias/claude-statusline/issues/16)) ([660269b](https://github.com/felipeelias/claude-statusline/commit/660269b2c50ac4a237b8ce5bc7df0fa18abfa480))
* make model details and context pressure visible ([#51](https://github.com/felipeelias/claude-statusline/issues/51)) ([a4d74ac](https://github.com/felipeelias/claude-statusline/commit/a4d74ac1c564ea0520616b9ac6de2db8dde9db19))
* project scaffolding ([8ff6c74](https://github.com/felipeelias/claude-statusline/commit/8ff6c74892f118e1159aab0f387493f5d97841d8))
* wire main.go with full rendering pipeline ([203cbe6](https://github.com/felipeelias/claude-statusline/commit/203cbe682029be12f68206acbbc3ac1e33fed344))


### Bug Fixes

* **bar:** show small non-zero progress ([#60](https://github.com/felipeelias/claude-statusline/issues/60)) ([50e0d0b](https://github.com/felipeelias/claude-statusline/commit/50e0d0bfb70e968f0f3e3a2c358bea40ff7f1943))
* improve themes command readability ([#10](https://github.com/felipeelias/claude-statusline/issues/10)) ([aacd817](https://github.com/felipeelias/claude-statusline/commit/aacd817a7a91877084caae0985c20b622ec8c8d1))
* resolve all 58 golangci-lint issues ([da8cdf4](https://github.com/felipeelias/claude-statusline/commit/da8cdf4c9922c593761644d53def80f19f233604))
* support Windows-style backslash paths in directory truncation ([#48](https://github.com/felipeelias/claude-statusline/issues/48)) ([83a75fe](https://github.com/felipeelias/claude-statusline/commit/83a75fedacc5db7c23e9e425b34b6a49c7fe156b))

## [0.10.0](https://github.com/felipeelias/claude-statusline/compare/v0.9.0...v0.10.0) (2026-09-20)


### Features

* add effort statusline module ([#64](https://github.com/felipeelias/claude-statusline/issues/64)) ([968da28](https://github.com/felipeelias/claude-statusline/commit/968da28ce46d9a63c55e4c9b4dcf74f264cededd))
* make model details and context pressure visible ([#51](https://github.com/felipeelias/claude-statusline/issues/51)) ([a4d74ac](https://github.com/felipeelias/claude-statusline/commit/a4d74ac1c564ea0520616b9ac6de2db8dde9db19))


### Bug Fixes

* **bar:** show small non-zero progress ([#60](https://github.com/felipeelias/claude-statusline/issues/60)) ([50e0d0b](https://github.com/felipeelias/claude-statusline/commit/50e0d0bfb70e968f0f3e3a2c358bea40ff7f1943))
* support Windows-style backslash paths in directory truncation ([#48](https://github.com/felipeelias/claude-statusline/issues/48)) ([83a75fe](https://github.com/felipeelias/claude-statusline/commit/83a75fedacc5db7c23e9e425b34b6a49c7fe156b))

## [0.9.0](https://github.com/felipeelias/claude-statusline/compare/v0.8.0...v0.9.0) (2026-03-31)


### Features

* add agent_name module ([#25](https://github.com/felipeelias/claude-statusline/issues/25)) ([e7118e9](https://github.com/felipeelias/claude-statusline/commit/e7118e952e4b4ccd60191261d0ba592bd8a8c82b))
* add model name formatting options ([#22](https://github.com/felipeelias/claude-statusline/issues/22)) ([cbae790](https://github.com/felipeelias/claude-statusline/commit/cbae7902cdab42e20bf1ac1b5d1217d57122817b))
* add OSC 8 clickable hyperlinks for git_branch and directory ([#26](https://github.com/felipeelias/claude-statusline/issues/26)) ([9b7f3af](https://github.com/felipeelias/claude-statusline/commit/9b7f3aff19c30668d64c4676d9696ed8109cb8dd))
* add vim_mode module ([#24](https://github.com/felipeelias/claude-statusline/issues/24)) ([b2e5807](https://github.com/felipeelias/claude-statusline/commit/b2e5807a4c9b7dd68b5d9c8f8ad5cecd98268962))

## [0.8.0](https://github.com/felipeelias/claude-statusline/compare/v0.7.0...v0.8.0) (2026-03-29)


### Features

* add bar_style presets for progress bars ([#19](https://github.com/felipeelias/claude-statusline/issues/19)) ([9518c14](https://github.com/felipeelias/claude-statusline/commit/9518c14437d205f748eaa4605ffc78f0c9342fe7))
* add burn rate and API duration to cost module ([#18](https://github.com/felipeelias/claude-statusline/issues/18)) ([313e0df](https://github.com/felipeelias/claude-statusline/commit/313e0df0fea9d872ff601221bc63050dd0e34b9e))
* add version module ([#20](https://github.com/felipeelias/claude-statusline/issues/20)) ([d6c7981](https://github.com/felipeelias/claude-statusline/commit/d6c7981faf46132ced8b0647881bfa087e3234a5))
* expand input payload to match full Claude Code schema ([#16](https://github.com/felipeelias/claude-statusline/issues/16)) ([660269b](https://github.com/felipeelias/claude-statusline/commit/660269b2c50ac4a237b8ce5bc7df0fa18abfa480))

## [0.7.0](https://github.com/felipeelias/claude-statusline/compare/v0.6.0...v0.7.0) (2026-03-29)


### Features

* automate releases with release-please ([ff49222](https://github.com/felipeelias/claude-statusline/commit/ff492221da76cb41011b24b3771ee9cb7550cbbf))

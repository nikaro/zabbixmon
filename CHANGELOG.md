<a name="unreleased"></a>
## 0.9.6 (2024-05-29)

### Fix

- **deps**: update all

## 0.9.5 (2024-05-20)

### Fix

- **deps**: update module github.com/gen2brain/beeep to v0.0.0-20240516210008-9c006672e7f4 (#120)

## 0.9.4 (2024-05-20)

### Fix

- **deps**: update module github.com/charmbracelet/bubbletea to v0.26.2 (#119)

## 0.9.3 (2024-05-06)

### Fix

- **deps**: update module github.com/charmbracelet/bubbletea to v0.26.1 (#117)

## 0.9.2 (2024-03-11)

### Fix

- **deps**: update module github.com/charmbracelet/lipgloss to v0.10.0

## 0.9.1 (2024-02-05)

### Fix

- **deps**: update module github.com/charmbracelet/bubbles to v0.18.0 (#92)

## 0.9.0 (2024-01-17)

### Feat

- add setting to skip tls verification

### Refactor

- replace unmaintained zabbix lib

## 0.8.7 (2024-01-15)

### Fix

- **deps**: update github.com/gen2brain/beeep digest to c7bb2cd (#84)

## 0.8.6 (2023-12-25)

### Fix

- **deps**: update module github.com/spf13/viper to v1.18.2 (#81)

## 0.8.5 (2023-12-18)

### Fix

- **deps**: update module github.com/charmbracelet/bubbles to v0.17.1 (#78)

## 0.8.4 (2023-12-11)

### Fix

- **deps**: update module github.com/spf13/viper to v1.18.1 (#75)

## 0.8.3 (2023-12-04)

### Fix

- **deps**: update module github.com/samber/lo to v1.39.0 (#73)

## 0.8.2 (2023-11-06)

### Fix

- **deps**: update module github.com/spf13/cobra to v1.8.0 (#63)

## 0.8.1 (2023-11-05)

### Fix

- replace deprecated tea.Program.Start by Run

## 0.8.0 (2023-10-27)

### Fix

- **deps**: update module github.com/charmbracelet/lipgloss to v0.9.1 (#55)
- **deps**: update module github.com/charmbracelet/lipgloss to v0.9.0 (#54)
- **deps**: update module github.com/spf13/viper to v1.17.0 (#52)
- **deps**: update github.com/gen2brain/beeep digest to 1a38885 (#47)

## v0.7.0 (2023-08-28)

### Refactor

- replace zerolog by stdlib slog

## [v0.6.3] - 2023-03-23
### Chore
- release v0.6.3
- **deps:** bump github.com/charmbracelet/lipgloss from 0.6.0 to 0.7.1
- **deps:** bump github.com/charmbracelet/bubbletea


<a name="v0.6.2"></a>
## [v0.6.2] - 2023-02-04
### Chore
- **deps:** bump all
- **deps:** bump github.com/rs/zerolog from 1.28.0 to 1.29.0
- **deps:** bump github.com/spf13/viper from 1.14.0 to 1.15.0
- **deps:** bump github.com/samber/lo from 1.36.0 to 1.37.0
- **deps:** bump github.com/samber/lo from 1.34.0 to 1.36.0
- **deps:** bump github.com/charmbracelet/bubbletea
- **deps:** bump github.com/charmbracelet/bubbletea
- **deps:** bump github.com/samber/lo from 1.33.0 to 1.34.0
- **deps:** bump github.com/spf13/viper from 1.13.0 to 1.14.0
- **deps:** bump github.com/spf13/cobra from 1.6.0 to 1.6.1
- **deps:** bump github.com/spf13/cobra from 1.5.0 to 1.6.0
- **deps:** bump github.com/samber/lo from 1.31.0 to 1.33.0
- **deps:** bump github.com/samber/lo from 1.29.0 to 1.31.0
- **deps:** bump github.com/samber/lo from 1.28.2 to 1.29.0
- **deps:** bump github.com/samber/lo from 1.28.0 to 1.28.2
- **deps:** bump github.com/charmbracelet/lipgloss from 0.5.0 to 0.6.0
- **deps:** update viper and bubbles
- **deps:** bump github.com/samber/lo from 1.27.1 to 1.28.0


<a name="v0.6.1"></a>
## [v0.6.1] - 2022-08-31
### Bug Fixes
- exit on zabbix connect/fetch errors

### Chore
- **deps:** bump github.com/rs/zerolog from 1.27.0 to 1.28.0
- **deps:** bump github.com/samber/lo from 1.27.0 to 1.27.1


<a name="v0.6.0"></a>
## [v0.6.0] - 2022-08-24
### Features
- **tui:** add `s` keybinding to open ssh://host


<a name="v0.5.0"></a>
## [v0.5.0] - 2022-08-21
### Bug Fixes
- keep cursor position after table update

### Features
- add event time to output


<a name="v0.4.0"></a>
## [v0.4.0] - 2022-08-20
### Chore
- **deps:** bump github.com/pterm/pterm from 0.12.44 to 0.12.45
- **deps:** bump github.com/samber/lo from 1.26.0 to 1.27.0
- **deps:** bump github.com/pterm/pterm from 0.12.42 to 0.12.44
- **deps:** bump github.com/samber/lo from 1.25.0 to 1.26.0
- **deps:** bump github.com/samber/lo from 1.21.0 to 1.25.0

### Features
- add navigation between items


<a name="v0.3.4"></a>
## [v0.3.4] - 2022-06-24
### Chore
- better error messages

### Code Refactoring
- use a more simple tui framework


<a name="v0.3.3"></a>
## [v0.3.3] - 2022-06-19
### Code Refactoring
- move data loop into a goroutine
- remove unnecessary type arguments
- update table instead of rebuild it
- use termui for ui rendering


<a name="v0.3.2"></a>
## [v0.3.2] - 2022-06-15
### Code Refactoring
- use zabbix pacakge constants
- merge cmd and config files


<a name="v0.3.1"></a>
## [v0.3.1] - 2022-06-14
### Bug Fixes
- use buffered chan

### Chore
- simplify flag/config bind
- check error on clean exec

### Code Refactoring
- revert split in modules


<a name="v0.3.0"></a>
## [v0.3.0] - 2022-06-13
### Bug Fixes
- fail if not config file found

### Chore
- update deps
- log instead of panic
- do not re-format vendored files
- **deps:** bump github.com/rs/zerolog from 1.26.1 to 1.27.0

### Features
- add 'grep' flag to filter hosts with regexp
- check mandatory settings are set
- override json keys
- marshal items instead of table for better json ouput


<a name="v0.2.0"></a>
## [v0.2.0] - 2022-06-07
### Features
- dump json when output is redirected


<a name="v0.1.0"></a>
## v0.1.0 - 2022-06-06
### Chore
- move config into its own package
- split into packages
- fix output filename
- update go-zabbix dependency


[Unreleased]: https://github.com/nikaro/zabbixmon/compare/v0.6.3...HEAD
[v0.6.3]: https://github.com/nikaro/zabbixmon/compare/v0.6.2...v0.6.3
[v0.6.2]: https://github.com/nikaro/zabbixmon/compare/v0.6.1...v0.6.2
[v0.6.1]: https://github.com/nikaro/zabbixmon/compare/v0.6.0...v0.6.1
[v0.6.0]: https://github.com/nikaro/zabbixmon/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/nikaro/zabbixmon/compare/v0.4.0...v0.5.0
[v0.4.0]: https://github.com/nikaro/zabbixmon/compare/v0.3.4...v0.4.0
[v0.3.4]: https://github.com/nikaro/zabbixmon/compare/v0.3.3...v0.3.4
[v0.3.3]: https://github.com/nikaro/zabbixmon/compare/v0.3.2...v0.3.3
[v0.3.2]: https://github.com/nikaro/zabbixmon/compare/v0.3.1...v0.3.2
[v0.3.1]: https://github.com/nikaro/zabbixmon/compare/v0.3.0...v0.3.1
[v0.3.0]: https://github.com/nikaro/zabbixmon/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/nikaro/zabbixmon/compare/v0.1.0...v0.2.0

<a name="unreleased"></a>
## [Unreleased]


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

### Code Refactoring
- revert split in modules


<a name="v0.3.0"></a>
## [v0.3.0] - 2022-06-13
### Bug Fixes
- fail if not config file found

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

[Unreleased]: https://github.com/nikaro/zabbixmon/compare/v0.3.3...HEAD
[v0.3.3]: https://github.com/nikaro/zabbixmon/compare/v0.3.2...v0.3.3
[v0.3.2]: https://github.com/nikaro/zabbixmon/compare/v0.3.1...v0.3.2
[v0.3.1]: https://github.com/nikaro/zabbixmon/compare/v0.3.0...v0.3.1
[v0.3.0]: https://github.com/nikaro/zabbixmon/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/nikaro/zabbixmon/compare/v0.1.0...v0.2.0

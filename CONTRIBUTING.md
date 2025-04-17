# Contributing Guidelines

## Changelog Format

### Entry Format

Each changelog entry should follow this format:

* [\<module\>] \<description\> (\<PR\>)

Where:

* \<module\>: The module name (e.g., x/bank, app, cli, tests)
* \<description\>: A clear description of the change
* \<PR\>: Link to the pull request

### Components

| Component | Description | Example |
|-----------|-------------|---------|
| `<module>` | The module name | `x/pevm`, `app`, `api` |
| `<description>` | A clear description of the change | "Add new transfer validation" |
| `<PR>` | Link to the pull request | `#1234` |

### Module Names

| Category | Module | Description |
|----------|--------|-------------|
| **Core** | `[app]` | Application core |
| | `[cli]` | Command line interface |
| | `[api]` | API related |
| | `[store]` | Storage layer |
| **Development** | `[e2e]` | End-to-end testing |
| | `[docs]` | Documentation |
| | `[build]` | Build system |
| | `[ci]` | CI/CD related |
| **Modules** | `[x/authority]` | Authority module |
| | `[x/emissions]` | Emission module |
| | `[x/lightclient]` | Light client module |
| | `[x/pevm]` | pell evm module |
| | `[x/relayer]` | pell crosschain relayer module |
| | `[x/restaking]` | pell restaking manager module |
| | `[x/xmsg]` | pell crosschain messaging module |

### Description Format

Use these standard commit types:

| Type | Description |
|------|-------------|
| `bug` | Something isn't working |
| `build` | Changes that affect the build system or external dependencies |
| `chore` | Other changes that don't modify src or test files |
| `ci` | Changes to our CI configuration files and scripts |
| `documentation` | Improvements or additions to documentation |
| `enhancement` | New feature or request |
| `perf` | A code change that improves performance |
| `refactor` | A code change that neither fixes a bug nor adds a feature |
| `test` | Adding missing tests or correcting existing tests |

### Example Entries

* [x/pevm] bug: Fix panic during transaction validation ([#1234](https://github.com/0xPellNetwork/chain/pull/1234))
* [cli] enhancement: Add new query command output format ([#1235](https://github.com/0xPellNetwork/chain/pull/1235))
* [app] perf: Optimize transaction processing ([#1236](https://github.com/0xPellNetwork/chain/pull/1236))

### Categories Order

1. 🚨 State Machine Breaking
2. 🔄 API Breaking
3. ✨ Features
4. 📈 Improvements
5. 🐛 Bug Fixes
6. 📦 Dependencies
7. 📝 Documentation
8. 🧪 Testing

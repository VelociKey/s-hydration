# Antigravity Skill: workspace_preserve
This skill stages, commits, and pushes conformed changes in a workspace (or an entire silo of workspaces) up to GitHub using `fab-construct.exe`.

## Capability Intents
- **"preserve workspace <name>"**: Safely commits and pushes workspace changes in the default 00flow silo.
- **"preserve workspace <name> in silo <silo>"**: Commits and pushes workspace changes in a specific silo directory.
- **"preserve silo <silo>"**: Commits and pushes all active workspaces inside the specified silo directory.

## Execution Steps
Run the following Go-native construct command:
```powershell
C:\aCogSpaceSeed\00flow\s-forge\97000-internal-toolchains\fab-construct.exe preserve -silo <silo> -name <name>
```
Where:
- `<silo>` defaults to `00flow` if unspecified.
- `<name>` represents the target workspace. If `<name>` is omitted (or empty), the utility automatically scans and preserves all workspaces in the target silo.


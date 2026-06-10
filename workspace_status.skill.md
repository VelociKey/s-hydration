# Antigravity Skill: workspace_status
This skill runs `fab-construct.exe` to inspect the git status and pending updates of all workspaces inside a target silo.

## Capability Intents
- **"check workspace status"**: Scans default 00flow silo workspaces.
- **"check workspace status for <silo>"**: Scans workspaces in the specified silo directory.

## Execution Steps
Run the following Go-native construct command:
```powershell
C:\aCogSpaceSeed\00flow\s-forge\97000-internal-toolchains\fab-construct.exe status -silo <silo>
```
Where `<silo>` defaults to `00flow` if unspecified.

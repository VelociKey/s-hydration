# Antigravity Skill: workspace_preserve
This skill stages, commits, and pushes conformed changes in a workspace up to GitHub using `fab-construct.exe`.

## Capability Intents
- **"preserve workspace <name>"**: Safely commits and pushes workspace changes in the default 00flow silo.
- **"preserve workspace <name> in silo <silo>"**: Commits and pushes workspace changes in a specific silo directory.

## Execution Steps
Run the following Go-native construct command:
```powershell
C:\aCogSpaceSeed\00flow\s-forge\97000-internal-toolchains\fab-construct.exe preserve -silo <silo> -name <name>
```
Where `<silo>` defaults to `00flow` if unspecified.

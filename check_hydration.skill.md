# Antigravity Skill: check_hydration
This skill triggers dry-run checks across our external SBOM registry and internal workspace harnesses.

## Capability Intents
- **"check hydration for external artifacts"**: Executes a dry-run check of the external SBOM dependencies.
- **"check hydration for internal artifacts"**: Executes verification across internal workspace harnesses.
- **"check hydration for all artifacts"**: Executes verification across all external and internal configurations.

## Verification Steps
When called, execute the following command:
```powershell
C:\aCogSpaceSeed\00flow\s-hydration\hydrate-check.exe -scope <scope>
```
Where `<scope>` is mapped from the input phrase:
- "external" for "check hydration for external artifacts"
- "internal" for "check hydration for internal artifacts"
- "all" for "check hydration for all artifacts"

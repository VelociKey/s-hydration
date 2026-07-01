# Active Task Goal: Sovereign Workspace Construction & Semantic Stage/Publish Engine

**Task ID:** `00flow-TASK-FAB-CONSTRUCT-ENHANCEMENT`  
**Sponsor:** Conductor / NATVS Engine  
**Status:** PROPOSED (Negotiation & Assimilation Phase)  
**Target Tracks:** **[fab_construct_enhancement_track.md](file:///c:/aCogSpaceSeed/conductor/tracks/fab_construct_enhancement_track.md)**

---

## 🎯 Global Goal Mandate

The **NATVS Engine** or **Gemini-CLI** must enhance the Go-based **`fab-construct`** utility inside the `s-fab-aides` workspace. The tool must support modular workspace/GitHub twin construction and implement high-level semantic Git operations (`add` and `publish`) backed by native security, vulnerability, and license compliance auditing (using **Trivy**).

---

## 📋 Directives & Invariant Constraints

1.  **Twin Separability Invariant:** Users/Agents must be able to invoke `fab-construct workspace` to build local layouts only, `fab-construct github` to provision GitHub twins only, and `fab-construct both` for unified initialization.
2.  **License Compliance Invariant:** The semantic `add` command must automatically scan all project dependencies using **Trivy** to assert that no viral/copyleft licenses (e.g., GPLv3) are introduced. Staging must halt immediately if non-compliant licenses are detected.
3.  **Security & Maturity Gates:** The semantic `publish` (push) command must execute vulnerability and secret scanning (via Trivy) and assess software maturity metrics (documentation presence, test coverage) before pushing. It must physically block any remote push to the `zenith` branch if any safety assertion fails.

---

## 🚀 Active Roadmap to Execute

*   [ ] **Phase 1: Refactor CLI Subcommands**
    *   Enhance subcommands to handle isolation between local workspace directory hydration and GitHub API remote instantiation.
*   [ ] **Phase 2: Implement Semantic Staging (`add` subcommand)**
    *   Integrate Trivy license scanner execution via Go's `os/exec`.
    *   Add automated Software Bill of Materials (SBOM) generation and mapping.
    *   Verify commit message schema compliance before calling local Git commit.
*   [ ] **Phase 3: Implement Semantic Publishing (`publish` subcommand)**
    *   Integrate Trivy vulnerability and secret scanner.
    *   Develop a Software Maturity Assessment routine (checking for test presence and `AGENTS.md` compliance).
    *   Execute `git push origin zenith` only upon $100\%$ gate clearance.

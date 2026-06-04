//go:build !production

package hydration

import (
	"os"
	"time"
)

func (h *SovereignHydrator) GetExperience() map[string]ExperienceRecord {
	return h.experience
}

func (h *SovereignHydrator) SetExperience(name string, rec ExperienceRecord) {
	h.experience[name] = rec
}

func (h *SovereignHydrator) SetABTest(val bool) {
	h.abtest = val
}

func (h *SovereignHydrator) SetSequential(val bool) {
	h.sequential = val
}

func (h *SovereignHydrator) DetermineExecutionPlan(records []ArtifactRecord) ([]ArtifactRecord, []ArtifactRecord, map[string][]ArtifactRecord) {
	return h.determineExecutionPlan(records)
}

func (h *SovereignHydrator) SaveExperience(path string) error {
	return h.saveExperience(path)
}

func (h *SovereignHydrator) LoadExperience(path string) error {
	return h.loadExperience(path)
}

func (h *SovereignHydrator) ShouldPrune(path string, info os.FileInfo) (bool, bool) {
	return h.shouldPrune(path, info)
}

func (h *SovereignHydrator) PrintABTestReport(records []ArtifactRecord, seqDurations map[string]time.Duration, totalSeq time.Duration, parDurations map[string]time.Duration, totalPar time.Duration) {
	h.printABTestReport(records, seqDurations, totalSeq, parDurations, totalPar)
}

func ParseGoWork(path string) ([]string, error) {
	return parseGoWork(path)
}

func FindDownstreamNodes(graph map[string][]string, node string) map[string]bool {
	return findDownstreamNodes(graph, node)
}

func TopologicalSort(graph map[string][]string, subset map[string]bool) ([]string, error) {
	return topologicalSort(graph, subset)
}

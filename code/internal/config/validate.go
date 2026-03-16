package config

import (
	"fmt"
	"strings"

	"juyu-ai-platform/internal/types"
)

func Validate(cfg types.Config) error {
	masterByCode := map[string]types.MasterAgent{}
	masterByScene := map[string]types.MasterAgent{}
	abilityByCode := map[string]types.AbilityConfig{}

	if len(cfg.MasterAgents) == 0 {
		return fmt.Errorf("config: master_agents is empty")
	}
	if len(cfg.AbilityAgents) == 0 {
		return fmt.Errorf("config: ability_agents is empty")
	}

	for _, m := range cfg.MasterAgents {
		if strings.TrimSpace(m.Code) == "" {
			return fmt.Errorf("config: master agent code is required")
		}
		if strings.TrimSpace(m.SceneType) == "" {
			return fmt.Errorf("config: master agent %s missing scene_type", m.Code)
		}
		if _, ok := masterByCode[m.Code]; ok {
			return fmt.Errorf("config: duplicate master agent code %s", m.Code)
		}
		if existing, ok := masterByScene[m.SceneType]; ok {
			return fmt.Errorf("config: duplicate scene_type %s on master agents %s and %s", m.SceneType, existing.Code, m.Code)
		}
		masterByCode[m.Code] = m
		if m.Enabled {
			masterByScene[m.SceneType] = m
		}
	}

	for _, a := range cfg.AbilityAgents {
		if strings.TrimSpace(a.Code) == "" {
			return fmt.Errorf("config: ability agent code is required")
		}
		if _, ok := abilityByCode[a.Code]; ok {
			return fmt.Errorf("config: duplicate ability code %s", a.Code)
		}
		abilityByCode[a.Code] = a
	}

	if len(cfg.Bindings) == 0 {
		return fmt.Errorf("config: bindings is empty")
	}
	for _, b := range cfg.Bindings {
		masterCode := strings.TrimSpace(b.MasterAgent)
		if masterCode == "" {
			return fmt.Errorf("config: binding missing master_agent")
		}
		if _, ok := masterByCode[masterCode]; !ok {
			return fmt.Errorf("config: binding references unknown master agent %s", masterCode)
		}
		if len(b.Workflow) == 0 && len(b.Abilities) == 0 {
			return fmt.Errorf("config: binding for %s has no workflow or abilities", masterCode)
		}
		seenSteps := map[int]struct{}{}
		for _, step := range b.Workflow {
			if step.Step <= 0 {
				return fmt.Errorf("config: binding %s has invalid workflow step %d", masterCode, step.Step)
			}
			if _, exists := seenSteps[step.Step]; exists {
				return fmt.Errorf("config: binding %s has duplicate workflow step %d", masterCode, step.Step)
			}
			seenSteps[step.Step] = struct{}{}
			if strings.TrimSpace(step.Ability) == "" {
				return fmt.Errorf("config: binding %s step %d missing ability", masterCode, step.Step)
			}
			if _, ok := abilityByCode[step.Ability]; !ok {
				return fmt.Errorf("config: binding %s step %d references unknown ability %s", masterCode, step.Step, step.Ability)
			}
		}
		for _, ability := range b.Abilities {
			if _, ok := abilityByCode[ability]; !ok {
				return fmt.Errorf("config: binding %s references unknown ability %s", masterCode, ability)
			}
		}
	}
	return nil
}

func ValidateRuntime(storeDriver, sqlitePath, pgDSN, apiKey string, dryRun bool) error {
	driver := strings.TrimSpace(strings.ToLower(storeDriver))
	switch driver {
	case "", "memory":
	case "sqlite":
		if strings.TrimSpace(sqlitePath) == "" {
			return fmt.Errorf("config: JUYU_SQLITE_PATH is required when JUYU_STORE=sqlite")
		}
	case "postgres", "pg":
		if strings.TrimSpace(pgDSN) == "" {
			return fmt.Errorf("config: JUYU_PG_DSN is required when JUYU_STORE=postgres")
		}
	default:
		return fmt.Errorf("config: unsupported JUYU_STORE %q", storeDriver)
	}

	if !dryRun && strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("config: JUYU_API_KEY is required when JUYU_ADAPTER_DRY_RUN=false")
	}
	return nil
}

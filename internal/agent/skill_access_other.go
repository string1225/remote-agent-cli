//go:build !windows

package agent

func PrepareUserSkillServiceAccess(string) error { return nil }

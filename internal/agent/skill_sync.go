package agent

import (
	"log"

	"github.com/string1225/remote-agent-cli/internal/buildinfo"
	"github.com/string1225/remote-agent-cli/internal/codexsetup"
)

func syncActivatedWorkerSkill(profile ServiceProfile) {
	if _, err := codexsetup.SyncUserRemoteAgentSkill(profile.Environment["USERPROFILE"], profile.Environment["CODEX_HOME"], buildinfo.Current()); err != nil {
		log.Printf("remote-agent skill synchronization failed; client continues: %v", err)
	}
}

package bot

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
)

func onJoin(s *discordgo.Session, g *discordgo.GuildCreate) {
	fmt.Println("Joined guild:", g.Name)
	ensureRole(s, g)
}

func ensureRole(s *discordgo.Session, g *discordgo.GuildCreate) {
	for _, role := range g.Roles {
		if role.Name == teacher {
			return
		}
	}

	fmt.Println("Creating teacher role for guild:", g.Name)
	role, err := s.GuildRoleCreate(g.ID)
	if err != nil {
		fmt.Println(err)
		return
	}

	s.GuildRoleEdit(g.ID, role.ID, teacher, 0, false, 0, false)
}
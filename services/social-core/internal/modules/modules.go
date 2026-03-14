package modules

// Module names track the bounded contexts that move into the consolidated
// social-core runtime during the product rebuild.
var ModuleNames = []string{
	"identity",
	"social",
	"invite",
	"private-chat",
	"guild-basics",
	"party-basics",
}

package modules

// Module names document the only ingress concerns that the rebuilt runtime
// should own in the first product phase.
var ModuleNames = []string{
	"authn",
	"rate-limit",
	"session-context",
	"client-http",
	"client-websocket",
}
